package host

type HostOption struct {
	Hosts          string
	Port           int
	Username       string
	Password       string
	PrivateKey     string
	PrivateKeyPath string
}
type ScriptOption struct {
	HostOption
	Content string
	Sudo    bool
}
type FileOption struct {
	HostOption
	LocalFile  string
	RemoteFile string
	Direction  string
	Sudo       bool
}
