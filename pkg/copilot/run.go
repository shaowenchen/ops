package copilot

import (
	"encoding/json"
	"github.com/shaowenchen/ops/pkg/agent"
	"github.com/shaowenchen/ops/pkg/log"
)

func RunPipeline(logger *log.Logger, history *RoleContentList, pipelinerunsManager *agent.LLMPipelineRunsManager, input string, maxTry int, creator string) (pipelinerun *agent.LLMPipelineRun, err error) {
	client := GetClient(GlobalCopilotOption.Endpoint, GlobalCopilotOption.Key)
	model := GlobalCopilotOption.Model
AGAIN:
	history.WithHistory(0)
	calls, err := ChatTools(logger, client, model, history, input, GetToolsPrompt(), 0, pipelinerunsManager.GetPipelineTools())
	if err != nil {
		return nil, err
	}
	logger.Debug.Printf("> calls: %v\n", calls)
	if len(calls) == 0 {
		if maxTry > 0 {
			maxTry--
			goto AGAIN
		}
		calls = GetDefaultToolCall()
	}
	f := calls[0].Function.Name
	a := calls[0].Function.Arguments
	vars := make(map[string]string)
	err = json.Unmarshal([]byte(a), &vars)
	if err != nil {
		return nil, err
	}
	typeRef := vars["typeRef"]
	if typeRef == "" {
		typeRef = "cluster"
	}
	nameRef := vars["nameRef"]
	nodeName := vars["nodeName"]
	pipelinerun = &agent.LLMPipelineRun{
		Creator:     creator,
		Desc:        input,
		Namespace:   "ops-system",
		PipelineRef: f,
		TypeRef:     typeRef,
		NameRef:     nameRef,
		NodeName:    nodeName,
		Variables:   vars,
	}
	logger.Debug.Printf("> run pipeline %s on %s %s , variables: %v\n", pipelinerun.PipelineRef, pipelinerun.TypeRef, pipelinerun.NameRef, pipelinerun.Variables)
	err = pipelinerunsManager.Run(logger, pipelinerun)
	return
}
