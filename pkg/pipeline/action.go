package pipeline

import (
	"fmt"

	"strings"

	"github.com/kyokomi/emoji/v2"
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/shaowenchen/opscli/pkg/log"
	"github.com/shaowenchen/opscli/pkg/utils"
)

func ActionPipeline(logger *log.Logger, option PipelineOption) (err error) {
	pipelines, err := readPipelineYaml(utils.GetAbsoluteFilePath(option.FilePath))
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	for _, p := range pipelines {
		p.Logger = logger
		globalVariables := make(map[string]string)
		// cli > env > yaml
		utils.MergeMap(globalVariables, p.Variables)
		utils.MergeMap(globalVariables, utils.GetRuntimeInfo())
		utils.MergeMap(globalVariables, utils.GetAllOsEnv())
		utils.MergeMap(globalVariables, option.Variables)

		globalVariables = p.renderVarsVariables(globalVariables)
		logger.Info.Println(emoji.Sprint(":pizza:") + "[pipeline] " + p.Name)
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
			logger.Info.Println("please set variable: ", emptyVariable)
			break
		}
		// run every pipeline
		for _, addr := range utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts)) {
			globalVariables["result"] = ""
			logger.Info.Print(utils.PlaceMiddle(fmt.Sprintf("[%s]", addr), "*"))
			for si, s := range p.Steps {
				logger.Info.Println(fmt.Sprintf("(%d/%d) %s", si+1, len(p.Steps), s.Name))
				s.When = p.renderWhen(s.When, p.renderVarsVariables(globalVariables))
				if !CheckWhen(s.When) {
					logger.Info.Println("Skip!")
					continue
				}
				s = p.renderStepVariables(s, globalVariables)
				err = p.renderFunc(&s)
				if err != nil {
					logger.Error.Println(err)
				}
				if option.Debug && len(s.Script) > 0 {
					logger.Info.Println(s.Script)
				}
				stepFunc := p.getStepFunc(s)
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
