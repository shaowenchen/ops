package host

import "fmt"

func NewHost(name string, address string, internalAddress string, port int, user string, password string, privateKey string, privateKeyPath string, timeout int64) (*Host, error) {
	if len(privateKeyPath) == 0{
		privateKeyPath = "~/.ssh/id_rsa"
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

	return host, nil
}

func (host *Host) AddHost(domain string, ip string) {
	fmt.Println("host.AddHost")
	stdout, _, _ := host.Exec(fmt.Sprintf("echo '%s %s' >> /etc/hosts", ip, domain))
	fmt.Println(stdout)
}
