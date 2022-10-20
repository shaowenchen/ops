package pipeline

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Pipeline struct {
	Variables map[string]string
	Steps     []Step
	Name      string
}

type Step struct {
	Name       string
	Script     string
	LocalFile  string
	RemoteFile string
	Direction  string
}

func renderFunc(step *Step) (err error) {
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
			return err
		}
	}
	return
}

func renderStepVariables(step Step, vars map[string]string) Step {
	for key, value := range vars {
		step.Name = strings.ReplaceAll(step.Name, fmt.Sprintf("${%s}", key), value)
		step.Script = strings.ReplaceAll(step.Script, fmt.Sprintf("${%s}", key), value)
		step.LocalFile = strings.ReplaceAll(step.LocalFile, fmt.Sprintf("${%s}", key), value)
		step.RemoteFile = strings.ReplaceAll(step.RemoteFile, fmt.Sprintf("${%s}", key), value)
	}
	return step
}

func renderVarsVariables(vars map[string]string) map[string]string {
	for valueKey, value := range vars {
		for key, keyValue := range vars {
			if strings.Contains(value, key) {
				vars[valueKey] = strings.ReplaceAll(value, fmt.Sprintf("${%s}", key), keyValue)
			}
		}
	}
	return vars
}

func readPipelineYaml(filePath string) (pipelines []Pipeline, err error) {
	fileArray, err := getFileArray(filePath)
	if err != nil {
		return
	}
	for _, f := range fileArray {
		yfile, err1 := ioutil.ReadFile(f)
		if err1 != nil {
			return nil, err1
		}
		pipeline := Pipeline{}
		pipeline.Variables = make(map[string]string, 0)
		err = yaml.Unmarshal(yfile, &pipeline)
		if err != nil {
			return
		}
		pipelines = append(pipelines, pipeline)
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
