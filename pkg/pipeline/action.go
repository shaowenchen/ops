package pipeline

import (
	"fmt"

	"strings"

	"github.com/kyokomi/emoji/v2"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
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
		// check variable in pipeline is not empty
		emptyVariable := ""
		for key := range p.Variables {
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
		hosts, _ := utils.AnalysisHostsParameter(option.Hosts)
		for _, addr := range hosts {
			globalVariables["result"] = ""
			logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", addr)))
			host, err := host.NewHost(addr, option.Port, option.Username, option.Password, option.Password)
			if err != nil {
				logger.Error.Println(err)
				continue
			}
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
				stepResult, isSuccessed := stepFunc(host, s, tempOption)
				globalVariables["result"] = stepResult
				if s.AllowFailure == false && isSuccessed == false {
					break
				}
			}
		}
	}
	return
}
