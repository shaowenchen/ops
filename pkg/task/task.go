package task

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

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
	// Map to store step outputs for path references: map[stepName]output
	stepOutputs := make(map[string]string)
	logger.Debug.Println("> Run Task", t.GetUniqueKey(), "on", hc.Host.Spec.Address)
	for si, s := range t.Spec.Steps {
		var sp = &s
		sp = RenderStepVariablesWithPathRefs(sp, allVars, nil)
		// Also support steps.{stepName}.output references
		sp = RenderStepVariablesWithStepRefs(sp, allVars, stepOutputs)
		logger.Debug.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderStringWithStepRefs(s.When, allVars, stepOutputs)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if !result {
			logger.Debug.Println("Skip!")
			continue
		}
		if err != nil {
			logger.Error.Println(err)
		}
		stepFunc := GetHostStepFunc(s)
		stepStatus, stepOutput, stepErr := stepFunc(t, hc, s, taskOpt)
		stepStatus = GetValidStatusError(stepStatus, stepErr)
		tr.Status.AddOutputStep(hc.Host.Name, s.Name, s.Content, stepOutput, stepStatus)
		// Store step output for path references
		stepOutputs[s.Name] = strings.ReplaceAll(stepOutput, "\"", "")
		allVars["result"] = strings.ReplaceAll(stepOutput, "\"", "")
		allVars["output"] = strings.ReplaceAll(stepOutput, "\"", "")
		allVars["status"] = stepStatus
		logger.Debug.Println(stepOutput)
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
	// Map to store step outputs for path references: map[stepName]output
	stepOutputs := make(map[string]string)
	logger.Debug.Println("> Run Task", t.GetUniqueKey(), "on Node", node.Name)

	// Collect all steps that need to be executed
	stepsToExecute := []opsv1.Step{}
	for si, s := range t.Spec.Steps {
		var sp = &s
		sp = RenderStepVariablesWithPathRefs(sp, allVars, nil)
		// Also support steps.{stepName}.output references
		sp = RenderStepVariablesWithStepRefs(sp, allVars, stepOutputs)
		logger.Debug.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderStringWithStepRefs(s.When, allVars, stepOutputs)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		if !result {
			logger.Debug.Println("Skip!")
			continue
		}
		stepsToExecute = append(stepsToExecute, s)
	}

	if len(stepsToExecute) == 0 {
		logger.Debug.Println("No steps to execute")
		return nil
	}

	// Determine the node to use (master if kubectl is needed)
	execNode := node
	for _, s := range stepsToExecute {
		if strings.Contains(s.Content, "kubectl") && !kc.IsMaster(node) {
			execNode, err = kc.GetAnyMaster()
			if err != nil {
				logger.Error.Println(err)
				return err
			}
			break
		}
	}

	// Build step container configurations
	stepConfigs := []kube.StepContainerConfig{}
	for _, s := range stepsToExecute {
		stepConfig := kube.StepContainerConfig{
			StepName:     s.Name,
			Content:      s.Content,
			LocalFile:    s.LocalFile,
			RemoteFile:   s.RemoteFile,
			Direction:    s.Direction,
			RuntimeImage: s.RuntimeImage,
			AllowFailure: s.AllowFailure,
		}

		// Determine mode and if it's a file step
		if len(s.Content) > 0 {
			// Shell step
			// Use container mode if mounts are configured
			if len(kubeOpt.Mounts) > 0 {
				stepConfig.Mode = opsconstants.ModeContainer
			} else {
				stepConfig.Mode = opsconstants.ModeHost
			}
			stepConfig.IsFileStep = false
		} else {
			// File step
			stepConfig.IsFileStep = true
			fileOpt := option.FileOption{
				Sudo:       taskOpt.Sudo,
				Direction:  s.Direction,
				LocalFile:  s.LocalFile,
				RemoteFile: s.RemoteFile,
				Api:        taskOpt.Variables["api"],
				AesKey:     taskOpt.Variables["aeskey"],
				AK:         taskOpt.Variables["ak"],
				SK:         taskOpt.Variables["sk"],
				Region:     taskOpt.Variables["region"],
				Endpoint:   taskOpt.Variables["endpoint"],
				Bucket:     taskOpt.Variables["bucket"],
				KubeOption: kubeOpt,
			}
			if s.RuntimeImage != "" {
				fileOpt.RuntimeImage = s.RuntimeImage
			} else {
				fileOpt.RuntimeImage = kubeOpt.RuntimeImage
			}
			stepConfig.FileOpt = &fileOpt
		}

		stepConfigs = append(stepConfigs, stepConfig)
	}

	// Create pod with multiple containers (one per step)
	namespacedName, err := utils.GetOrCreateNamespacedName(kc.Client, kubeOpt.Namespace, fmt.Sprintf("ops-task-%s-%d", time.Now().Format("2006-01-02-15-04-05"), rand.Intn(10000)))
	if err != nil {
		logger.Error.Println(err)
		return err
	}

	pod, err := kc.RunTaskStepsOnNode(execNode, namespacedName, stepConfigs, kubeOpt.RuntimeImage, kubeOpt.Mounts)
	if err != nil {
		logger.Error.Println(err)
		return err
	}

	// Wait for pod to complete and collect logs from each container
	err = kc.WaitForTaskStepsPod(logger, pod, stepConfigs, tr, node.Name, allVars, stepOutputs)
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
	// Use container mode if mounts are configured
	mode := opsconstants.ModeHost
	if len(kubeOpt.Mounts) > 0 {
		mode = opsconstants.ModeContainer
	}
	// Use step-level runtimeImage if specified, otherwise use kubeOpt.RuntimeImage
	stepKubeOpt := kubeOpt
	if step.RuntimeImage != "" {
		stepKubeOpt.RuntimeImage = step.RuntimeImage
	}
	output, err = kc.ShellOnNode(
		logger,
		node,
		option.ShellOption{
			Sudo:    taksOpt.Sudo,
			Content: step.Content,
			Mode:    mode,
		},
		stepKubeOpt)
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
	// Use step-level runtimeImage if specified, otherwise use kubeOpt.RuntimeImage
	stepKubeOpt := kubeOpt
	if step.RuntimeImage != "" {
		stepKubeOpt.RuntimeImage = step.RuntimeImage
	}
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
		KubeOption: stepKubeOpt,
	}
	output, err = kc.FileNode(
		logger,
		node,
		fileOpt,
	)
	return
}
