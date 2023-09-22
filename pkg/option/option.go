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
}

type KubeOption struct {
	NodeName     string
	RuntimeImage string
	All          bool
	InCluster    bool
}

type S3FileOption struct {
	Region   string
	Endpoint string
	Bucket   string
	AK       string
	SK       string
}

type TaskOption struct {
	Sudo      bool
	FilePath  string
	Variables map[string]string
}

type ShellOption struct {
	Content string
	Sudo    bool
}

type CopilotOption struct {
	Endpoint string
	Model    string
	Key      string
	History  int
	Silence  bool
}

type FileOption struct {
	KubeOption
	LocalFile    string
	RemoteFile   string
	StorageType  string
	StorageImage string
	Direction    string
	Sudo         bool
}

func (f *FileOption) Filling() {
	remoteSplit := strings.Split(f.RemoteFile, "://")
	if len(remoteSplit) == 1 {
		f.StorageType = constants.RemoteStorageTypeLocal
		return
	} else if remoteSplit[0] == "s3" {
		f.StorageType = constants.RemoteStorageTypeS3
		f.RemoteFile = remoteSplit[1]
		return
	} else {
		f.StorageType = constants.RemoteStorageTypeImage
		f.StorageImage = remoteSplit[0]
		f.RemoteFile = remoteSplit[1]
		return
	}
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
