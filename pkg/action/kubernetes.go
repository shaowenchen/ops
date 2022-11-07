package action

import (
	"fmt"
	"time"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/shaowenchen/ops/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
)

func KubernetesScript(logger *log.Logger, option kubernetes.ScriptOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		logger.Error.Println(err)
		return err
	}
	nodes, err := utils.GetAllNodes(client)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	nodeList := []v1.Node{}
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
	for _, node := range nodeList {
		time.Sleep(time.Second * 1)
		namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsNamespace, fmt.Sprintf("script-%s", time.Now().Format("2006-01-02-15-04-05")))
		if err != nil {
			logger.Error.Println(err)
		}
		_, err = kubernetes.RunScriptOnNode(client, node, namespacedName, option.Image, option.Content)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	return
}

func KubernetesFile(logger *log.Logger, option kubernetes.FileOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		logger.Error.Println(err)
		return err
	}
	nodes, err := utils.GetAllNodes(client)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	nodeList := []v1.Node{}
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
	for _, node := range nodeList {
		time.Sleep(time.Second * 1)
		namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsNamespace, fmt.Sprintf("file-%s", time.Now().Format("2006-01-02-15-04-05")))
		if err != nil {
			logger.Error.Println(err)
		}
		_, err = kubernetes.DownloadFileOnNode(client, node, namespacedName, option.Image, option.RemoteFile, option.LocalFile)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	return
}
