package option

import (
	"strings"

	"github.com/shaowenchen/ops/pkg/constants"
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
	return strings.ToLower(k.NodeName) == "all"
}

type TaskOption struct {
	Sudo      bool
	FilePath  string
	Proxy     string
	Variables map[string]string
	Clear     bool
}

type ShellOption struct {
	Content string
	Sudo    bool
}

type CopilotOption struct {
	Endpoint     string
	Model        string
	Key          string
	History      int
	Silence      bool
	OpsServer    string
	OpsToken     string
	RuntimeImage string
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
		f.StorageType = constants.RemoteStorageTypeServer
	} else if remoteSplit[0] == "s3" {
		f.StorageType = constants.RemoteStorageTypeS3
		f.RemoteFile = remoteSplit[1]
	} else {
		f.StorageType = constants.RemoteStorageTypeImage
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
	Kubeconfig string
	Clear      bool
}
