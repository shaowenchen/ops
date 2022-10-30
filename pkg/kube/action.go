package kube

import (
	"fmt"
	"time"

	"github.com/shaowenchen/opscli/pkg/constants"
	"github.com/shaowenchen/opscli/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

func ActionScript(option ScriptOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	nodes, err := utils.GetAllNodes(client)
	if err != nil {
		return utils.LogError(err)
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
		namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsCliNamespace, fmt.Sprintf("script-%s", time.Now().Format("2006-01-02-15-04-05")))
		if err != nil {
			utils.LogError(err)
		}
		_, err = RunScriptOnNode(client, node, namespacedName, option.Image, option.Content)
		if err != nil {
			utils.LogError(err)
		}
	}
	return
}

func ActionFile(option FileOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	nodes, err := utils.GetAllNodes(client)
	if err != nil {
		return utils.LogError(err)
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
		namespacedName, err := utils.GetOrCreateNamespacedName(client, constants.OpsCliNamespace, fmt.Sprintf("file-%s", time.Now().Format("2006-01-02-15-04-05")))
		if err != nil {
			utils.LogError(err)
		}
		_, err = DownloadFileOnNode(client, node, namespacedName, option.Image, option.RemoteFile, option.LocalFile)
		if err != nil {
			utils.LogError(err)
		}
	}
	return
}
