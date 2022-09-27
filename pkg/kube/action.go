package kube

import (
	"fmt"
	"strings"
	"time"

	"github.com/shaowenchen/opscli/pkg/script"
	v1 "k8s.io/api/core/v1"
)

func ActionClear(option ClearOption) (err error) {
	client, err := NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}
	namespaces := SplitStr(option.Namespace)
	if option.All {
		namespaces, err = GetAllNamespaces(client)
		if err != nil {
			return
		}
	}
	for _, namespace := range namespaces {
		err = ClearPod(client, namespace, SplitStr(option.Status))
		if err != nil {
			fmt.Println(err)
		}
	}
	return
}

func ActionDescheduler(option DeschedulerOption) (err error) {
	config, err := GetRestConfig(option.Kubeconfig)
	if config == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}
	client, err := NewKubernetesClient(option.Kubeconfig)
	err = RunDeScheduler(config, client, option.RemoveDuplicates, option.NodeUtilization, option.HighPercent)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func ActionHostRun(option HostRunOption) (err error) {
	client, err := NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}
	nodes, err := GetAllNodes(client)
	if err != nil {
		return PrintError(ErrorMsgGetClient(err))
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
		namespacedName, err := GetOpscliNamespacedName(client, fmt.Sprintf("script-runhost-%s", time.Now().Format("2006-01-02-15-04-05")))
		if err != nil {
			PrintError(ErrorMsgRunScriptOnNode(err))
		}
		_, err = RunScriptOnNode(client, node, namespacedName, option.Script)
		if err != nil {
			PrintError(ErrorMsgRunScriptOnNode(err))
		}
	}
	return
}

func ActionEtcHostsOnNode(option EtcHostsOption) (err error) {
	client, err := NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}
	nodes, err := GetAllNodes(client)
	if err != nil {
		return PrintError(ErrorMsgGetClient(err))
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
	if option.Clear {
		namespacedName, err := GetOpscliNamespacedName(client, fmt.Sprintf("deleteetchosts-%s", option.Domain))
		if err != nil {
			return PrintError(ErrorMsgRunScriptOnNode(err))
		}
		err = RunScriptOnNodes(client, nodeList, namespacedName, script.DeleteHost(option.Domain))
	} else {
		namespacedName, err := GetOpscliNamespacedName(client, fmt.Sprintf("addetchosts-%s-%s", option.Domain, strings.ReplaceAll(option.IP, ".", "-")))
		if err != nil {
			return PrintError(ErrorMsgRunScriptOnNode(err))
		}
		err = RunScriptOnNodes(client, nodeList, namespacedName, script.AddHost(option.IP, option.Domain))
	}
	if err != nil {
		return PrintError(ErrorMsgRunScriptOnNode(err))
	}
	return
}

func ActionImagePullSecret(option ImagePulllSecretOption) (err error) {
	client, err := NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}

	namespacedNames := SplitNamespacedName(option.Name)
	if option.All {
		namespacedNames, err = SplitAllNamespacedName(client, option.Name)
		if err != nil {
			PrintError(err.Error())
		}
	}

	for _, namespacedName := range namespacedNames {
		if option.Clear {
			_, err = DeleteSecret(client, namespacedName)
		} else {
			_, err = CreateImagePullSecret(client, namespacedName, option.Host, option.Username, option.Password)
		}
		if err != nil {
			PrintError(ErrorMsgImagePullSecret(err))
		}
	}
	return
}

func ActionNodeSelector(option NodeSelectorOption) (err error) {
	client, err := NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}
	namespacedNames := SplitNamespacedName(option.Name)
	for _, namespacedName := range namespacedNames {
		if option.Clear {
			option.KeyLabels = ""
		}
		labePairs := SplitKeyValues(option.KeyLabels)
		err = SetDeploymentNodeSelector(client, namespacedName, labePairs)
		if err != nil {
			return PrintError(ErrorMsgNodeSelector(err))
		}
	}
	return
}

func ActionNodeName(option NodeNameOption) (err error) {
	client, err := NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}
	for _, namespacedName := range SplitNamespacedName(option.Name) {
		if option.Clear {
			option.NodeName = ""
		}
		err = SetDeploymentNodeName(client, option.NodeName, namespacedName)
		if err != nil {
			return PrintError(ErrorMsgNodeName(err))
		}
	}
	return
}

func ActionLimitRange(option LimitRangeOption) (err error) {
	client, err := NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return PrintError(ErrorMsgGetClient(err))
	}
	namespacedNames := SplitNamespacedName(option.Name)
	if option.All {
		namespacedNames, err = SplitAllNamespacedName(client, option.Name)
		if err != nil {
			PrintError(err.Error())
		}
	}
	for _, namespacedName := range namespacedNames {
		if option.Clear {
			err = DeleteLimitRange(client, namespacedName)
		} else {
			err = CreateLimitRange(client, namespacedName, option.ReqMem, option.LimitMem, option.ReqCPU, option.LimitCPU)
		}
		if err != nil {
			return PrintError(ErrorMsgLimitRange(err))
		}
	}
	return
}
