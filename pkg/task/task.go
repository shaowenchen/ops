package task

import (
	"context"
	"fmt"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func GetValidStatusError(status string, err error) string {
	if err != nil {
		return opsconstants.StatusFailed
	}
	if status == "" {
		return opsconstants.StatusSuccessed
	}
	return status
}

func RunTaskOnHost(ctx context.Context, logger *opslog.Logger, t *opsv1.Task, tr *opsv1.TaskRun, hc *host.HostConnection, taskOpt option.TaskOption) error {
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
		tr.Status.AddOutputStep(hc.Host.Name, s.Name, s.Content, stepOutput, stepStatus)
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

func RunTaskOnKube(logger *opslog.Logger, t *opsv1.Task, tr *opsv1.TaskRun, kc *kube.KubeConnection, node *corev1.Node, taskOpt option.TaskOption, kubeOpt option.KubeOption) error {
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
		tr.Status.AddOutputStep(node.Name, s.Name, s.Content, stepOutput, stepStatus)
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
	if len(step.Content) > 0 {
		return runStepShellOnHost
	}
	return runStepFileOnHost
}

func runStepShellOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (status, stdout string, err error) {
	stdout, err = c.Shell(context.TODO(), option.Sudo, step.Content)
	return
}

func runStepFileOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, taskOpt option.TaskOption) (status, output string, err error) {
	fileOpt := option.FileOption{
		Sudo:       taskOpt.Sudo,
		Direction:  step.Direction,
		LocalFile:  step.LocalFile,
		RemoteFile: step.RemoteFile,
		Api:        taskOpt.Variables["api"],
		AesKey:     taskOpt.Variables["aeskey"],
		Region:     taskOpt.Variables["region"],
		Endpoint:   taskOpt.Variables["endpoint"],
		Bucket:     taskOpt.Variables["bucket"],
		AK:         taskOpt.Variables["ak"],
		SK:         taskOpt.Variables["sk"],
	}
	output, err = c.File(context.Background(), fileOpt)
	return
}

func GetKubeStepFunc(step opsv1.Step) func(logger *opslog.Logger, t *opsv1.Task, c *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (string, string, error) {
	if len(step.Content) > 0 {
		return runStepShellOnKube
	} else {
		return runStepFileOnKube
	}
}

func runStepShellOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taksOpt option.TaskOption, kubeOpt option.KubeOption) (status, output string, err error) {
	mode := opsconstants.ModeHost
	if strings.Contains(step.Content, "/host") {
		mode = opsconstants.ModeContainer
	}
	output, err = kc.ShellOnNode(
		logger,
		node,
		option.ShellOption{
			Sudo:    taksOpt.Sudo,
			Content: step.Content,
			Mode:    mode,
		},
		kubeOpt)
	if len(output) == 0 {
		if err != nil {
			output = err.Error()
		} else {
			output = opsconstants.NoOutput
		}
	}
	return
}

func runStepFileOnKube(logger *opslog.Logger, t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (status, output string, err error) {
	fileOpt := option.FileOption{
		Sudo:       taskOpt.Sudo,
		Direction:  step.Direction,
		LocalFile:  step.LocalFile,
		RemoteFile: step.RemoteFile,
		Api:        taskOpt.Variables["api"],
		AesKey:     taskOpt.Variables["aeskey"],
		AK:         taskOpt.Variables["ak"],
		SK:         taskOpt.Variables["sk"],
		Region:     taskOpt.Variables["region"],
		Endpoint:   taskOpt.Variables["endpoint"],
		Bucket:     taskOpt.Variables["bucket"],
		KubeOption: kubeOpt,
	}
	output, err = kc.FileNode(
		logger,
		node,
		fileOpt,
	)
	return
}
