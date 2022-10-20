package kube

import (
	"fmt"
	"strings"
	"time"

	"github.com/shaowenchen/opscli/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func ActionClear(option ClearOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	namespaces := utils.SplitStr(option.Namespace)
	if option.All {
		namespaces, err = utils.GetAllNamespaces(client)
		if err != nil {
			return
		}
	}
	for _, namespace := range namespaces {
		err = ClearPod(client, namespace, utils.SplitStr(option.Status))
		if err != nil {
			fmt.Println(err)
		}
	}
	return
}

func ActionDescheduler(option DeschedulerOption) (err error) {
	config, err := utils.GetRestConfig(option.Kubeconfig)
	if config == nil || err != nil {
		return utils.LogError(err)
	}
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	err = RunDeScheduler(config, client, option.RemoveDuplicates, option.NodeUtilization, option.HighPercent)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

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
		namespacedName, err := utils.GetOrCreateNamespacedName(client, OpsCliNamespace, fmt.Sprintf("script-%s", time.Now().Format("2006-01-02-15-04-05")))
		if err != nil {
			utils.LogError(err)
		}
		_, err = RunScriptOnNode(client, node, namespacedName, option.Content)
		if err != nil {
			utils.LogError(err)
		}
	}
	return
}

func ActionEtcHostsOnNode(option EtcHostsOption) (err error) {
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
	if option.Clear {
		namespacedName, err := utils.GetOrCreateNamespacedName(client, OpsCliNamespace, fmt.Sprintf("deleteetchosts-%s", option.Domain))
		if err != nil {
			return utils.LogError(err)
		}
		err = RunScriptOnNodes(client, nodeList, namespacedName, utils.ScriptDeleteHost(option.Domain))
	} else {
		namespacedName, err := utils.GetOrCreateNamespacedName(client, OpsCliNamespace, fmt.Sprintf("addetchosts-%s-%s", option.Domain, strings.ReplaceAll(option.IP, ".", "-")))
		if err != nil {
			return utils.LogError(err)
		}
		err = RunScriptOnNodes(client, nodeList, namespacedName, utils.ScriptAddHost(option.IP, option.Domain))
	}
	if err != nil {
		return utils.LogError(err)
	}
	return
}

func ActionImagePullSecret(option ImagePulllSecretOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	namespacedNames := []types.NamespacedName{utils.BuildNamespacedName(option.Namespace, option.Name)}
	if option.All {
		namespacedNames, err = utils.SplitAllNamespacedName(client, option.Name)
		if err != nil {
			utils.LogError(err)
		}
	}

	for _, namespacedName := range namespacedNames {
		if option.Clear {
			_, err = DeleteSecret(client, namespacedName)
		} else {
			_, err = CreateImagePullSecret(client, namespacedName, option.Host, option.Username, option.Password)
		}
		if err != nil {
			utils.LogError(err)
		}
	}
	return
}

func ActionNodeSelector(option NodeSelectorOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	namespacedNames := utils.SplitNamespacedName(option.Name)
	for _, namespacedName := range namespacedNames {
		if option.Clear {
			option.KeyLabels = ""
		}
		labePairs := utils.SplitKeyValues(option.KeyLabels)
		err = SetDeploymentNodeSelector(client, namespacedName, labePairs)
		if err != nil {
			return utils.LogError(err)
		}
	}
	return
}

func ActionNodeName(option NodeNameOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	for _, namespacedName := range utils.SplitNamespacedName(option.Name) {
		if option.Clear {
			option.NodeName = ""
		}
		err = SetDeploymentNodeName(client, option.NodeName, namespacedName)
		if err != nil {
			return utils.LogError(err)
		}
	}
	return
}

func ActionLimitRange(option LimitRangeOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	namespacedNames := utils.SplitNamespacedName(option.Name)
	if option.All {
		namespacedNames, err = utils.SplitAllNamespacedName(client, option.Name)
		if err != nil {
			utils.LogError(err)
		}
	}
	for _, namespacedName := range namespacedNames {
		if option.Clear {
			err = DeleteLimitRange(client, namespacedName)
		} else {
			err = CreateLimitRange(client, namespacedName, option.ReqMem, option.LimitMem, option.ReqCPU, option.LimitCPU)
		}
		if err != nil {
			return utils.LogError(err)
		}
	}
	return
}

func ActionAnnotate(option AnnotateOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if client == nil || err != nil {
		return utils.LogError(err)
	}
	if option.Type != "velero" {
		fmt.Println("not support this types")
		return
	}
	namespaces := []string{option.Namespace}
	if option.All {
		namespaces, err = utils.GetAllNamespaces(client)
		if err != nil {
			utils.LogError(err)
		}
	}

	for _, namespace := range namespaces {
		updatedPodNames, err := AnnotateVeleroPod(client, namespace, option.Clear)
		if len(updatedPodNames) > 0 {
			fmt.Println("updatedPodNames:\n", strings.Join(updatedPodNames, "\n"))
		}
		if err != nil {
			utils.LogError(err)
		}
	}
	return
}
