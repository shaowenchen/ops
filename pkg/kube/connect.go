package kube

import (
	"context"
	"errors"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

type KubeConnection struct {
	Cluster    *opsv1.Cluster
	Client     *kubernetes.Clientset
	RestConfig *rest.Config
}

func NewKubeConnection(c *opsv1.Cluster) (kc *KubeConnection, err error) {
	if c == nil {
		return kc, errors.New("cluster is nil")
	}
	kc = &KubeConnection{
		Cluster: c,
	}
	// try config
	config, err := utils.DecodingBase64ToString(c.Spec.Config)
	if err != nil {
		return
	}
	kc.RestConfig, err = utils.GetRestConfigByContent(config)
	if err != nil {
		return
	}
	kc.Client, err = utils.GetClientByRestconfig(kc.RestConfig)
	if err != nil {
		return
	}
	// try others
	return
}

func (kc *KubeConnection) GetStatus() (status *opsv1.ClusterStatus, err error) {
	version, _ := kc.GetVersion()
	nodes, _ := kc.GetNodes()
	status = &opsv1.ClusterStatus{
		Version:       version,
		NodeNumber:    len(nodes.Items),
		LastHeartTime: &metav1.Time{Time: time.Now()},
	}
	return
}

func (kc *KubeConnection) GetVersion() (version string, err error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kc.RestConfig)
	if err != nil {
		return
	}
	info, err := discoveryClient.ServerVersion()
	if err != nil {
		return
	}
	return info.String(), err
}

func (kc *KubeConnection) GetNodes() (*corev1.NodeList, error) {
	return kc.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
}
