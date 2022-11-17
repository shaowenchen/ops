package create

import (
	opsv1 "github.com/shaowenchen/ops/api/v1"
)

type HostOption struct {
	opsv1.HostSpec
	Hosts      string
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
