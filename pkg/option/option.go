package option

import (
	"fmt"
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

type MountConfig struct {
	HostPath  string
	MountPath string
	Secret    *SecretMountConfig
	ConfigMap *ConfigMapMountConfig
}

// SecretMountConfig defines a secret mount configuration
type SecretMountConfig struct {
	Name      string
	MountPath string
}

// ConfigMapMountConfig defines a configMap mount configuration
type ConfigMapMountConfig struct {
	Name      string
	MountPath string
}

type KubeOption struct {
	Debug        bool
	Namespace    string
	NodeName     string
	RuntimeImage string
	Mounts       []MountConfig
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

// ParseMountConfig parses a mount configuration string
// format: hostPath:mountPath
// example: /opt/data:/data
func ParseMountConfig(mountStr string) (MountConfig, error) {
	parts := strings.Split(mountStr, ":")
	if len(parts) != 2 {
		return MountConfig{}, fmt.Errorf("invalid mount format: %s, expected hostPath:mountPath", mountStr)
	}

	hostPath := strings.TrimSpace(parts[0])
	mountPath := strings.TrimSpace(parts[1])

	// validate path format
	if !strings.HasPrefix(hostPath, "/") {
		return MountConfig{}, fmt.Errorf("hostPath must be absolute path: %s", hostPath)
	}
	if !strings.HasPrefix(mountPath, "/") {
		return MountConfig{}, fmt.Errorf("mountPath must be absolute path: %s", mountPath)
	}

	return MountConfig{
		HostPath:  hostPath,
		MountPath: mountPath,
	}, nil
}

// ParseMountConfigs parses multiple mount configurations
func ParseMountConfigs(mountStrs []string) ([]MountConfig, error) {
	configs := make([]MountConfig, 0, len(mountStrs))
	for _, str := range mountStrs {
		config, err := ParseMountConfig(str)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}
