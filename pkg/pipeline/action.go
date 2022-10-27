package pipeline

import (
	"fmt"

	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/shaowenchen/opscli/pkg/utils"
	"strings"
)

func ActionPipeline(option PipelineOption) (err error) {
	pipelines, err := readPipelineYaml(option.FilePath)
	if err != nil {
		fmt.Println(err)
	}
	for _, p := range pipelines {
		globalVariables := make(map[string]string)
		// override Variables
		utils.MergeMap(globalVariables, p.Variables)
		utils.MergeMap(globalVariables, utils.GetAllOsEnv())
		utils.MergeMap(globalVariables, utils.GetRuntimeInfo())
		utils.MergeMap(globalVariables, option.Variables)

		globalVariables = renderVarsVariables(globalVariables)
		fmt.Println("[pipeline] " + p.Name)
		if len(option.Hosts) == 0 {
			option.Hosts = host.LocalHostIP
		}
		// check variable in pipeline is not empty
		emptyVariable := ""
		for key, _ := range p.Variables {
			if len(strings.TrimSpace(globalVariables[key])) == 0 {
				emptyVariable = key
				break
			}
		}
		if len(emptyVariable) > 0 {
			fmt.Println("please set variable: ", emptyVariable)
			break
		}
		// run every pipeline
		for _, addr := range utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts)) {
			globalVariables["result"] = ""
			for _, s := range p.Steps {
				fmt.Println(fmt.Sprintf("[%s] %s", addr, s.Name))
				s.When = renderWhen(s.When, renderVarsVariables(globalVariables))
				if !CheckWhen(s.When) {
					fmt.Println("Skip!")
					continue
				}
				s = renderStepVariables(s, globalVariables)
				err = renderFunc(&s)
				if err != nil {
					utils.LogError(err)
				}
				if option.Debug {
					fmt.Println(s.Script)
				}
				stepFunc := getStepFunc(s)
				var tempOption = option
				tempOption.Hosts = addr
				stepResult, isSuccessed := stepFunc(s, tempOption)
				globalVariables["result"] = stepResult
				if s.AllowFailure == false && isSuccessed == false {
					break
				}
			}
		}
	}
	return
}
