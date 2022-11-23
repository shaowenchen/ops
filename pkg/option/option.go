package option

import (
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
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
}

type S3FileOption struct {
	Region   string
	Endpoint string
	Bucket   string
	AK       string
	SK       string
}

type TaskOption struct {
	Debug     bool
	Sudo      bool
	FilePath  string
	Variables map[string]string
}

type ScriptOption struct {
	KubeOption
	Script string
	Sudo   bool
}

type FileOption struct {
	KubeOption
	LocalFile  string
	RemoteFile string
	Direction  string
	Sudo       bool
}

func (option *FileOption) GetStorageType() string {
	remoteSplit := strings.Split(option.RemoteFile, "://")
	if len(remoteSplit) == 1 {
		return constants.RemoteStorageTypeLocal
	}
	if remoteSplit[0] == "s3" {
		return constants.RemoteStorageTypeS3
	}
	return constants.RemoteStorageTypeImage
}

type ClusterOption struct {
	Namespace  string
	Name       string
	Kubeconfig string
	Clear      bool
}

type CreateHostOption struct {
	ClusterOption
	HostOption
}

type CreateClusterOption struct {
	ClusterOption
	opsv1.ClusterSpec
}

type CreateTaskOption struct {
	ClusterOption
	Filepath string
}
