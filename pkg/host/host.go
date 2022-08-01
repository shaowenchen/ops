package host

import (
	"bufio"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"

	"io"
	"net"

	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"strconv"
	"strings"
)

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
		user = "root"
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
	if err := host.connecting(); err != nil {
		fmt.Println("Failed to connect to host:", host.Address, err.Error())
		return nil, err
	}
	return host, nil
}

func (host *Host) connecting() error {
	fmt.Println("Connecting to host:", host.Address)
	authMethods := make([]ssh.AuthMethod, 0)
	if len(host.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(host.Password))
	}

	if len(host.PrivateKey) == 0 && len(host.PrivateKeyPath) > 0 {
		content, err := ioutil.ReadFile(host.PrivateKeyPath)
		if err != nil {
			return errors.Wrapf(err, "Failed to read keyfile %q", host.PrivateKeyPath)
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
	fmt.Println("Connected to host:", host.Address)
	return nil
}

func (host *Host) exec(cmd string) (stdout string, code int, err error) {
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

func (host *Host) scp(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := host.Conn.sftpclient.Create(dst)
	if err != nil {
		return err
	}
	fileStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed %v", err)
	}
	if err := dstFile.Chmod(fileStat.Mode()); err != nil {
		return fmt.Errorf("chmod remote file failed %v", err)
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func (host *Host) Fetch(src, dst string) error {
	output, _, err := host.exec(fmt.Sprintf("sudo cat %s | base64 -w 0", src))
	if err != nil {
		return fmt.Errorf("open src file failed %v, src path: %s", err, src)
	}
	dstDir := filepath.Dir(dst)
	if isExist, _ := PathExists(dstDir); !isExist {
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

	return nil
}
