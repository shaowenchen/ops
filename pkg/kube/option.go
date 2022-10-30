package kube

type KubeOption struct {
	Kubeconfig string
	NodeName   string
	All        bool
}

type ScriptOption struct {
	KubeOption
	Content string
}

type FileOption struct {
	KubeOption
	LocalFile  string
	RemoteFile string
	Image      string
}
