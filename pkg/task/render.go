package task

import (
	"fmt"

	"errors"
	"os"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"gopkg.in/yaml.v3"
)

func GetRealVariables(t *opsv1.Task, taskOpt option.TaskOption) (map[string]string, error) {
	globalVariables := make(map[string]string)
	// cli > env > yaml
	utils.MergeMap(globalVariables, t.Spec.Variables)
	utils.MergeMap(globalVariables, utils.GetAllOsEnv())
	utils.MergeMap(globalVariables, taskOpt.Variables)

	globalVariables = RenderVarsVariables(globalVariables)
	// check variable in task is not empty
	for key := range t.Spec.Variables {
		if len(globalVariables[key]) == 0 {
			return nil, errors.New("please set variable: " + key)
		}
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

func ReadTaskYaml(filePath string) (tasks []opsv1.Task, err error) {
	fileArray, err := utils.GetFileArray(filePath)
	if err != nil {
		return
	}
	for _, f := range fileArray {
		yfile, err1 := os.ReadFile(f)
		if err1 != nil {
			return nil, err1
		}
		task := opsv1.Task{}
		task.Spec.Variables = make(map[string]string, 0)
		err = yaml.Unmarshal(yfile, &task)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}

	return
}

func RenderStepVariables(step *opsv1.Step, vars map[string]string) *opsv1.Step {
	// replace all
	// but in the case, result=${message} will cause error
	// so we need to replace twice
	// this implementation is not good
	f := func() {
		for key, value := range vars {
			step.Name = strings.ReplaceAll(step.Name, fmt.Sprintf(`${%s}`, key), value)
			step.Content = strings.ReplaceAll(step.Content, fmt.Sprintf(`${%s}`, key), value)
			step.LocalFile = strings.ReplaceAll(step.LocalFile, fmt.Sprintf(`${%s}`, key), value)
			step.RemoteFile = strings.ReplaceAll(step.RemoteFile, fmt.Sprintf(`${%s}`, key), value)
			step.Kubernetes.Action = strings.ReplaceAll(step.Kubernetes.Action, fmt.Sprintf(`${%s}`, key), value)
			step.Kubernetes.Kind = strings.ReplaceAll(step.Kubernetes.Kind, fmt.Sprintf(`${%s}`, key), value)
			step.Kubernetes.Namespace = strings.ReplaceAll(step.Kubernetes.Namespace, fmt.Sprintf(`${%s}`, key), value)
			step.Kubernetes.Name = strings.ReplaceAll(step.Kubernetes.Name, fmt.Sprintf(`${%s}`, key), value)
			step.Alert.Url = strings.ReplaceAll(step.Alert.Url, fmt.Sprintf(`${%s}`, key), value)
			step.Alert.If = strings.ReplaceAll(step.Alert.If, fmt.Sprintf(`${%s}`, key), value)
		}
	}
	f()
	f()

	return step
}

func RenderVarsVariables(vars map[string]string) map[string]string {
	for key := range vars {
		vars[key] = RenderString(vars[key], vars)
	}
	return vars
}

func RenderString(target string, vars map[string]string) string {
	for key, value := range vars {
		if strings.Contains(target, fmt.Sprintf(`${%s}`, key)) {
			target = strings.ReplaceAll(target, fmt.Sprintf(`${%s}`, key), value)
		}
	}
	return target
}
