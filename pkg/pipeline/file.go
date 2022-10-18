package pipeline

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func renderVariables(step Step, vars map[string]string) Step {
	for key, value := range vars {
		vars[key] = strings.ReplaceAll(vars[key], "$"+key, value)
	}
	for key, value := range vars {
		step.Name = strings.ReplaceAll(step.Name, "$"+key, value)
		step.Script = strings.ReplaceAll(step.Script, "$"+key, value)
		step.LocalFile = strings.ReplaceAll(step.LocalFile, "$"+key, value)
		step.RemoteFile = strings.ReplaceAll(step.RemoteFile, "$"+key, value)
	}
	return step
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
