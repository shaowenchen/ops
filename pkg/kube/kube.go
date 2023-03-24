package kube

import (
	"context"
	"fmt"
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

func Shell(logger *log.Logger, client *kubernetes.Clientset, node v1.Node, shellOpt option.ShellOption, kubeOpt option.KubeOption) (stdout string, err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsNamespace, fmt.Sprintf("shell-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		logger.Error.Println(err)
	}
	pod, err := RunShellOnNode(client, &node, namespacedName, kubeOpt.RuntimeImage, shellOpt.Content)
	if err != nil {
		logger.Error.Println(err)
	}
	stdout, err = GetPodLog(logger, context.TODO(), false, client, pod)
	logger.Info.Println(stdout)
	return
}

func File(logger *log.Logger, client *kubernetes.Clientset, node v1.Node, option option.FileOption, kubeOption option.KubeOption) (stdout string, err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsNamespace, fmt.Sprintf("file-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		logger.Error.Println(err)
	}
	pod, err := DownloadFileOnNode(client, &node, namespacedName, option.StorageImage, option.RemoteFile, option.LocalFile)
	if err != nil {
		logger.Error.Println(err)
	}
	stdout, err = GetPodLog(logger, context.TODO(), false, client, pod)
	logger.Info.Println(stdout)
	return
}

func GetPodLog(logger *log.Logger, ctx context.Context, debug bool, client *kubernetes.Clientset, pod *v1.Pod) (logs string, err error) {
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
			if utils.IsStopedPod(pod) {
				if !debug {
					client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
				}
				return
			}
		}
	}
	return
}

func GetNodes(logger *log.Logger, client *kubernetes.Clientset, kubeOpt option.KubeOption) (nodeList []v1.Node, err error) {
	nodes, err := utils.GetAllReadyNodesByClient(client)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if len(kubeOpt.NodeName) > 0 {
		for _, node := range nodes.Items {
			if kubeOpt.NodeName == constants.AnyMaster && utils.IsMasterNode(&node) {
				nodeList = append(nodeList, node)
				return
			} else if kubeOpt.NodeName == node.Name {
				nodeList = append(nodeList, node)
			}
		}
	}
	if kubeOpt.All {
		nodeList = nodes.Items
	}
	return
}

func GetOpsClient(logger *log.Logger, restConfig *rest.Config) (client runtimeClient.Client, err error) {
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
