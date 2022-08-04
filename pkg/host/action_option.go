package host

type EtcHostsOption struct {
	Input          string
	Username       string
	Domain         string
	IP             string
	PrivateKeyPath string
	Clear          bool
}

type InstallOption struct {
	Input          string
	Username       string
	Password       string
	PrivateKeyPath string
	Name           string
	Clear          bool
}

type KubeconfigOption struct {
	Input          string
	Username       string
	Password       string
	PrivateKeyPath string
	Clear          bool
}
