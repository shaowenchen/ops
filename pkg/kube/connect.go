package kube

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsopt "github.com/shaowenchen/ops/pkg/option"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	if c.IsCurrentCluster() {
		kc.RestConfig, err = opsutils.GetInClusterConfig()
		if err != nil {
			kc.RestConfig, err = opsutils.GetRestConfig(opsconstants.GetCurrentUserKubeConfigPath())
			if err != nil {
				return
			}
		}
		err = kc.BuildClients()
		return
	}
	// try config
	config, err := opsutils.DecodingBase64ToString(c.Spec.Config)
	if err != nil {
		return
	}
	kc.RestConfig, err = opsutils.GetRestConfigByContent(config)
	if err != nil {
		return
	}
	err = kc.BuildClients()
	return
}

func NewKubeConnection(kubeconfigPath string) (kc *KubeConnection, err error) {
	kc = &KubeConnection{}
	kc.RestConfig, err = opsutils.GetRestConfig(kubeconfigPath)
	if err != nil {
		kc.RestConfig, err = opsutils.GetInClusterConfig()
	}
	if err != nil {
		return
	}
	err = kc.BuildClients()
	return
}

func (kc *KubeConnection) IsMaster(node *corev1.Node) bool {
	return opsutils.IsMasterNode(node)
}

func (kc *KubeConnection) GetAnyMaster() (node *corev1.Node, err error) {
	return opsutils.GetAnyMaster(kc.Client)
}

func (kc *KubeConnection) SyncTasks(isDeleted bool, objs []opsv1.Task) (err error) {
	if kc == nil {
		return errors.New("synctasks kube connection is nil")
	}
	if kc.OpsClient == nil {
		return errors.New("synctasks ops client is nil")
	}
	for _, t := range objs {
		copyObj := (&t).CopyWithOutVersion()
		if isDeleted {
			err = (*kc.OpsClient).Delete(context.TODO(), copyObj)
			if err != nil {
				return
			}
		} else {
			err = (*kc.OpsClient).Create(context.TODO(), copyObj)
			if k8serrors.IsAlreadyExists(err) {
				otherObj := &opsv1.Task{}
				err = (*kc.OpsClient).Get(context.TODO(), types.NamespacedName{Name: copyObj.Name, Namespace: copyObj.Namespace}, otherObj)
				if err == nil {
					(*kc.OpsClient).Update(context.TODO(), copyObj.MergeVersion(otherObj))
				}
			}
		}
	}
	return
}

func (kc *KubeConnection) SyncPipelines(isDeleted bool, objs []opsv1.Pipeline) (err error) {
	for _, t := range objs {
		copyObj := (&t).CopyWithOutVersion()
		if isDeleted {
			err = (*kc.OpsClient).Delete(context.TODO(), copyObj)
			if err != nil {
				return
			}
		} else {
			err = (*kc.OpsClient).Create(context.TODO(), copyObj)
			if k8serrors.IsAlreadyExists(err) {
				otherObj := &opsv1.Pipeline{}
				err = (*kc.OpsClient).Get(context.TODO(), types.NamespacedName{Name: copyObj.Name, Namespace: copyObj.Namespace}, otherObj)
				if err == nil {
					(*kc.OpsClient).Update(context.TODO(), copyObj.MergeVersion(otherObj))
				}
			}
		}
	}
	return
}

func (kc *KubeConnection) CreatePipelineRun(pr *opsv1.PipelineRun) (err error) {
	existingPR := &opsv1.PipelineRun{}
	err = (*kc.OpsClient).Get(context.TODO(), types.NamespacedName{Name: pr.Name, Namespace: pr.Namespace}, existingPR)
	if err == nil {
		return nil
	}
	return (*kc.OpsClient).Create(context.TODO(), pr.CopyWithOutVersion())
}

func (kc *KubeConnection) GetPipelineRun(pr *opsv1.PipelineRun) (err error) {
	return (*kc.OpsClient).Get(context.TODO(), types.NamespacedName{Name: pr.Name, Namespace: pr.Namespace}, pr)
}

func (kc *KubeConnection) GetHost(namespace, hostname string) (host *opsv1.Host, err error) {
	hostList := &opsv1.HostList{}
	err = (*kc.OpsClient).List(context.TODO(), hostList, runtimeClient.InNamespace(namespace))
	if err != nil {
		return
	}
	for i := range hostList.Items {
		if hostList.Items[i].Name == hostname || hostList.Items[i].Status.Hostname == hostname {
			host = &hostList.Items[i]
			return
		}
	}
	return
}

func (kc *KubeConnection) BuildClients() (err error) {
	// Set timeout for RestConfig to prevent hanging on unreachable clusters
	if kc.RestConfig != nil && kc.RestConfig.Timeout == 0 {
		kc.RestConfig.Timeout = 30 * time.Second
	}

	kc.Client, err = opsutils.GetClientByRestconfig(kc.RestConfig)
	if err != nil {
		return
	}
	scheme, err := opsv1.SchemeBuilder.Build()
	if err != nil {
		return
	}
	err = corev1.AddToScheme(scheme)
	if err != nil {
		return
	}

	// Use recover to prevent panic from discovery client timeout
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while building clients (likely due to cluster connection timeout): %v", r)
		}
	}()

	opsClient, err := runtimeClient.New(kc.RestConfig, runtimeClient.Options{Scheme: scheme})
	if err == nil {
		kc.OpsClient = &opsClient
	}
	// try others
	return
}

func (kc *KubeConnection) GetUID() (uid string, err error) {
	return opsutils.GetClusterUID(*kc.OpsClient)
}

func (kc *KubeConnection) GetStatus() (status *opsv1.ClusterStatus, err error) {
	anyOneIsOk := false
	version, err1 := kc.GetVersion()
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	nodes, err1 := kc.GetNodes()
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	allPods, err1 := kc.GetAllPods()
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	allRunningPods, err1 := kc.GetAllRunningPods()
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	days, err1 := kc.GetExpiredDays()
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	uid, err1 := kc.GetUID()
	if err1 == nil {
		anyOneIsOk = true
	} else {
		err = err1
	}

	status = &opsv1.ClusterStatus{
		Version:          version,
		Node:             len(nodes.Items),
		Pod:              len(allPods.Items),
		RunningPod:       len(allRunningPods.Items),
		HeartTime:        &metav1.Time{Time: time.Now()},
		HeartStatus:      opsconstants.StatusSuccessed,
		CertNotAfterDays: days,
		UID:              uid,
	}

	if !anyOneIsOk {
		status.HeartStatus = opsconstants.StatusFailed
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

func (kc *KubeConnection) GetExpiredDays() (days int, err error) {
	return opsutils.GetCertNotAfterDays(kc.RestConfig)
}

func (kc *KubeConnection) ShellOnNode(logger *opslog.Logger, node *corev1.Node, shellOpt opsopt.ShellOption, kubeOpt opsopt.KubeOption) (stdout string, err error) {
	namespacedName, err := opsutils.GetOrCreateNamespacedName(kc.Client, kubeOpt.Namespace, fmt.Sprintf("ops-shell-%s-%d", time.Now().Format("2006-01-02-15-04-05"), rand.Intn(10000)))
	if err != nil {
		return
	}

	pod, err := RunShellOnNode(kc.Client, node, namespacedName, kubeOpt.RuntimeImage, shellOpt.Mode, shellOpt.Content)
	if err != nil {
		return
	}
	stdout, err = GetPodLog(logger, context.TODO(), kubeOpt.Debug, kc.Client, pod)
	return
}

func (kc *KubeConnection) Shell(logger *opslog.Logger, shellOpt opsopt.ShellOption, kubeOpt opsopt.KubeOption) (err error) {
	nodes, err := kc.GetNodeByName(kubeOpt.NodeName)

	if err != nil {
		return
	}
	if kubeOpt.IsAllNodes() {
		nodes, err = kc.GetNodes()
	}
	for _, node := range nodes.Items {
		kc.ShellOnNode(logger, &node, shellOpt, kubeOpt)
	}

	return
}

func (kc *KubeConnection) FileNode(logger *opslog.Logger, node *corev1.Node, fileOpt opsopt.FileOption) (stdout string, err error) {
	namespacedName, err := opsutils.GetOrCreateNamespacedName(kc.Client, opsconstants.OpsNamespace, fmt.Sprintf("ops-file-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		return
	}
	pod := &corev1.Pod{}
	if fileOpt.GetStorageType() == opsconstants.RemoteStorageTypeS3 {
		if fileOpt.IsUploadDirection() {
			pod, err = RunFileOnNode(kc.Client, node, namespacedName, fileOpt)
			if err != nil {
				return
			}
		}
	}
	return GetPodLog(logger, context.TODO(), false, kc.Client, pod)
}

func (kc *KubeConnection) FileNodes(logger *opslog.Logger, runtimeImage string, fileOpt opsopt.FileOption) (err error) {
	nodes, err := kc.GetNodeByName(fileOpt.NodeName)
	if fileOpt.IsAllNodes() {
		nodes, err = kc.GetNodes()
	}
	for _, node := range nodes.Items {
		kc.FileNode(logger, &node, fileOpt)
	}
	return
}
