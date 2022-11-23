package task

import (
	"fmt"

	"errors"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
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

func ReadTaskYaml(filePath string) (tasks []opsv1.Task, err error) {
	fileArray, err := utils.GetFileArray(filePath)
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
