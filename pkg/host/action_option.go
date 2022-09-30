package host

type HostOption struct {
	Hosts          string
	Username       string
	Password       string
	PrivateKeyPath string
	Clear          bool
}
type ScriptOption struct {
	HostOption
	Content string
}
type EtcHostsOption struct {
	HostOption
	Domain string
	IP     string
}

type InstallOption struct {
	HostOption
	Name string
}

type KubeconfigOption struct {
	HostOption
}
