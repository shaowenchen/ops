package host

import (
	"bufio"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"net"

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
	Conn            HostConnection
}

func (host *Host) Connecting() error {
	authMethods := make([]ssh.AuthMethod, 0)
	if len(host.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(host.Password))
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
		Timeout:         time.Duration(host.Timeout),
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	targetHost := host.Address
	targetPort := strconv.Itoa(host.Port)

	endpoint := net.JoinHostPort(targetHost, targetPort)

	client, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}

	endpointBehindBastion := net.JoinHostPort(host.Address, strconv.Itoa(host.Port))

	conn, err := client.Dial("tcp", endpointBehindBastion)
	if err != nil {
		return errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	sshConfig.User = host.User
	ncc, chans, reqs, err := ssh.NewClientConn(conn, endpointBehindBastion, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	host.Conn.sshclient = ssh.NewClient(ncc, chans, reqs)
	sftpClient, err := sftp.NewClient(host.Conn.sshclient)
	if err != nil {
		return errors.Wrapf(err, "new sftp client failed: %v", err)
	}
	host.Conn.sftpclient = sftpClient
	return nil
}


func (host *Host) Exec(cmd string) (stdout string, code int, err error) {
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
