package host

import (
	"bufio"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"

	"io"
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

const LocalHostIP = "127.0.0.1"

type Host struct {
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	Address         string `yaml:"address,omitempty" json:"address,omitempty"`
	InternalAddress string `yaml:"internalAddress,omitempty" json:"internalAddress,omitempty"`
	Port            int    `yaml:"port,omitempty" json:"port,omitempty"`
	User            string `yaml:"user,omitempty" json:"user,omitempty"`
	Password        string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey      string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath  string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Timeout         int64  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Conn            *HostConnection
}

func newHost(name string, address string, internalAddress string, port int, user string, password string, privateKey string, privateKeyPath string, timeout int64) (*Host, error) {
	if len(privateKeyPath) == 0 {
		privateKeyPath = GetCurrentUserPrivateKeyPath()
	}
	if port == 0 {
		port = 22
	}

	if len(user) == 0 {
		user = GetCurrentUser()
	}

	if timeout == 0 {
		timeout = 10
	}
	host := &Host{
		Name:            name,
		Address:         address,
		InternalAddress: internalAddress,
		Port:            port,
		User:            user,
		Password:        password,
		PrivateKey:      privateKey,
		PrivateKeyPath:  privateKeyPath,
		Timeout:         timeout,
	}
	// local host
	if name == LocalHostIP || address == LocalHostIP {
		return host, nil
	}
	// remote host
	if err := host.connecting(); err != nil {
		fmt.Println("Failed connect host:", host.Address, err.Error())
		return nil, err
	}
	return host, nil
}

func (host *Host) connecting() error {
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
		User:            host.User,
		Timeout:         time.Duration(host.Timeout) * time.Second,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	targetHost := host.Address
	targetPort := strconv.Itoa(host.Port)

	endpoint := net.JoinHostPort(targetHost, targetPort)

	client, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "ssh.Dial failed %s", endpoint)
	}

	endpointBehindBastion := net.JoinHostPort(host.Address, strconv.Itoa(host.Port))

	conn, err := client.Dial("tcp", endpointBehindBastion)
	if err != nil {
		return errors.Wrapf(err, "client.Dial failed %s", endpointBehindBastion)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, endpointBehindBastion, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "ssh.NewClientConn failed %s", endpointBehindBastion)
	}
	host.Conn = &HostConnection{}
	host.Conn.sshclient = ssh.NewClient(ncc, chans, reqs)
	sftpClient, err := sftp.NewClient(host.Conn.sshclient)
	if err != nil {
		fmt.Printf("sftp.NewClient failed: %v\n", err)
	}
	host.Conn.sftpclient = sftpClient
	return nil
}

func (host *Host) exec(cmd string) (stdout string, code int, err error) {
	// run in localhost
	if host.Name == LocalHostIP || host.Address == LocalHostIP || host.InternalAddress == LocalHostIP {
		runner := exec.Command("sh", "-c", cmd)
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
	err = sess.Start(strings.TrimSpace(cmd))
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
	outStr := strings.TrimPrefix(string(output), fmt.Sprintf("[sudo] password for %s:", host.User))

	// preserve original error
	return strings.TrimSpace(outStr), exitCode, errors.Wrapf(err, "Failed to exec command: %s \n%s", cmd, strings.TrimSpace(outStr))
}

func (host *Host) pullContent(src, dst string) (size string, err error) {
	output, _, err := host.exec(fmt.Sprintf("sudo cat %s | base64 -w 0", src))
	if err != nil {
		return "", fmt.Errorf("open src file failed %v, src path: %s", err, src)
	}
	dstDir := filepath.Dir(dst)
	if isExist, _ := isExistsFile(dstDir); !isExist {
		err = os.MkdirAll(dstDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("create dst dir failed %v, dst dir: %s", err, dstDir)
		}
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create dst file failed %v", err)
	}
	defer dstFile.Close()

	if base64Str, err := base64.StdEncoding.DecodeString(output); err != nil {
		return "", err
	} else {
		if _, err = dstFile.WriteString(string(base64Str)); err != nil {
			return "", err
		}
	}

	return "", nil
}

func (host *Host) pull(src, dst string) (size string, err error) {
	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	srcFile, err := host.Conn.sftpclient.Open(src)
	if err != nil {
		return
	}

	transferBytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		size = humanize.Bytes(uint64(transferBytes))
		return
	}

	err = dstFile.Sync()
	if err != nil {
		return
	}
	return
}

func (host *Host) push(src, dst string) (size string, err error) {
	dstFile, err := host.Conn.sftpclient.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return
	}

	transferBytes, err := io.Copy(dstFile, srcFile)
	size = humanize.Bytes(uint64(transferBytes))
	return
}
