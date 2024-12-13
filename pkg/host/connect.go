package host

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsoption "github.com/shaowenchen/ops/pkg/option"
	opsstorage "github.com/shaowenchen/ops/pkg/storage"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HostConnectionCache struct {
	cache map[string]*HostConnection
	Mutex *sync.RWMutex
}

func (c *HostConnectionCache) Get(key string) *HostConnection {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.cache[key]
}

func (c *HostConnectionCache) Set(key string, value *HostConnection) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.cache[key] = value
}

type HostConnection struct {
	Host      *opsv1.Host
	scpclient *scp.Client
	sshclient *ssh.Client
}

var hcCache = HostConnectionCache{cache: make(map[string]*HostConnection), Mutex: &sync.RWMutex{}}

func NewHostConnBase64(h *opsv1.Host) (hc *HostConnection, err error) {
	if h == nil {
		h = &opsv1.Host{}
	}
	hc = &HostConnection{}
	hc.Host = h
	// empty address is local host
	if h.Spec.Address == "" {
		h.Spec.Address = opsconstants.LocalHostIP
	}
	// local host
	if h.Spec.Address == opsconstants.LocalHostIP {
		return hc, nil
	}
	key := fmt.Sprintf("%s:%d", h.Spec.Address, h.Spec.Port)
	if hc := hcCache.Get(key); hc != nil {
		return hc, nil
	}
	err = hc.connecting()
	if err != nil {
		return nil, err
	}
	hcCache.Set(key, hc)
	return
}

func (c *HostConnection) Shell(ctx context.Context, sudo bool, content string) (stdout string, err error) {
	reg := regexp.MustCompile(`\${[^\}]*}`)
	funcStrList := reg.FindAllString(content, -1)
	for _, callFunc := range funcStrList {
		rawCallFunc := callFunc
		callFunc = callFunc[2 : len(callFunc)-1]
		stdout, err = c.shellFuncMap(ctx, sudo, callFunc)
		if err != nil {
			return stdout, err
		}
		content = strings.ReplaceAll(content, rawCallFunc, stdout)
	}
	return c.execScript(ctx, sudo, content)
}

func (c *HostConnection) shellFuncMap(ctx context.Context, sudo bool, funcFull string) (stdout string, err error) {
	if funcFull == "installOpscli()" {
		return c.install(ctx, sudo, "opscli")
	} else if funcFull == "distribution()" {
		return c.getDistribution(ctx, sudo)
	}
	return
}

func (c *HostConnection) install(ctx context.Context, sudo bool, component string) (stdout string, err error) {
	if component == "opscli" {
		proxy := ""
		if !c.isInChina(ctx) {
			proxy = opsconstants.DefaultProxy
		}
		return c.execScript(ctx, sudo, opsutils.ShellInstallOpscli(proxy))
	}
	return
}

func (c *HostConnection) isInChina(ctx context.Context) (ok bool) {
	_, err := c.execScript(ctx, false, opsutils.ShellIsInChina())
	if err != nil {
		return true
	}
	return false
}

func (c *HostConnection) File(ctx context.Context, fileOpt opsoption.FileOption) (out string, err error) {
	switch fileOpt.GetStorageType() {
	case opsconstants.RemoteStorageTypeS3:
		return c.fileS3(ctx, fileOpt)
	case opsconstants.RemoteStorageTypeServer:
		return c.filseServer(ctx, fileOpt)
	default:
		err = errors.New("invalid storage type")
	}
	return
}

func (c *HostConnection) fileS3(ctx context.Context, fileOpt opsoption.FileOption) (output string, err error) {
	if c.Host.Spec.Address == opsconstants.LocalHostIP {
		// use func to
		return opsstorage.S3File(fileOpt)
	}
	// use opscli to transfer file
	cmd := ""
	if fileOpt.IsUploadDirection() {
		cmd = opsutils.ShellOpscliDownS3(fileOpt.Region, fileOpt.Endpoint, fileOpt.Bucket, fileOpt.AK, fileOpt.SK, fileOpt.LocalFile, fileOpt.RemoteFile)
	} else if fileOpt.IsDownloadDirection() {
		cmd = opsutils.ShellOpscliUploadS3(fileOpt.Region, fileOpt.Endpoint, fileOpt.Bucket, fileOpt.AK, fileOpt.SK, fileOpt.LocalFile, fileOpt.RemoteFile)
	}
	if cmd != "" {
		_, err = c.execScript(ctx, fileOpt.Sudo, cmd)
		return
	} else {
		errors.New("invalid direction")
	}
	return
}

func (c *HostConnection) filseServer(ctx context.Context, fileOpt opsoption.FileOption) (output string, err error) {
	if c.Host.Spec.Address == opsconstants.LocalHostIP {
		// use func to
		return opsstorage.ServerFile(fileOpt)
	}
	// use opscli to transfer file
	cmd := ""
	if fileOpt.IsUploadDirection() {
		cmd = opsutils.ShellOpscliDownServer(fileOpt.Api, fileOpt.AesKey, fileOpt.LocalFile, fileOpt.RemoteFile)
	} else if fileOpt.IsDownloadDirection() {
		cmd = opsutils.ShellOpscliUploadServer(fileOpt.Api, fileOpt.AesKey, fileOpt.LocalFile, fileOpt.RemoteFile)
	}
	if cmd != "" {
		_, err = c.execScript(ctx, fileOpt.Sudo, cmd)
		return
	} else {
		errors.New("invalid direction")
	}
	return
}

func (c *HostConnection) GetStatus(ctx context.Context, sudo bool) (status *opsv1.HostStatus, err error) {
	anyOneIsOk := false
	hostname, err1 := c.getHosname(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	kerneVersion, err1 := c.getKernelVersion(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	distribution, err1 := c.getDistribution(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	arch, err1 := c.getArch(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	diskTotal, err1 := c.getDiskTotal(ctx, sudo, opsconstants.DefaultShellTimeoutSeconds)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	diskUsagePercent, err1 := c.getDiskUsagePercent(ctx, sudo, opsconstants.DefaultShellTimeoutSeconds)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	cpuTotal, err1 := c.getCPUTotal(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	cpuLoad1, err1 := c.getCPULoad1(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	cpuUsagePercent, err1 := c.getCPUUsagePercent(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	memTotal, err1 := c.getMemTotal(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	memUsagePercent, err1 := c.getMemUsagePercent(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	accelVendor, err1 := c.getAcceleratorVendor(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}
	accelModel, err1 := c.getAcceleratorModel(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	accelCount, err1 := c.getAcceleratorCount(ctx, sudo)
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	status = &opsv1.HostStatus{
		Hostname:          hostname,
		KernelVersion:     kerneVersion,
		Distribution:      distribution,
		Arch:              arch,
		DiskTotal:         diskTotal,
		DiskUsagePercent:  diskUsagePercent,
		CPUTotal:          cpuTotal,
		CPULoad1:          cpuLoad1,
		CPUUsagePercent:   cpuUsagePercent,
		MemTotal:          memTotal,
		MemUsagePercent:   memUsagePercent,
		AcceleratorVendor: accelVendor,
		AcceleratorModel:  accelModel,
		AcceleratorCount:  accelCount,
		HeartTime:         &metav1.Time{Time: time.Now()},
		HeartStatus:       opsconstants.StatusSuccessed,
	}
	if !anyOneIsOk {
		status.HeartStatus = opsconstants.StatusFailed
	}
	return
}

func (c *HostConnection) session() (*ssh.Session, error) {
	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}
	sess, err := c.sshclient.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = sess.RequestPty("xterm", 100, 50, modes)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (c *HostConnection) connecting() (err error) {
	password, err := opsutils.DecodingBase64ToString(c.Host.Spec.Password)
	if err != nil {
		return err
	}
	privateKey, err := opsutils.DecodingBase64ToString(c.Host.Spec.PrivateKey)
	if err != nil {
		return err
	}
	authMethods := make([]ssh.AuthMethod, 0)
	if len(password) > 0 {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(privateKey) > 0 {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return errors.New("The given SSH key could not be parsed")
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	sshConfig := &ssh.ClientConfig{
		User:            c.Host.Spec.Username,
		Timeout:         time.Duration(c.Host.Spec.TimeOutSeconds) * time.Second,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Config:          ssh.Config{},
	}

	endpointBehindBastion := net.JoinHostPort(c.Host.Spec.Address, strconv.Itoa(c.Host.Spec.Port))

	c.sshclient, err = ssh.Dial("tcp", endpointBehindBastion, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "client.Dial failed %s", c.Host.Spec.Address)
	}
	client, err := scp.NewClientBySSH(c.sshclient)
	c.scpclient = &client
	if err != nil {
		return errors.Wrapf(err, "scp.NewClient failed")
	}
	return nil
}

func (c *HostConnection) close() {
	if c.sshclient != nil {
		c.sshclient.Close()
	}
	if c.scpclient != nil {
		c.scpclient.Close()
	}
}

func (c *HostConnection) execSh(ctx context.Context, sudo bool, cmd string) (stdout string, err error) {
	return c.ExecWithExecutor(ctx, sudo, "sh", "-c", cmd)
}

func (c *HostConnection) execPython(ctx context.Context, sudo bool, cmd string) (stdout string, err error) {
	return c.ExecWithExecutor(ctx, sudo, "python3", "-c", cmd)
}

func (c *HostConnection) execScript(ctx context.Context, sudo bool, cmd string) (stdout string, err error) {
	lines := strings.Split(cmd, "\n")
	if len(lines) > 1 && strings.Contains(lines[0], "python") {
		return c.execPython(ctx, sudo, cmd)
	}
	return c.execSh(ctx, sudo, cmd)
}

func (c *HostConnection) ExecWithExecutor(ctx context.Context, sudo bool, executor, param, rawCmd string) (stdout string, err error) {
	cmd := opsutils.BuildBase64CmdWithExecutor(sudo, rawCmd, executor)
	// run in localhost
	if c.Host.Spec.Address == opsconstants.LocalHostIP {
		runner := exec.Command("bash", "-c", cmd)
		if sudo {
			runner = exec.Command("sudo", "bash", "-c", cmd)
		}
		var out, errout bytes.Buffer
		runner.Stdout = &out
		runner.Stderr = &errout
		err = runner.Run()
		if err != nil {
			stdout = errout.String()
			return
		}
		stdout = out.String()
		return
	}
	sess, err := c.session()
	if err != nil {
		return "", errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	in, _ := sess.StdinPipe()
	out, _ := sess.StdoutPipe()
	err = sess.Start(cmd)
	if err != nil {
		return "", err
	}
	//TODO: add timeout and improve the code
	if strings.Contains(rawCmd, "reboot") || strings.Contains(rawCmd, "halt") || strings.Contains(rawCmd, "shutdown") || strings.Contains(rawCmd, "ipmitool") {
		return "", nil
	}
	var (
		output []byte
		line   = ""
		r      = bufio.NewReader(out)
	)
	printLogStream := false
	hasPrintCache := false
	time.AfterFunc(time.Second*3, func() {
		printLogStream = true
	})
	for {
		select {
		case <-ctx.Done():
			goto END
		default:
			b, err := r.ReadByte()
			if err != nil {
				goto END
			}
			output = append(output, b)
			if printLogStream && !hasPrintCache {
				fmt.Print(string(output))
				hasPrintCache = !hasPrintCache
			}
			if b == byte('\n') {
				if printLogStream {
					fmt.Print(line)
				}
				line = ""
				continue
			}

			line += string(b)

			if (strings.HasPrefix(line, "[sudo] password for ") || strings.HasPrefix(line, "Password")) && strings.HasSuffix(line, ": ") {
				_, err = in.Write([]byte(c.Host.Spec.Password + "\n"))
				if err != nil {
					goto END
				}
			}
		}
	}
END:
	err = sess.Wait()
	return strings.TrimRight(string(output), "\r\n"), err
}

func (c *HostConnection) mv(ctx context.Context, sudo bool, src, dst string) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellMv(src, dst))
}

func (c *HostConnection) copy(ctx context.Context, sudo bool, src, dst string) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellCopy(src, dst))
}

func (c *HostConnection) chown(ctx context.Context, sudo bool, idU, idG, src string) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellChown(idU, idG, src))
}

func (c *HostConnection) rm(ctx context.Context, sudo bool, dst string) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellRm(dst))
}

func (c *HostConnection) cmdPull(ctx context.Context, sudo bool, src, dst string) (err error) {
	srcmd5, err := c.fileMd5(ctx, sudo, src)
	if err != nil {
		return err
	}
	output, err := c.execScript(ctx, sudo, fmt.Sprintf("cat %s | base64 -w 0", src))
	if err != nil {
		return fmt.Errorf("open src file failed %v, src path: %s", err, src)
	}
	dstDir := filepath.Dir(dst)
	if opsutils.IsExistsFile(dstDir) {
		err = os.MkdirAll(dstDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dst dir failed %v, dst dir: %s", err, dstDir)
		}
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst file failed %v", err)
	}
	defer dstFile.Close()

	if base64Str, err := base64.StdEncoding.DecodeString(output); err != nil {
		return err
	} else {
		if _, err = dstFile.WriteString(string(base64Str)); err != nil {
			return err
		}
	}

	dstmd5, err := opsutils.FileMD5(dst)
	if err != nil {
		return
	}

	if dstmd5 != srcmd5 {
		return errors.New(fmt.Sprintf("md5 error: dstfile is %s, srcfile is %s", dstmd5, srcmd5))
	}

	return nil
}

func (c *HostConnection) scpPull(ctx context.Context, sudo bool, src, dst string) (err error) {
	originSrc := src
	src = c.getTempfileName(ctx, src)
	stdout, err := c.copy(ctx, sudo, originSrc, src)
	if err != nil {
		return errors.New(stdout)
	}
	idU, err := c.getIDU(ctx)
	if err != nil {
		return errors.New(stdout)
	}
	idG, err := c.getIDG(ctx)
	if err != nil {
		return errors.New(stdout)
	}
	stdout, err = c.chown(ctx, sudo, idU, idG, src)
	if err != nil {
		return errors.New(stdout)
	}
	srcmd5, err := c.fileMd5(ctx, sudo, originSrc)
	if err != nil {
		return err
	}
	dst = opsutils.GetAbsoluteFilePath(dst)
	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	err = c.scpclient.CopyFromRemote(ctx, dstFile, src)

	if err != nil {
		return
	}

	stdout, err = c.rm(ctx, sudo, src)
	if err != nil {
		return errors.New(stdout)
	}

	dstmd5, err := opsutils.FileMD5(dst)
	if err != nil {
		return
	}
	if dstmd5 != srcmd5 {
		err = errors.New(fmt.Sprintf("md5 error: dstfile is %s, srcfile is %s", dstmd5, srcmd5))
		return
	}

	return
}

func (c *HostConnection) scpPush(ctx context.Context, sudo bool, src, dst string) (err error) {
	originDst := dst
	dst = c.getTempfileName(ctx, dst)
	if c.Host.Spec.Address == opsconstants.LocalHostIP {
		return errors.New("remote address is localhost")
	}
	srcmd5, err := opsutils.FileMD5(src)
	if err != nil {
		return err
	}
	err = c.makeDir(ctx, sudo, originDst)
	if err != nil {
		return err
	}
	src = opsutils.GetAbsoluteFilePath(src)
	srcFile, err := os.Open(src)
	err = c.scpclient.CopyFromFile(context.Background(), *srcFile, dst, "0655")

	if err != nil {
		return err
	}
	stdout, err := c.mv(ctx, sudo, dst, originDst)
	if err == nil && len(stdout) > 0 {
		err = errors.New(stdout)
	}

	dstmd5, err1 := c.fileMd5(ctx, sudo, originDst)
	if err1 != nil {
		return err1
	}

	if dstmd5 != srcmd5 {
		return errors.New(fmt.Sprintf("md5 error: dstfile is %s, srcfile is %s", dstmd5, srcmd5))
	}
	return
}

func (c *HostConnection) fileMd5(ctx context.Context, sudo bool, filepath string) (md5 string, err error) {
	filepath = opsutils.GetAbsoluteFilePath(filepath)
	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", filepath)
	if sudo {
		cmd = fmt.Sprintf("sudo %s", cmd)
	}
	return c.execScript(ctx, sudo, cmd)
}

func (c *HostConnection) makeDir(ctx context.Context, sudo bool, filepath string) (err error) {
	_, err = c.execScript(ctx, sudo, opsutils.ShellMakeDir(opsutils.SplitDirPath(filepath)))
	return
}

func (c *HostConnection) getIDU(ctx context.Context) (idu string, err error) {
	return c.execScript(ctx, false, fmt.Sprintf("id -u"))
}

func (c *HostConnection) getIDG(ctx context.Context) (idg string, err error) {
	return c.execScript(ctx, false, fmt.Sprintf("id -g"))
}

func (c *HostConnection) getCPUTotal(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellCPUTotal())
}

func (c *HostConnection) getCPULoad1(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellCPULoad1())
}

func (c *HostConnection) getCPUUsagePercent(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellCPUUsagePercent())
}

func (c *HostConnection) getMemTotal(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellMemTotal())
}

func (c *HostConnection) getMemUsagePercent(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellMemUsagePercent())
}

func (c *HostConnection) getHosname(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellHostname())
}

func (c *HostConnection) getDiskTotal(ctx context.Context, sudo bool, timeoutSeconds int) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellDiskTotal(timeoutSeconds))
}

func (c *HostConnection) getArch(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellArch())
}

func (c *HostConnection) getDiskUsagePercent(ctx context.Context, sudo bool, timeoutSeconds int) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellDiskUsagePercent(timeoutSeconds))
}

func (c *HostConnection) getKernelVersion(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellKernelVersion())
}

func (c *HostConnection) getDistribution(ctx context.Context, sudo bool) (cpu string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellDistribution())
}

func (c *HostConnection) getAcceleratorVendor(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellAcceleratorVendor())
}

func (c *HostConnection) getAcceleratorModel(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellAcceleratorModel())
}

func (c *HostConnection) getAcceleratorCount(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execScript(ctx, sudo, opsutils.ShellAcceleratorCount())
}

func (c *HostConnection) getTempfileName(ctx context.Context, name string) string {
	nameSplit := strings.Split(name, "/")
	name = nameSplit[len(nameSplit)-1]
	cmd := "pwd"
	stdout, err := c.execScript(ctx, false, cmd)
	if err != nil {
		return name
	}
	return fmt.Sprintf("%s/.%s-%d", strings.TrimSpace(stdout), name, time.Now().UnixNano())
}
