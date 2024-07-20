package kube

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/log"
	option "github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Shell(logger *log.Logger, client *kubernetes.Clientset, node v1.Node, shellOpt option.ShellOption, kubeOpt option.KubeOption) (err error) {
	logger.Info.Println("> Run shell on ", node.Name)
	namespacedName, err := utils.GetOrCreateNamespacedName(client, kubeOpt.OpsNamespace, fmt.Sprintf("ops-shell-%s-%d", time.Now().Format("2006-01-02-15-04-05"), rand.Intn(10000)))
	if err != nil {
		logger.Error.Println(err)
	}
	pod, err := RunShellOnNode(client, &node, namespacedName, kubeOpt.RuntimeImage, shellOpt.Content)
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

func File(logger *log.Logger, client *kubernetes.Clientset, node v1.Node, fileOpt option.FileOption) (stdout string, err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(client, fileOpt.OpsNamespace, fmt.Sprintf("ops-file-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		logger.Error.Println(err)
	}
	pod := &v1.Pod{}
	pod, err = RunFileOnNode(client, &node, namespacedName, fileOpt)
	if err != nil {
		logger.Error.Println(err)
	}
	stdout, err = GetPodLog(logger, context.TODO(), false, client, pod)
	logger.Info.Println(stdout)
	return
}

func GetPodLog(logger *log.Logger, ctx context.Context, debug bool, client *kubernetes.Clientset, pod *v1.Pod) (logs string, err error) {
	defer func() {
		if !debug {
			client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		}
	}()
	for range time.Tick(time.Second * 1) {
		select {
		default:
			pod, err = client.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if utils.IsPendingPod(pod) {
				continue
			}
			logs, err = utils.GetPodLog(ctx, client, pod.Namespace, pod.Name)
			if err != nil {
				return
			}
			if utils.IsSucceededPod(pod) {
				return
			}
			if utils.IsFailedPod(pod) {
				if len(logs) == 0 && err != nil {
					logs = err.Error()
				}
				err = errors.New("status failed, logs: " + logs)
				return
			}
		}
	}
	return
}

func GetNodes(ctx context.Context, logger *log.Logger, client *kubernetes.Clientset, kubeOpt option.KubeOption) (nodeList []v1.Node, err error) {
	nodes, err := utils.GetAllReadyNodesByClient(client)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if kubeOpt.All {
		nodeList = nodes.Items
		return
	}
	for _, node := range nodes.Items {
		if len(kubeOpt.NodeName) == 0 && utils.IsMasterNode(&node) {
			nodeList = append(nodeList, node)
			return
		} else if kubeOpt.NodeName == constants.AnyMaster && utils.IsMasterNode(&node) {
			nodeList = append(nodeList, node)
		} else if kubeOpt.NodeName == node.Name {
			nodeList = append(nodeList, node)
		}
	}
	if len(nodeList) == 0 {
		err = errors.New("no node found")
	}
	// if anymaster, random a master to return
	if kubeOpt.NodeName == constants.AnyMaster && len(nodeList) > 1 {
		randomIndex := rand.Intn(len(nodeList))
		nodeList = []v1.Node{nodeList[randomIndex]}
	}
	return
}

func GetOpsClient(ctx context.Context, logger *log.Logger, restConfig *rest.Config) (client runtimeClient.Client, err error) {
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
