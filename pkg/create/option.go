package create

import (
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/host"
)

type HostOption struct {
	host.HostOption
	Namespace  string
	Name       string
	Kubeconfig string
	Clear      bool
}

type ClusterOption struct {
	opsv1.ClusterSpec
	Namespace  string
	Name       string
	Kubeconfig string
	Clear      bool
}

type TaskOption struct {
	Namespace  string
	HostRef    string
	Name       string
	Filepath   string
	Kubeconfig string
	Clear      bool
}
