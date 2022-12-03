package task

import (
	"fmt"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func RunTaskOnHost(logger *opslog.Logger, t *opsv1.Task, hc *host.HostConnection, taskOpt option.TaskOption) error {
	allVars, err := GetRealVariables(t, taskOpt)
	if err != nil {
		return err
	}
	for si, s := range t.Spec.Steps {
		var sp = &s
		logger.Info.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderWhen(s.When, allVars)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if !result {
			logger.Error.Println("Skip!")
			continue
		}
		sp = RenderStepVariables(sp, allVars)
		if err != nil {
			logger.Error.Println(err)
		}
		if taskOpt.Debug && len(s.Content) > 0 {
			logger.Error.Println(s.Content)
		}
		stepFunc := GetHostStepFunc(s)
		stepOutput, stepErr := stepFunc(t, hc, s, taskOpt)
		t.Status.AddOutputStep(hc.Host.Name, s.Name, s.Content, stepOutput, opsv1.GetRunStatus(stepErr))
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
	for si, s := range t.Spec.Steps {
		var sp = &s
		logger.Info.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderWhen(s.When, allVars)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if !result {
			logger.Error.Println("Skip!")
			continue
		}
		sp = RenderStepVariables(sp, allVars)
		if err != nil {
			logger.Error.Println(err)
		}
		if taskOpt.Debug && len(s.Content) > 0 {
			logger.Info.Println(s.Content)
		}
		stepFunc := GetKubeStepFunc(s)
		stepOutput, stepErr := stepFunc(logger, t, kc, node, s, taskOpt, kubeOpt)
		t.Status.AddOutputStep(node.Name, s.Name, s.Content, stepOutput, opsv1.GetRunStatus(stepErr))
		allVars["result"] = stepOutput
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

func GetHostStepFunc(step opsv1.Step) func(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, to option.TaskOption) (string, error) {
	if len(step.Content) > 0 {
		return runStepShellOnHost
	}
	return runStepCopyOnHost
}

func runStepShellOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (stdout string, err error) {
	stdout, err = c.Shell(option.Sudo, step.Content)
	return
}

func runStepCopyOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (result string, err error) {
	err = c.File(option.Sudo, step.Direction, step.LocalFile, step.RemoteFile)

	return
}

func GetKubeStepFunc(step opsv1.Step) func(logger *opslog.Logger, t *opsv1.Task, c *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (string, error) {
	if len(step.Content) > 0 {
		return runStepShellOnKube
	}
	return runStepCopyOnKube
}

func runStepShellOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taksOpt option.TaskOption, kubeOpt option.KubeOption) (result string, err error) {
	stdout, err := kc.ShellOnNode(
		logger,
		node,
		option.ShellOption{
			Sudo:       taksOpt.Sudo,
			Content:    step.Content,
			KubeOption: kubeOpt,
		})
	return stdout, err
}

func runStepCopyOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (result string, err error) {
	stdout, err := kc.FileonNode(
		logger,
		node,
		option.FileOption{
			Sudo:       taskOpt.Sudo,
			Direction:  step.Direction,
			LocalFile:  step.LocalFile,
			RemoteFile: step.RemoteFile,
		})
	return stdout, err
}
