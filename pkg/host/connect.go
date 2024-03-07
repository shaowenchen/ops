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
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HostConnection struct {
	Host      *opsv1.Host
	scpclient scp.Client
	sshclient *ssh.Client
}

func NewHostConnBase64(h *opsv1.Host) (hc *HostConnection, err error) {
	if h == nil {
		h = &opsv1.Host{}
	}
	hc = &HostConnection{}
	hc.Host = h
	// empty address is local host
	if h.Spec.Address == "" {
		h.Spec.Address = constants.LocalHostIP
	}
	// local host
	if h.Spec.Address == constants.LocalHostIP {
		return hc, nil
	}
	// remote host
	if err := hc.connecting(); err != nil {
		return hc, err
	}
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
	return c.execSh(ctx, sudo, content)
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
			proxy = constants.DefaultProxy
		}
		return c.execSh(ctx, sudo, utils.ShellInstallOpscli(proxy))
	}
	return
}

func (c *HostConnection) isInChina(ctx context.Context) (ok bool) {
	_, err := c.execSh(ctx, false, utils.ShellIsInChina())
	if err != nil {
		return true
	}
	return false
}

func (c *HostConnection) File(ctx context.Context, sudo bool, direction, localfile, remotefile string) (err error) {
	if utils.IsDownloadDirection(direction) {
		err = c.scpPull(ctx, sudo, remotefile, localfile)
		if err != nil {
			return err
		}
	} else if utils.IsUploadDirection(direction) {
		err = c.scpPush(ctx, sudo, localfile, remotefile)
		if err != nil {
			return err
		}
	} else {
		return errors.New("invalid file transfer direction")
	}
	return
}

func (c *HostConnection) GetStatus(ctx context.Context, sudo bool) (status *opsv1.HostStatus, err error) {
	hostname, err1 := c.getHosname(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel := context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	kerneVersion, err1 := c.getKernelVersion(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	distribution, err1 := c.getDistribution(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	arch, err1 := c.getArch(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	diskTotal, err1 := c.getDiskTotal(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	diskUsagePercent, err1 := c.getDiskUsagePercent(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	cpuTotal, err1 := c.getCPUTotal(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	cpuLoad1, err1 := c.getCPULoad1(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	cpuUsagePercent, err1 := c.getCPUUsagePercent(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	memTotal, err1 := c.getMemTotal(ctx, sudo)
	err = utils.MergeError(err, err1)

	ctx, cancel = context.WithTimeout(ctx, constants.DefaultShellTimeoutDuration)
	defer cancel()
	memUsagePercent, err1 := c.getMemUsagePercent(ctx, sudo)
	err = utils.MergeError(err, err1)

	status = &opsv1.HostStatus{
		Hostname:         hostname,
		KernelVersion:    kerneVersion,
		Distribution:     distribution,
		Arch:             arch,
		DiskTotal:        diskTotal,
		DiskUsagePercent: diskUsagePercent,
		CPUTotal:         cpuTotal,
		CPULoad1:         cpuLoad1,
		CPUUsagePercent:  cpuUsagePercent,
		MemTotal:         memTotal,
		MemUsagePercent:  memUsagePercent,
		HeartTime:        &metav1.Time{Time: time.Now()},
		HeartStatus:      opsv1.StatusSuccessed,
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
	password, err := utils.DecodingBase64ToString(c.Host.Spec.Password)
	if err != nil {
		return err
	}
	privateKey, err := utils.DecodingBase64ToString(c.Host.Spec.PrivateKey)
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
		Timeout:         time.Duration(c.Host.GetSpec().TimeOutSeconds) * time.Second,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	endpointBehindBastion := net.JoinHostPort(c.Host.Spec.Address, strconv.Itoa(c.Host.Spec.Port))

	c.sshclient, err = ssh.Dial("tcp", endpointBehindBastion, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "client.Dial failed %s", c.Host.Spec.Address)
	}
	c.scpclient, err = scp.NewClientBySSH(c.sshclient)
	if err != nil {
		return errors.Wrapf(err, "scp.NewClient failed")
	}
	return nil
}

func (c *HostConnection) execSh(ctx context.Context, sudo bool, cmd string) (stdout string, err error) {
	return c.ExecWithExecutor(ctx, sudo, "sh", "-c", cmd)
}

func (c *HostConnection) ExecWithExecutor(ctx context.Context, sudo bool, executor, param, cmd string) (stdout string, err error) {
	cmd = utils.BuildBase64CmdWithExecutor(sudo, cmd, executor)
	// run in localhost
	if c.Host.Spec.Address == constants.LocalHostIP {
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
	return c.execSh(ctx, sudo, utils.ShellMv(src, dst))
}

func (c *HostConnection) copy(ctx context.Context, sudo bool, src, dst string) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellCopy(src, dst))
}

func (c *HostConnection) chown(ctx context.Context, sudo bool, idU, idG, src string) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellChown(idU, idG, src))
}

func (c *HostConnection) rm(ctx context.Context, sudo bool, dst string) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellRm(dst))
}

func (c *HostConnection) cmdPull(ctx context.Context, sudo bool, src, dst string) (err error) {
	srcmd5, err := c.fileMd5(ctx, sudo, src)
	if err != nil {
		return err
	}
	output, err := c.execSh(ctx, sudo, fmt.Sprintf("cat %s | base64 -w 0", src))
	if err != nil {
		return fmt.Errorf("open src file failed %v, src path: %s", err, src)
	}
	dstDir := filepath.Dir(dst)
	if utils.IsExistsFile(dstDir) {
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

	dstmd5, err := utils.FileMD5(dst)
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
	dst = utils.GetAbsoluteFilePath(dst)
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

	dstmd5, err := utils.FileMD5(dst)
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
	if c.Host.Spec.Address == constants.LocalHostIP {
		return errors.New("remote address is localhost")
	}
	srcmd5, err := utils.FileMD5(src)
	if err != nil {
		return err
	}
	err = c.makeDir(ctx, sudo, originDst)
	if err != nil {
		return err
	}
	src = utils.GetAbsoluteFilePath(src)
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
	filepath = utils.GetAbsoluteFilePath(filepath)
	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", filepath)
	if sudo {
		cmd = fmt.Sprintf("sudo %s", cmd)
	}
	return c.execSh(ctx, sudo, cmd)
}

func (c *HostConnection) makeDir(ctx context.Context, sudo bool, filepath string) (err error) {
	_, err = c.execSh(ctx, sudo, utils.ShellMakeDir(utils.SplitDirPath(filepath)))
	return
}

func (c *HostConnection) getIDU(ctx context.Context) (idu string, err error) {
	return c.execSh(ctx, false, fmt.Sprintf("id -u"))
}

func (c *HostConnection) getIDG(ctx context.Context) (idg string, err error) {
	return c.execSh(ctx, false, fmt.Sprintf("id -g"))
}

func (c *HostConnection) getCPUTotal(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellCPUTotal())
}

func (c *HostConnection) getCPULoad1(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellCPULoad1())
}

func (c *HostConnection) getCPUUsagePercent(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellCPUUsagePercent())
}

func (c *HostConnection) getMemTotal(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellMemTotal())
}

func (c *HostConnection) getMemUsagePercent(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellMemUsagePercent())
}

func (c *HostConnection) getHosname(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellHostname())
}

func (c *HostConnection) getDiskTotal(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellDiskTotal())
}

func (c *HostConnection) getArch(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellArch())
}

func (c *HostConnection) getDiskUsagePercent(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellDiskUsagePercent())
}

func (c *HostConnection) getKernelVersion(ctx context.Context, sudo bool) (stdout string, err error) {
	return c.execSh(ctx, sudo, utils.ShellKernelVersion())
}

func (c *HostConnection) getDistribution(ctx context.Context, sudo bool) (cpu string, err error) {
	return c.execSh(ctx, sudo, utils.ShellDistribution())
}

func (c *HostConnection) getTempfileName(ctx context.Context, name string) string {
	nameSplit := strings.Split(name, "/")
	name = nameSplit[len(nameSplit)-1]
	cmd := "pwd"
	stdout, err := c.execSh(ctx, false, cmd)
	if err != nil {
		return name
	}
	return fmt.Sprintf("%s/.%s-%d", strings.TrimSpace(stdout), name, time.Now().UnixNano())
}
