package host

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type HostConnection struct {
	sftpclient *sftp.Client
	sshclient  *ssh.Client
}

func (c *HostConnection) session() (*ssh.Session, error) {
	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}
	fmt.Println("Creating Session to host")
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
	fmt.Println("Created Session to host")
	return sess, nil
}
