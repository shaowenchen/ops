package task

import (
	"fmt"

	"errors"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"os"
	"path/filepath"
	"strings"
)

func GetRealVariables(t *opsv1.Task, option option.TaskOption) (map[string]string, error) {
	globalVariables := make(map[string]string)
	// cli > env > yaml
	utils.MergeMap(globalVariables, t.Spec.Variables)
	utils.MergeMap(globalVariables, utils.GetAllOsEnv())
	utils.MergeMap(globalVariables, option.Variables)

	globalVariables = RenderVarsVariables(globalVariables)
	// check variable in task is not empty
	emptyVariable := ""
	for key := range t.Spec.Variables {
		if len(strings.TrimSpace(globalVariables[key])) == 0 {
			emptyVariable = key
			break
		}
	}
	if len(emptyVariable) > 0 {
		return nil, errors.New("please set variable: " + emptyVariable)
	}
	return globalVariables, nil
}

func RenderTask(t *opsv1.Task, allVars map[string]string) (*opsv1.Task, error) {
	for i, s := range t.Spec.Steps {
		sp := RenderStepVariables(&s, allVars)
		t.Spec.Steps[i] = *sp
	}
	return t, nil
}

func RunTaskOnHost(t *opsv1.Task, hc *host.HostConnection, taskOpt option.TaskOption) (*opsv1.Task, error) {
	allVars, err := GetRealVariables(t, taskOpt)
	for si, s := range t.Spec.Steps {
		var sp = &s
		fmt.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderWhen(s.When, allVars)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			fmt.Println(err)
			return t, err
		}
		if !result {
			fmt.Println("Skip!")
			continue
		}
		sp = RenderStepVariables(sp, allVars)
		if err != nil {
			fmt.Println(err)
		}
		if taskOpt.Debug && len(s.Script) > 0 {
			fmt.Println(s.Script)
		}
		stepFunc := GetHostStepFunc(s)
		stepResult, isSuccessed := stepFunc(t, hc, s, taskOpt)
		t.Status.AddOutputStep(hc.Host.Name, s.Name, stepResult, isSuccessed)
		fmt.Println(stepResult)
		allVars["result"] = stepResult
		result, err = utils.LogicExpression(s.AllowFailure, false)
		if err != nil {
			fmt.Println(err)
			return t, err
		}
		if result == false && isSuccessed == false {
			break
		}
	}
	return t, err
}

func RunTaskOnKube(t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, taskOpt option.TaskOption, kubeOpt option.KubeOption) (*opsv1.Task, error) {
	allVars, err := GetRealVariables(t, taskOpt)
	for si, s := range t.Spec.Steps {
		var sp = &s
		fmt.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderWhen(s.When, allVars)
		result, err := utils.LogicExpression(s.When, true)
		if err != nil {
			fmt.Println(err)
			return t, err
		}
		if !result {
			fmt.Println("Skip!")
			continue
		}
		sp = RenderStepVariables(sp, allVars)
		if err != nil {
			fmt.Println(err)
		}
		if taskOpt.Debug && len(s.Script) > 0 {
			fmt.Println(s.Script)
		}
		stepFunc := GetKubeStepFunc(s)
		stepResult, isSuccessed := stepFunc(t, kc, node, s, taskOpt, kubeOpt)
		t.Status.AddOutputStep(node.Name, s.Name, stepResult, isSuccessed)
		fmt.Println(stepResult)
		allVars["result"] = stepResult
		result, err = utils.LogicExpression(s.AllowFailure, false)
		if err != nil {
			fmt.Println(err)
			return t, err
		}
		if result == false && isSuccessed == false {
			break
		}
	}
	return t, err
}

func fillHostByOption(h *opsv1.Host, option *option.HostOption) *opsv1.Host {
	if option.Username != "" && h.GetSpec().Username == "" {
		h.Spec.Username = option.Username
	}
	if option.Password != "" && h.GetSpec().Password == "" {
		h.Spec.Password = option.Password
	}
	if option.Port != 0 && h.GetSpec().Port == 0 {
		h.Spec.Port = option.Port
	}
	if option.PrivateKey != "" && h.GetSpec().PrivateKey == "" {
		h.Spec.PrivateKey = option.PrivateKey
	}
	if option.PrivateKeyPath != "" && h.GetSpec().PrivateKeyPath == "" {
		h.Spec.PrivateKeyPath = option.PrivateKeyPath
	}
	return h
}

func getFileArray(filePath string) (fileArray []string, err error) {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return
	}
	if info.IsDir() {
		files, err := ioutil.ReadDir(filePath)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			fileArray = append(fileArray, filepath.Join(filePath, f.Name()))
		}
	} else {
		fileArray = append(fileArray, filePath)
	}
	return
}

func ReadTaskYaml(filePath string) (tasks []opsv1.Task, err error) {
	fileArray, err := getFileArray(filePath)
	if err != nil {
		return
	}
	for _, f := range fileArray {
		yfile, err1 := ioutil.ReadFile(f)
		if err1 != nil {
			return nil, err1
		}
		task := opsv1.Task{}
		task.Spec.Variables = make(map[string]string, 0)
		err = yaml.Unmarshal(yfile, &task)
		fmt.Println(task.Spec.Steps[0].AllowFailure)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}

	return
}

func RenderStepVariables(step *opsv1.Step, vars map[string]string) *opsv1.Step {
	for key, value := range vars {
		step.Name = strings.ReplaceAll(step.Name, fmt.Sprintf("${%s}", key), value)
		step.Script = strings.ReplaceAll(step.Script, fmt.Sprintf("${%s}", key), value)
		step.LocalFile = strings.ReplaceAll(step.LocalFile, fmt.Sprintf("${%s}", key), value)
		step.RemoteFile = strings.ReplaceAll(step.RemoteFile, fmt.Sprintf("${%s}", key), value)
	}
	return step
}

func RenderVarsVariables(vars map[string]string) map[string]string {
	for valueKey, value := range vars {
		for key, keyValue := range vars {
			if strings.Contains(value, key) {
				vars[valueKey] = strings.ReplaceAll(value, fmt.Sprintf("${%s}", key), keyValue)
			}
		}
	}
	return vars
}

func RenderWhen(when string, vars map[string]string) string {
	for key, value := range vars {
		if strings.Contains(when, key) {
			when = strings.ReplaceAll(when, fmt.Sprintf("${%s}", key), value)
		}
	}
	return when
}

func GetHostStepFunc(step opsv1.Step) func(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, to option.TaskOption) (string, bool) {
	if len(step.Script) > 0 {
		return runStepScriptOnHost
	}
	return runStepCopyOnHost
}

func runStepScriptOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (result string, isSuccessed bool) {
	stdout, exit, _ := c.Script(option.Sudo, step.Script)
	return stdout, exit == 0
}

func runStepCopyOnHost(t *opsv1.Task, c *host.HostConnection, step opsv1.Step, option option.TaskOption) (result string, isSuccessed bool) {
	return "", c.File(option.Sudo, step.Direction, step.LocalFile, step.RemoteFile) == nil
}

func GetKubeStepFunc(step opsv1.Step) func(t *opsv1.Task, c *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (string, bool) {
	if len(step.Script) > 0 {
		return runStepScriptOnKube
	}
	return runStepCopyOnKube
}

func runStepScriptOnKube(t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taksOpt option.TaskOption, kubeOpt option.KubeOption) (result string, isSuccessed bool) {
	stdout, err := kc.ScriptOnNode(node,
		option.ScriptOption{
			Sudo:       taksOpt.Sudo,
			Script:     step.Script,
			KubeOption: kubeOpt,
		})
	return stdout, err == nil
}

func runStepCopyOnKube(t *opsv1.Task, kc *kube.KubeConnection, node *corev1.Node, step opsv1.Step, taskOpt option.TaskOption, kubeOpt option.KubeOption) (result string, isSuccessed bool) {
	stdout, err := kc.FileonNode(node,
		option.FileOption{
			Sudo:       taskOpt.Sudo,
			Direction:  step.Direction,
			LocalFile:  step.LocalFile,
			RemoteFile: step.RemoteFile,
		})
	return stdout, err == nil
}
