package task

import (
	"fmt"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/prom"
	"github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func GetValidStatusError(status string, err error) string {
	if err != nil {
		return opsv1.StatusFailed
	}
	if status == "" {
		return opsv1.StatusSuccessed
	}
	return status
}

func RunTaskOnHost(logger *opslog.Logger, t *opsv1.Task, hc *host.HostConnection, taskOpt option.TaskOption) error {
	allVars, err := GetRealVariables(t, taskOpt)
	if err != nil {
		return err
	}
	logger.Info.Println("> Run Task ", t.GetUniqueKey(), " on ", hc.Host.Spec.Address)
	for si, s := range t.Spec.Steps {
		var sp = &s
		sp = RenderStepVariables(sp, allVars)
		logger.Info.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderString(s.When, allVars)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if !result {
			logger.Info.Println("Skip!")
			continue
		}
		if err != nil {
			logger.Error.Println(err)
		}
		stepFunc := GetHostStepFunc(s)
		stepStatus, stepOutput, stepErr := stepFunc(t, hc, s, taskOpt)
		stepStatus = GetValidStatusError(stepStatus, stepErr)
		t.Status.AddOutputStep(hc.Host.Name, s.Name, s.Content, stepOutput, stepStatus)
		allVars["result"] = strings.ReplaceAll(stepOutput, "\"", "")
		allVars["status"] = stepStatus
		logger.Debug.Println("Content: ", s.Content)
		logger.Debug.Println("Status: ", stepStatus)
		logger.Info.Println(stepOutput)
		result, err = utils.LogicExpression(s.AllowFailure, false)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if result == false && stepErr != nil {
			break
		}
	}
	return err
}

func RunTaskOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, taskOpt option.TaskOption, kubeOpt option.KubeOption) error {
	allVars, err := GetRealVariables(t, taskOpt)
	if err != nil {
		return err
	}
	logger.Info.Println("> Run Task ", t.GetUniqueKey(), " on Node ", node.Name)
	for si, s := range t.Spec.Steps {
		var sp = &s
		sp = RenderStepVariables(sp, allVars)
		logger.Info.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderString(s.When, allVars)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if !result {
			logger.Info.Println("Skip!")
			continue
		}
		if err != nil {
			logger.Error.Println(err)
		}
		stepFunc := GetKubeStepFunc(s)
		stepStatus, stepOutput, stepErr := stepFunc(logger, t, kc, node, s, taskOpt, kubeOpt)
		stepStatus = GetValidStatusError(stepStatus, stepErr)
		t.Status.AddOutputStep(node.Name, s.Name, s.Content, stepOutput, stepStatus)
		allVars["result"] = strings.ReplaceAll(stepOutput, "\"", "")
		allVars["status"] = stepStatus
		logger.Debug.Println("Content: ", s.Content)
		logger.Debug.Println("Status: ", stepStatus)
		logger.Info.Println(stepOutput)
		result, err = utils.LogicExpression(s.AllowFailure, false)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if result == false && stepErr != nil {
			break
		}
	}
	return err
}

func GetHostStepFunc(step opsv1.Step) func(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, to option.TaskOption) (status string, output string, err error) {
	if len(step.Alert.Url) > 0 {
		return runStepAlertOnHost
	} else if len(step.Content) > 0 {
		return runStepShellOnHost
	}
	return runStepCopyOnHost
}

func runStepShellOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (status, stdout string, err error) {
	stdout, err = c.Shell(option.Sudo, step.Content)
	return
}

func runStepCopyOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (status, output string, err error) {
	err = c.File(option.Sudo, step.Direction, step.LocalFile, step.RemoteFile)
	return
}

func GetKubeStepFunc(step opsv1.Step) func(logger *opslog.Logger, t *opsv1.Task, c *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (string, string, error) {
	if len(step.Kubernetes.Action) > 0 {
		return runStepKubernetesOnKube
	} else if len(step.Content) > 0 {
		return runStepShellOnKube
	} else {
		return runStepCopyOnKube
	}
}

func runStepAlertOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (status string, output string, err error) {
	return prom.AlertPromQuery(step.Alert.Url, step.Alert.If)
}

func runStepKubernetesOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taksOpt option.TaskOption, kubeOpt option.KubeOption) (status, output string, err error) {
	option := option.KubernetesOption{
		Kind:   step.Kubernetes.Kind,
		Action: step.Kubernetes.Action,
	}
	option.Metadata.Name = step.Kubernetes.Name
	option.Metadata.Namespace = step.Kubernetes.Namespace
	err = kc.SetRequestLimit(
		logger,
		option)
	return
}

func runStepShellOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taksOpt option.TaskOption, kubeOpt option.KubeOption) (status, output string, err error) {
	output, err = kc.ShellOnNode(
		logger,
		node,
		option.ShellOption{
			Sudo:    taksOpt.Sudo,
			Content: step.Content,
		},
		kubeOpt)
	return
}

func runStepCopyOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (status, output string, err error) {
	output, err = kc.FileonNode(
		logger,
		node,
		option.FileOption{
			Sudo:       taskOpt.Sudo,
			Direction:  step.Direction,
			LocalFile:  step.LocalFile,
			RemoteFile: step.RemoteFile,
		})
	return
}
