package kube

import (
	"context"
	"errors"
	"fmt"
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsopt "github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeConnection struct {
	Cluster    *opsv1.Cluster
	Client     *kubernetes.Clientset
	RestConfig *rest.Config
	OpsClient  *runtimeClient.Client
}

func NewClusterConnection(c *opsv1.Cluster) (kc *KubeConnection, err error) {
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
	kc.BuildClients()
	return
}

func NewKubeConnection(kubeconfigPath string) (kc *KubeConnection, err error) {
	kc = &KubeConnection{}
	kc.RestConfig, err = utils.GetRestConfig(kubeconfigPath)
	if err != nil {
		return
	}
	kc.BuildClients()
	return
}

func (kc *KubeConnection) BuildClients() (err error) {
	kc.Client, err = utils.GetClientByRestconfig(kc.RestConfig)
	if err != nil {
		return
	}
	scheme, err := opsv1.SchemeBuilder.Build()
	if err != nil {
		return
	}

	opsClient, err := runtimeClient.New(kc.RestConfig, runtimeClient.Options{Scheme: scheme})
	if err == nil {
		kc.OpsClient = &opsClient
	}
	// try others
	return
}

func (kc *KubeConnection) GetStatus() (status *opsv1.ClusterStatus, err error) {
	version, err1 := kc.GetVersion()
	err = utils.MergeError(err, err1)
	nodes, err1 := kc.GetNodes()
	err = utils.MergeError(err, err1)
	allPods, err1 := kc.GetAllPods()
	err = utils.MergeError(err, err1)
	allRunningPods, err1 := kc.GetAllRunningPods()
	err = utils.MergeError(err, err1)

	status = &opsv1.ClusterStatus{
		Version:     version,
		Node:        len(nodes.Items),
		Pod:         len(allPods.Items),
		RunningPod:  len(allRunningPods.Items),
		HeartTime:   &metav1.Time{Time: time.Now()},
		HeartStatus: opsv1.StatusSuccessed,
	}
	return
}

func (kc *KubeConnection) GetVersion() (version string, err error) {
	info, err := kc.Client.DiscoveryClient.ServerVersion()
	if err != nil {
		return
	}
	return info.String(), err
}

func (kc *KubeConnection) GetNodes() (*corev1.NodeList, error) {
	return kc.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
}

func (kc *KubeConnection) GetNodeByName(nodeName string) (*corev1.NodeList, error) {
	nodes := &corev1.NodeList{}
	node, err := kc.Client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	nodes.Items = append(nodes.Items, *node)
	return nodes, err
}

func (kc *KubeConnection) GetAllPods() (allPod *corev1.PodList, err error) {
	return kc.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
}

func (kc *KubeConnection) GetAllRunningPods() (allPod *corev1.PodList, err error) {
	return kc.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
}

func (kc *KubeConnection) ShellOnNode(logger *opslog.Logger, node *corev1.Node, shellOpt opsopt.ShellOption, kubeOpt opsopt.KubeOption) (stdout string, err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(kc.Client, constants.OpsNamespace, fmt.Sprintf("shell-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		return
	}

	pod, err := RunShellOnNode(kc.Client, node, namespacedName, kubeOpt.RuntimeImage, shellOpt.Content)
	if err != nil {
		return
	}
	stdout, err = GetPodLog(logger, context.TODO(), false, kc.Client, pod)
	logger.Info.Println(stdout)
	return
}

func (kc *KubeConnection) Shell(logger *opslog.Logger, shellOpt opsopt.ShellOption, kubeOpt opsopt.KubeOption) (err error) {
	nodes, err := kc.GetNodeByName(kubeOpt.NodeName)

	if err != nil {
		return
	}
	if kubeOpt.All {
		nodes, err = kc.GetNodes()
	}
	for _, node := range nodes.Items {
		kc.ShellOnNode(logger, &node, shellOpt, kubeOpt)
	}

	return
}

func (kc *KubeConnection) FileonNode(logger *opslog.Logger, node *corev1.Node, option opsopt.FileOption) (stdout string, err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(kc.Client, constants.OpsNamespace, fmt.Sprintf("file-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		return
	}

	pod, err := DownloadFileOnNode(kc.Client, node, namespacedName, option.RuntimeImage, option.RemoteFile, option.LocalFile)
	if err != nil {
		return
	}
	return GetPodLog(logger, context.TODO(), false, kc.Client, pod)
}

func (kc *KubeConnection) File(logger *opslog.Logger, option opsopt.FileOption) (err error) {
	nodes, err := kc.GetNodeByName(option.NodeName)
	if option.All {
		nodes, err = kc.GetNodes()
	}
	for _, node := range nodes.Items {
		kc.FileonNode(logger, &node, option)
	}
	return

}

func (kc *KubeConnection) SetRequestLimit(logger *opslog.Logger, option opsopt.KubernetesOption) (err error) {
	namespacedName := types.NamespacedName{
		Namespace: option.Metadata.Namespace,
		Name:      option.Metadata.Name,
	}
	if option.Kind == "Deployment" {
		err = SetDeploymentRecommandResource(kc.Client, namespacedName)
		if err != nil {
			logger.Error.Println(err)
		}
		return err
	}
	if option.Kind == "StatefulSet" {
		err = SetStatefulSetRecommandResource(kc.Client, namespacedName)
		if err != nil {
			logger.Error.Println(err)
		}
		return err
	}
	if option.Kind == "DaemonSet" {
		err = SetDaemonSetRecommandResource(kc.Client, namespacedName)
		if err != nil {
			logger.Error.Println(err)
		}
		return err
	}
	logger.Info.Println("Unrecognized type: " + option.Kind)
	return
}
