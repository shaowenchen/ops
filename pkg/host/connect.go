package host

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	"github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type HostConnection struct {
	Host      *v1.Host
	scpclient scp.Client
	sshclient *ssh.Client
}

func NewHostConnection(address string, port int, username string, password string, privateKey string, privateKeyPath string) (c *HostConnection, err error) {

	if len(privateKeyPath) == 0 {
		privateKeyPath = constants.GetCurrentUserPrivateKeyPath()
	}
	if port == 0 {
		port = 22
	}
	if len(username) == 0 {
		username = constants.GetCurrentUser()
	}
	c = &HostConnection{
		Host: v1.NewHost(
			"", "", address, port, username, password, privateKey, privateKeyPath,
		),
	}
	// local host
	if address == constants.LocalHostIP {
		return c, nil
	}
	// remote host
	if err := c.connecting(); err != nil {
		return c, err
	}
	return
}

func (c *HostConnection) Script(sudo bool, content string) (stdout string, exit int, err error) {
	reg := regexp.MustCompile(`\${[^\}]*}`)
	funcStrList := reg.FindAllString(content, -1)
	for _, callFunc := range funcStrList {
		rawCallFunc := callFunc
		callFunc = callFunc[2 : len(callFunc)-1]
		stdout, err = c.scriptFuncMap(sudo, callFunc)
		if err != nil {
			return stdout, -1, err
		}
		content = strings.ReplaceAll(content, rawCallFunc, stdout)
	}
	stdout, exit, err = c.exec(sudo, content)
	if exit != 0 && len(stdout) == 0 {
		return err.Error(), 1, err
	}
	return
}

func (c *HostConnection) scriptFuncMap(sudo bool, funcFull string) (stdout string, err error) {
	if funcFull == "installOpscli()" {
		return c.install(sudo, "opscli")
	} else if funcFull == "distribution()" {
		return c.getDistribution(sudo)
	}
	return
}

func (c *HostConnection) install(sudo bool, component string) (stdout string, err error) {
	if component == "opscli" {
		proxy := ""
		if !c.isInChina() {
			proxy = constants.DefaultProxy
		}
		stdout, _, err = c.exec(sudo, utils.ScriptInstallOpscli(proxy))
		return
	}
	return
}

func (c *HostConnection) isInChina() (ok bool) {
	_, exitCode, err := c.exec(false, utils.ScriptIsInChina())
	if exitCode != 0 || err != nil {
		return false
	}
	return true
}

func (c *HostConnection) File(sudo bool, direction, localfile, remotefile string) (err error) {
	if utils.IsDownloadDirection(direction) {
		err = c.scpPull(sudo, remotefile, localfile)
		if err != nil {
			return err
		}
	} else if utils.IsUploadDirection(direction) {
		err = c.scpPush(sudo, localfile, remotefile)
		if err != nil {
			return err
		}
	} else {
		return errors.New("invalid file transfer direction")
	}
	return
}

func (c *HostConnection) UpdateStatus(sudo bool) (err error) {
	hostname, _ := c.getHosname(sudo)
	kerneVersion, _ := c.getKernelVersion(sudo)
	distribution, _ := c.getDistribution(sudo)
	diskTotal, _ := c.getDiskTotal(sudo)
	diskUsagePercent, _ := c.getDiskUsagePercent(sudo)
	cpuTotal, _ := c.getCPUTotal(sudo)
	cpuLoad1, _ := c.getCPULoad1(sudo)
	cpuUsagePercent, _ := c.getCPUUsagePercent(sudo)
	memTotal, _ := c.getMemTotal(sudo)
	memUsagePercent, _ := c.getMemUsagePercent(sudo)

	c.Host.Status.Hostname = hostname
	c.Host.Status.KernelVersion = kerneVersion
	c.Host.Status.Distribution = distribution
	c.Host.Status.DiskTotal = diskTotal
	c.Host.Status.DiskUsagePercent = diskUsagePercent
	c.Host.Status.CPUTotal = cpuTotal
	c.Host.Status.CPULoad1 = cpuLoad1
	c.Host.Status.CPUUsagePercent = cpuUsagePercent
	c.Host.Status.MemTotal = memTotal
	c.Host.Status.MemUsagePercent = memUsagePercent
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
	authMethods := make([]ssh.AuthMethod, 0)
	if len(c.Host.Spec.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(c.Host.Spec.Password))
	}

	if len(c.Host.Spec.PrivateKey) == 0 && len(c.Host.Spec.PrivateKeyPath) > 0 {
		content, err := ioutil.ReadFile(c.Host.Spec.PrivateKeyPath)
		if err != nil {
			return errors.New("Failed read keyfile")
		}
		c.Host.Spec.PrivateKey = string(content)
	}
	if len(c.Host.Spec.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey([]byte(c.Host.Spec.PrivateKey))
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

func (c *HostConnection) exec(sudo bool, cmd string) (stdout string, code int, err error) {
	cmd = utils.BuildBase64Cmd(sudo, cmd)
	// run in localhost
	if c.Host.Spec.Address == constants.LocalHostIP {
		runner := exec.Command("sh", "-c", cmd)
		if sudo {
			runner = exec.Command("sudo", "sh", "-c", cmd)
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
		return "", 1, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	exitCode := 0

	in, _ := sess.StdinPipe()
	out, _ := sess.StdoutPipe()
	err = sess.Start(cmd)
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
		return "", exitCode, err
	}

	var (
		output []byte
		line   = ""
		r      = bufio.NewReader(out)
	)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		output = append(output, b)

		if b == byte('\n') {
			line = ""
			continue
		}

		line += string(b)

		if (strings.HasPrefix(line, "[sudo] password for ") || strings.HasPrefix(line, "Password")) && strings.HasSuffix(line, ": ") {
			_, err = in.Write([]byte(c.Host.Spec.Password + "\n"))
			if err != nil {
				break
			}
		}
	}
	err = sess.Wait()
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}
	outStr := strings.TrimPrefix(string(output), fmt.Sprintf("[sudo] password for %s:", c.Host.Spec.Username))

	// preserve original error
	return strings.TrimSpace(outStr), exitCode, errors.Wrapf(err, "Failed to exec command: %s \n%s", cmd, strings.TrimSpace(outStr))
}

func (c *HostConnection) mv(sudo bool, src, dst string) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptMv(src, dst))
	return
}

func (c *HostConnection) copy(sudo bool, src, dst string) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptCopy(src, dst))
	return
}

func (c *HostConnection) chown(sudo bool, idU, idG, src string) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptChown(idU, idG, src))
	fmt.Println(stdout)
	return
}

func (c *HostConnection) rm(sudo bool, dst string) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptRm(dst))
	return
}

func (c *HostConnection) cmdPull(sudo bool, src, dst string) (err error) {
	srcmd5, err := c.fileMd5(sudo, src)
	if err != nil {
		return err
	}
	output, _, err := c.exec(sudo, fmt.Sprintf("cat %s | base64 -w 0", src))
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

func (c *HostConnection) scpPull(sudo bool, src, dst string) (err error) {
	originSrc := src
	src = c.getTempfileName(src)
	stdout, err := c.copy(sudo, originSrc, src)
	if err != nil {
		return errors.New(stdout)
	}
	idU, err := c.getIdU()
	if err != nil {
		return errors.New(stdout)
	}
	idG, err := c.getIdG()
	if err != nil {
		return errors.New(stdout)
	}
	stdout, err = c.chown(sudo, idU, idG, src)
	if err != nil {
		return errors.New(stdout)
	}
	srcmd5, err := c.fileMd5(sudo, originSrc)
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

	stdout, err = c.rm(sudo, src)
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

func (c *HostConnection) scpPush(sudo bool, src, dst string) (err error) {
	originDst := dst
	dst = c.getTempfileName(dst)
	if c.Host.Spec.Address == constants.LocalHostIP {
		return errors.New("remote address is localhost")
	}
	srcmd5, err := utils.FileMD5(src)
	if err != nil {
		return err
	}
	err = c.makeDir(sudo, originDst)
	if err != nil {
		return err
	}
	src = utils.GetAbsoluteFilePath(src)
	srcFile, err := os.Open(src)
	err = c.scpclient.CopyFromFile(context.Background(), *srcFile, dst, "0655")

	if err != nil {
		return err
	}
	stdout, err := c.mv(sudo, dst, originDst)
	if err == nil && len(stdout) > 0 {
		err = errors.New(stdout)
	}

	dstmd5, err1 := c.fileMd5(sudo, originDst)
	if err1 != nil {
		return err1
	}

	if dstmd5 != srcmd5 {
		return errors.New(fmt.Sprintf("md5 error: dstfile is %s, srcfile is %s", dstmd5, srcmd5))
	}
	return
}

func (c *HostConnection) fileMd5(sudo bool, filepath string) (md5 string, err error) {
	filepath = utils.GetAbsoluteFilePath(filepath)
	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", filepath)
	if sudo {
		cmd = fmt.Sprintf("sudo %s", cmd)
	}
	stdout, _, err := c.exec(sudo, cmd)
	if err != nil {
		return
	}
	md5 = strings.TrimSpace(stdout)
	return
}

func (c *HostConnection) makeDir(sudo bool, filepath string) (err error) {
	cmd := utils.ScriptMakeDir(utils.SplitDirPath(filepath))
	stdout, _, err := c.exec(false, cmd)
	if err == nil && len(stdout) > 0 {
		return errors.New(stdout)
	}
	return
}

func (c *HostConnection) getIdU() (idu string, err error) {
	cmd := fmt.Sprintf("id -u")
	stdout, _, err := c.exec(false, cmd)
	if err != nil {
		return
	}
	return stdout, err
}

func (c *HostConnection) getIdG() (idg string, err error) {
	cmd := fmt.Sprintf("id -g")
	stdout, _, err := c.exec(false, cmd)
	if err != nil {
		return
	}
	return stdout, err
}

func (c *HostConnection) getCPUTotal(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptCPUTotal())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getCPULoad1(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptCPULoad1())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getCPUUsagePercent(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptCPUUsagePercent())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getMemTotal(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptMemTotal())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getMemUsagePercent(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptMemUsagePercent())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getHosname(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptHostname())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getDiskTotal(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptDiskTotal())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getDiskUsagePercent(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptDiskUsagePercent())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getKernelVersion(sudo bool) (stdout string, err error) {
	stdout, _, err = c.exec(sudo, utils.ScriptKernelVersion())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getDistribution(sudo bool) (cpu string, err error) {
	cpu, _, err = c.exec(sudo, utils.ScriptDistribution())
	if err != nil {
		return
	}
	return
}

func (c *HostConnection) getTempfileName(name string) string {
	nameSplit := strings.Split(name, "/")
	name = nameSplit[len(nameSplit)-1]
	cmd := "pwd"
	stdout, _, err := c.exec(false, cmd)
	if err != nil {
		return name
	}
	return fmt.Sprintf("%s/.%s-%d", strings.TrimSpace(stdout), name, time.Now().UnixNano())
}
