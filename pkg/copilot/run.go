package copilot

import (
	"encoding/json"

	"github.com/shaowenchen/ops/pkg/agent"
	"github.com/shaowenchen/ops/pkg/log"
)

func RunPipeline(logger *log.Logger, chat func(string, string, *RoleContentList) (string, error), history *RoleContentList, pipelinerunsManager *agent.LLMPipelineRunsManager, input string, creator string) (pipelinerun *agent.LLMPipelineRun, err error) {
	history.WithHistory(0)
	tools := pipelinerunsManager.BuildTools()
	logger.Debug.Printf("> tools length: %v\n", len(tools))
	call, err := ChatTools(logger, input, GetIntentionPrompt, GetParametersPrompt, chat, history, tools)
	if err != nil {
		return nil, err
	}
	logger.Debug.Printf("> call: %v\n", call)
	if call == nil {
		defaultCall := GetDefaultToolCall()
		call = &defaultCall
	}
	f := call.Function.Name
	a := call.Function.Arguments
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
