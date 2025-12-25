package kube

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsoption "github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Shell(logger *opslog.Logger, client *kubernetes.Clientset, node v1.Node, shellOpt opsoption.ShellOption, kubeOpt opsoption.KubeOption) (err error) {
	logger.Info.Println("> Run shell on ", node.Name)
	namespacedName, err := utils.GetOrCreateNamespacedName(client, kubeOpt.Namespace, fmt.Sprintf("ops-shell-%s-%d", time.Now().Format("2006-01-02-15-04-05"), rand.Intn(10000)))
	if err != nil {
		logger.Error.Println(err)
	}
	pod, err := RunShellOnNode(client, &node, namespacedName, kubeOpt.RuntimeImage, shellOpt.Mode, shellOpt.Content, kubeOpt.Mounts)
	if err != nil {
		logger.Error.Println(err)
	}
	stdout, err := GetPodLog(logger, context.TODO(), kubeOpt.Debug, client, pod)
	if err != nil {
		logger.Error.Println(err)
	} else {
		logger.Info.Println(stdout)
	}
	return
}

func File(logger *opslog.Logger, client *kubernetes.Clientset, node v1.Node, fileOpt opsoption.FileOption) (stdout string, err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(client, fileOpt.Namespace, fmt.Sprintf("ops-file-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		logger.Error.Println(err)
	}
	pod := &v1.Pod{}
	pod, err = RunFileOnNode(client, &node, namespacedName, fileOpt)
	if err != nil {
		logger.Error.Println(err)
	}
	stdout, err = GetPodLog(logger, context.TODO(), fileOpt.Debug, client, pod)
	logger.Info.Println(stdout)
	return
}

func GetPodLog(logger *opslog.Logger, ctx context.Context, debug bool, client *kubernetes.Clientset, pod *v1.Pod) (logs string, err error) {
	defer func() {
		if !debug {
			client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		}
	}()
	for range time.Tick(time.Second * 2) {
		select {
		default:
			pod, err = client.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if utils.IsPendingPod(pod) {
				continue
			}
			logs, err = utils.GetPodLog(ctx, client, pod.Namespace, pod.Name)
			if len(logs) == 0 && err != nil {
				logs = err.Error()
			}
			// check node status, if not ready return success
			node, err1 := client.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
			if err1 != nil {
				logger.Error.Println(err1)
				err = err1
				return
			}
			if !utils.IsNodeReady(node) {
				err = nil
				return
			}
			if utils.IsSucceededPod(pod) {
				return
			}
			if utils.IsFailedPod(pod) {
				err = errors.New("status failed, logs: " + logs)
				return
			}
		}
	}
	return
}

func GetNodes(ctx context.Context, logger *opslog.Logger, client *kubernetes.Clientset, kubeOpt opsoption.KubeOption) (nodeList []v1.Node, err error) {
	nodes, err := utils.GetAllReadyNodesByClient(client)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if kubeOpt.IsAllNodes() {
		nodeList = nodes.Items
		return
	}

	masters := []v1.Node{}
	wokers := []v1.Node{}

	for _, node := range nodes.Items {
		if kubeOpt.NodeName == node.Name {
			nodeList = append(nodeList, node)
			return
		}
		if utils.IsMasterNode(&node) {
			masters = append(masters, node)
		} else {
			wokers = append(wokers, node)
		}
	}
	if kubeOpt.IsAllMasters() {
		nodeList = masters
		return
	} else if kubeOpt.IsAllWorkers() {
		nodeList = wokers
		return
	}
	// if allmaster
	if kubeOpt.IsAllMasters() {
		nodeList = masters
		return
	} else if kubeOpt.IsAllWorkers() {
		nodeList = wokers
		return
	}
	// random select one
	if kubeOpt.IsAnyMaster() {
		nodeList = masters
	} else if kubeOpt.IsAnyWorker() {
		nodeList = wokers
	} else if kubeOpt.IsAnyNode() {
		nodeList = nodes.Items
	}
	if len(nodeList) == 0 {
		return
	}
	randomIndex := rand.Intn(len(nodeList))
	nodeList = []v1.Node{nodeList[randomIndex]}
	return
}

func GetOpsClient(ctx context.Context, logger *opslog.Logger, restConfig *rest.Config) (client runtimeClient.Client, err error) {
	scheme, err := opsv1.SchemeBuilder.Build()
	if err != nil {
		return
	}

	client, err = runtimeClient.New(restConfig, runtimeClient.Options{Scheme: scheme})
	if err != nil {
		return
	}
	return
}
