package host

type HostOption struct {
	Hosts          string
	Port           int
	Username       string
	Password       string
	PrivateKeyPath string
}
type ScriptOption struct {
	HostOption
	Content string
}

type FileOption struct {
	HostOption
	LocalFile  string
	RemoteFile string
	Direction  string
}