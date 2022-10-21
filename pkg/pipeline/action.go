package pipeline

import (
	"fmt"
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/shaowenchen/opscli/pkg/utils"
)

func ActionPipeline(option PipelineOption) (err error) {
	pipelines, err := readPipelineYaml(option.FilePath)
	if err != nil {
		fmt.Println(err)
	}
	for _, p := range pipelines {
		// override Variables
		for key, value := range utils.GetAllOsEnv() {
			p.Variables[key] = value
		}
		for key, value := range utils.GetRuntimeInfo() {
			p.Variables[key] = value
		}
		for key, value := range option.Variables {
			p.Variables[key] = value
		}
		p.Variables = renderVarsVariables(p.Variables)
		fmt.Println("[pipeline] " + p.Name)
		if len(option.Hosts) == 0 {
			option.Hosts = host.LocalHostIP
		}
		for _, addr := range utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts)) {
			for _, s := range p.Steps {
				fmt.Println(fmt.Sprintf("[%s] %s", addr, s.Name))
				s.When = renderWhen(s.When, renderVarsVariables(p.Variables))
				if !CheckWhen(s.When) {
					fmt.Println("Skip!")
					continue
				}
				s = renderStepVariables(s, p.Variables)
				err = renderFunc(&s)
				if err != nil {
					return utils.LogError(err)
				}
				if option.Debug {
					fmt.Println(s.Script)
				}
				stepFunc, err1 := getStepFunc(s)
				if err != nil {
					fmt.Println(err)
					return err1
				}
				var tempOption = option
				tempOption.Hosts = addr
				err1 = stepFunc(s, tempOption)
				if err != nil {
					fmt.Println(err)
					return utils.LogError(err1)
				}
			}
		}
	}
	return
}
