package kube

type ClearOption struct {
	Kubeconfig string
	Namespace  string
	Status     string
	All        bool
}
type DeschedulerOption struct {
	Kubeconfig       string
	Namespace        string
	RemoveDuplicates bool
	NodeUtilization  bool
	HighPercent      int16
	All              bool
}

type EtcHostsOption struct {
	Kubeconfig string
	Domain     string
	IP         string
	NodeName   string
	All        bool
	Clear      bool
}

type ScriptOption struct {
	Kubeconfig string
	NodeName   string
	All        bool
	Content    string
}

type ImagePulllSecretOption struct {
	Kubeconfig string
	Namespace  string
	Name       string
	Host       string
	Username   string
	Password   string
	Clear      bool
	All        bool
}

type LimitRangeOption struct {
	Kubeconfig string
	Namespace  string
	Name       string
	ReqMem     string
	LimitMem   string
	ReqCPU     string
	LimitCPU   string
	Clear      bool
	All        bool
}

type AnnotateOption struct {
	Kubeconfig string
	Type       string
	Namespace  string
	Clear      bool
	All        bool
}

type NodeNameOption struct {
	Kubeconfig string
	NodeName   string
	Name       string
	Clear      bool
}

type NodeSelectorOption struct {
	Kubeconfig string
	Name       string
	KeyLabels  string
	Clear      bool
}
