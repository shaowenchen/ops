package host

import (
	"bufio"
	"context"
	"time"

	"io/ioutil"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"golang.org/x/crypto/ssh"

	"net"

	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"strconv"
	"strings"
)

type Host struct {
	Name           string `yaml:"name,omitempty" json:"name,omitempty"`
	Address        string `yaml:"address,omitempty" json:"address,omitempty"`
	Port           int    `yaml:"port,omitempty" json:"port,omitempty"`
	Username       string `yaml:"user,omitempty" json:"username,omitempty"`
	Password       string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey     string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Timeout        int64  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Conn           *HostConnection
	Logger         *log.Logger
}

func NewHost(address string, port int, username string, password string, privateKeyPath string) (*Host, error) {
	if len(privateKeyPath) == 0 {
		privateKeyPath = constants.GetCurrentUserPrivateKeyPath()
	}
	if port == 0 {
		port = 22
	}
	if len(username) == 0 {
		username = constants.GetCurrentUser()
	}
	host := &Host{
		Name:           "",
		Address:        address,
		Port:           port,
		Username:       username,
		Password:       password,
		PrivateKey:     "",
		PrivateKeyPath: privateKeyPath,
		Timeout:        10,
	}
	// local host
	if address == constants.LocalHostIP {
		return host, nil
	}
	// remote host
	if err := host.connecting(); err != nil {
		fmt.Println("Failed connect host:", host.Address, err.Error())
		return nil, err
	}
	return host, nil
}

func (host *Host) connecting() (err error) {
	authMethods := make([]ssh.AuthMethod, 0)
	if len(host.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(host.Password))
	}

	if len(host.PrivateKey) == 0 && len(host.PrivateKeyPath) > 0 {
		content, err := ioutil.ReadFile(host.PrivateKeyPath)
		if err != nil {
			return errors.Wrapf(err, "Failed read keyfile %q", host.PrivateKeyPath)
		}
		host.PrivateKey = string(content)
	}
	if len(host.PrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey([]byte(host.PrivateKey))
		if parseErr != nil {
			return errors.Wrap(parseErr, "The given SSH key could not be parsed")
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	sshConfig := &ssh.ClientConfig{
		User:            host.Username,
		Timeout:         time.Duration(host.Timeout) * time.Second,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	endpointBehindBastion := net.JoinHostPort(host.Address, strconv.Itoa(host.Port))

	host.Conn = &HostConnection{}
	host.Conn.sshclient, err = ssh.Dial("tcp", endpointBehindBastion, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "client.Dial failed %s", host.Address)
	}
	host.Conn.scpclient, err = scp.NewClientBySSH(host.Conn.sshclient)
	if err != nil {
		fmt.Printf("scp.NewClient failed: %v\n", err)
	}
	return nil
}

func (host *Host) exec(sudo bool, cmd string) (stdout string, code int, err error) {
	// run in localhost
	if host.Address == constants.LocalHostIP {
		runner := exec.Command("sudo", "sh", "-c", cmd)
		if sudo {
			runner = exec.Command("sh", "-c", cmd)
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
	sess, err := host.Conn.session()
	if err != nil {
		return "", 1, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	exitCode := 0

	in, _ := sess.StdinPipe()
	out, _ := sess.StdoutPipe()
	err = sess.Start(utils.BuildBase64Cmd(sudo, cmd))
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
			_, err = in.Write([]byte(host.Password + "\n"))
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
	outStr := strings.TrimPrefix(string(output), fmt.Sprintf("[sudo] password for %s:", host.Username))

	// preserve original error
	return strings.TrimSpace(outStr), exitCode, errors.Wrapf(err, "Failed to exec command: %s \n%s", cmd, strings.TrimSpace(outStr))
}

func (host *Host) pullContent(sudo bool, src, dst string) (err error) {
	srcmd5, err := host.fileMd5(sudo, src)
	if err != nil {
		return err
	}
	output, _, err := host.exec(sudo, fmt.Sprintf("cat %s | base64 -w 0", src))
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

func (host *Host) pull(sudo bool, src, dst string) (err error) {
	srcmd5, err := host.fileMd5(sudo, src)
	if err != nil {
		return err
	}
	dst = utils.GetAbsoluteFilePath(dst)
	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	err = host.Conn.scpclient.CopyFromRemote(context.Background(), dstFile, src)

	if err != nil {
		return
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

func (host *Host) push(sudo bool, src, dst string) (err error) {
	if host.Address == constants.LocalHostIP {
		return errors.New("remote address is localhost")
	}
	srcmd5, err := utils.FileMD5(src)
	if err != nil {
		return err
	}
	src = utils.GetAbsoluteFilePath(src)
	srcFile, err1 := os.Open(src)
	err1 = host.Conn.scpclient.CopyFromFile(context.Background(), *srcFile, dst, "0655")

	if err1 != nil {
		return err1
	}
	dstmd5, err1 := host.fileMd5(sudo, dst)
	if err1 != nil {
		return err1
	}

	if dstmd5 != srcmd5 {
		return errors.New(fmt.Sprintf("md5 error: dstfile is %s, srcfile is %s", dstmd5, srcmd5))
	}
	return
}

func (host *Host) fileMd5(sudo bool, filepath string) (md5 string, err error) {
	filepath = utils.GetAbsoluteFilePath(filepath)
	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", filepath)
	if sudo {
		cmd = fmt.Sprintf("sudo %s", cmd)
	}
	stdout, _, err := host.exec(sudo, cmd)
	if err != nil {
		return
	}
	md5 = strings.TrimSpace(stdout)
	return
}

func (host *Host) Script(logger *log.Logger, sudo bool, content string) (stdout string, exit int, err error) {
	stdout, exit, err = host.exec(sudo, content)
	if len(stdout) != 0 {
		logger.Info.Println(stdout)
	}
	if exit != 0 {
		return "", 1, err
	}
	return
}

func (host *Host) File(logger *log.Logger, sudo bool, direction, localfile, remotefile string) (err error) {
	if utils.IsDownloadDirection(direction) {
		err = host.pullContent(sudo, remotefile, localfile)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
	} else if utils.IsUploadDirection(direction) {
		err = host.push(sudo, localfile, remotefile)
		if err != nil {
			logger.Error.Println(err)
		}
	} else {
		logger.Error.Println("invalid file transfer direction")
	}
	return
}
