package option

import (
	"strings"

	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

type HostOption struct {
	Host           string
	Port           int
	Username       string
	Password       string
	PrivateKey     string
	PrivateKeyPath string
	SecretRef      string
}

type KubeOption struct {
	Debug        bool
	Namespace    string
	NodeName     string
	RuntimeImage string
}

func (k *KubeOption) IsAllNodes() bool {
	return strings.ToLower(k.NodeName) == opsconstants.AllNodes
}

func (k *KubeOption) IsAllMasters() bool {
	return strings.ToLower(k.NodeName) == opsconstants.AllMasters
}

func (k *KubeOption) IsAllWorkers() bool {
	return strings.ToLower(k.NodeName) == opsconstants.AllWorkers
}

func (k *KubeOption) IsAnyNode() bool {
	return strings.ToLower(k.NodeName) == opsconstants.AnyNode
}

func (k *KubeOption) IsAnyMaster() bool {
	return strings.ToLower(k.NodeName) == opsconstants.AnyMaster
}

func (k *KubeOption) IsAnyWorker() bool {
	return strings.ToLower(k.NodeName) == opsconstants.AnyWorker
}

type TaskOption struct {
	Sudo      bool
	FilePath  string
	Proxy     string
	Variables map[string]string
	Clear     bool
}

type ShellOption struct {
	Mode    string
	Content string
	Sudo    bool
}

type FileOption struct {
	KubeOption
	LocalFile    string
	RemoteFile   string
	StorageType  string
	StorageImage string
	Direction    string
	Sudo         bool
	AesKey       string
	Api          string
	Region       string
	Endpoint     string
	Bucket       string
	AK           string
	SK           string
}

func (f *FileOption) GetStorageType() string {
	if f.StorageType != "" {
		return f.StorageType
	}
	remoteSplit := strings.Split(f.RemoteFile, "://")
	if len(f.Api) != 0 {
		f.StorageType = opsconstants.RemoteStorageTypeServer
	} else if remoteSplit[0] == "s3" {
		f.StorageType = opsconstants.RemoteStorageTypeS3
		f.RemoteFile = remoteSplit[1]
	} else if len(remoteSplit) == 2 {
		f.StorageType = opsconstants.RemoteStorageTypeImage
		f.RuntimeImage = remoteSplit[0]
		f.RemoteFile = remoteSplit[1]
	}
	return f.StorageType
}

func (f *FileOption) IsUploadDirection() bool {
	return strings.Contains(strings.ToLower(f.Direction), "up")
}

func (f *FileOption) IsDownloadDirection() bool {
	return strings.Contains(strings.ToLower(f.Direction), "down")
}

type KubernetesOption struct {
	Action   string
	Kind     string
	Metadata struct {
		Name      string
		Namespace string
	}
}

type PrometheusOption struct {
	Endpoint string
	Query    string
}

type ClusterOption struct {
	Namespace  string
	Name       string
	Desc       string
	Kubeconfig string
	Clear      bool
}
