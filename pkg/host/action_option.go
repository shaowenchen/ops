package host

type HostOption struct {
	Hosts          string
	Port           int
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

type FileOption struct {
	HostOption
	LocalFile  string
	RemoteFile string
	Direction  string
	Overwrite  bool
}
