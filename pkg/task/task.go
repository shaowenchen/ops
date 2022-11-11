package task

import (
	"fmt"

	"strings"

	"github.com/kyokomi/emoji/v2"
	"github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

func RunTaskOnHost(logger *log.Logger, t v1.Task, h *v1.Host, option TaskOption)(err error){
	globalVariables := make(map[string]string)
	// cli > env > yaml
	utils.MergeMap(globalVariables, t.Spec.Variables)
	utils.MergeMap(globalVariables, utils.GetRuntimeInfo())
	utils.MergeMap(globalVariables, utils.GetAllOsEnv())
	utils.MergeMap(globalVariables, option.Variables)

	globalVariables = RenderVarsVariables(globalVariables)
	logger.Info.Println(emoji.Sprint(":pizza:") + "[task] " + t.Name)
	// check variable in task is not empty
	emptyVariable := ""
	for key := range t.Spec.Variables {
		if len(strings.TrimSpace(globalVariables[key])) == 0 {
			emptyVariable = key
			break
		}
	}
	if len(emptyVariable) > 0 {
		logger.Info.Println("please set variable: ", emptyVariable)
		return
	}

	globalVariables["result"] = ""
	logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", h.Spec.Address)))
	c, err := host.NewHostConnection(h.Spec.Address, option.Port, option.Username, option.Password, option.Password)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	for si, s := range t.Spec.Steps {
		var sp = &s
		logger.Info.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(t.Spec.Steps), s.Name))
		s.When = RenderWhen(s.When, globalVariables)
		if !CheckWhen(s.When) {
			logger.Info.Println("Skip!")
			continue
		}
		sp = RenderStepVariables(sp, globalVariables)
		sp, err = RenderFunc(&s)
		if err != nil {
			logger.Error.Println(err)
		}
		if option.Debug && len(s.Script) > 0 {
			logger.Info.Println(s.Script)
		}
		stepFunc := GetStepFunc(s)
		var tempOption = option
		tempOption.Hosts = h.Spec.Address
		stepResult, isSuccessed := stepFunc(&t, c, s, tempOption)
		logger.Info.Println(stepResult)
		globalVariables["result"] = stepResult
		if s.AllowFailure == false && isSuccessed == false {
			break
		}
	}
	return
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

func ReadTaskYaml(filePath string) (tasks []v1.Task, err error) {
	fileArray, err := getFileArray(filePath)
	if err != nil {
		return
	}
	for _, f := range fileArray {
		yfile, err1 := ioutil.ReadFile(f)
		if err1 != nil {
			return nil, err1
		}
		task := v1.Task{}
		task.Spec.Variables = make(map[string]string, 0)
		err = yaml.Unmarshal(yfile, &task)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}

	return
}

func RenderFunc(step *v1.Step) (*v1.Step, error) {
	var err error
	reg := regexp.MustCompile(`([a-zA-Z])+(\([^\)]*\))`)
	funcStrList := reg.FindAllString(step.Script, -1)
	for _, item := range funcStrList {
		funcResult := []reflect.Value{}
		funcInfo := strings.Split(item, "(")
		// there is no param
		if strings.HasPrefix(funcInfo[1], ")") {
			funcResult, err = CallMap(funcInfo[0])
		} else {
			funcParams := strings.Split(funcInfo[1], ")")
			params := strings.Split(funcParams[0], ",")
			var paramsi []interface{}
			for _, param := range params {
				paramsi = append(paramsi, param)
			}
			funcResult, err = CallMap(funcInfo[0], paramsi...)
		}
		if len(funcResult) > 0 {
			step.Script = strings.ReplaceAll(step.Script, item, funcResult[0].String())
		}

		if err != nil {
			return step, err
		}
	}
	return step, err
}

func RenderStepVariables(step *v1.Step, vars map[string]string) *v1.Step {
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

func GetStepFunc(step v1.Step) func(t *v1.Task, c *host.HostConnection, step v1.Step, to TaskOption) (string, bool) {
	if len(step.Script) > 0 {
		return runStepScript
	}
	return runStepCopy
}

func runStepScript(t *v1.Task, c *host.HostConnection, step v1.Step, option TaskOption) (result string, isSuccessed bool) {
	stdout, exit, _ := c.Script(option.Sudo, step.Script)
	return stdout, exit == 0
}

func runStepCopy(t *v1.Task, c *host.HostConnection, step v1.Step, option TaskOption) (result string, isSuccessed bool) {
	return "", c.File(option.Sudo, step.Direction, step.LocalFile, step.RemoteFile) == nil
}
