package kubernetes

import (
	"fmt"
	"time"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func Script(logger *log.Logger, client *kubernetes.Clientset, node v1.Node, option ScriptOption) (err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsNamespace, fmt.Sprintf("script-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		logger.Error.Println(err)
	}
	_, err = RunScriptOnNode(client, node, namespacedName, option.Image, option.Content)
	if err != nil {
		logger.Error.Println(err)
	}
	return
}

func File(logger *log.Logger, client *kubernetes.Clientset, node v1.Node, option FileOption) (err error) {
	namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsNamespace, fmt.Sprintf("file-%s", time.Now().Format("2006-01-02-15-04-05")))
	if err != nil {
		logger.Error.Println(err)
	}
	_, err = DownloadFileOnNode(client, node, namespacedName, option.Image, option.RemoteFile, option.LocalFile)
	if err != nil {
		logger.Error.Println(err)
	}
	return
}

func GetClientAndNodes(logger *log.Logger, option KubeOption) (client *kubernetes.Clientset, nodeList []v1.Node, err error) {
	client, err = utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		logger.Error.Println(err)
		return
	}
	nodes, err := utils.GetAllNodes(client)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if len(option.NodeName) > 0 {
		for _, node := range nodes.Items {
			if node.Name == option.NodeName {
				nodeList = append(nodeList, node)
			}
		}
	}
	if option.All {
		nodeList = nodes.Items
	}
	return
}
