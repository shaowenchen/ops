package copilot

import (
	"github.com/shaowenchen/ops/pkg/agent"
	"github.com/shaowenchen/ops/pkg/log"
)

func RunPipeline(logger *log.Logger, chat func(string, string, *RoleContentList) (string, error), history *RoleContentList, pipelinerunsManager *agent.LLMPipelineRunsManager, input string, creator string) (pipelinerun *agent.LLMPipelineRun, err error) {
	pipelines := pipelinerunsManager.GetLLMPipelines()
	logger.Debug.Println("available pipelines num: ", len(pipelines))
	// chat intention
	history.WithHistory(0)
	_, pipeline, pipelinerun, err := ChatIntention(logger, chat, GetIntentionPrompt, pipelines, history, input, 3)
	if err != nil {
		return
	}
	// chat parameters
	history.WithHistory(0)
	_, err = ChatParameters(logger, chat, GetParametersPrompt, pipelines, history, pipeline, pipelinerun, input, 3)
	if err != nil {
		return
	}
	// run pipelinerun
	logger.Debug.Printf("> run pipeline %s on %s %s , variables: %v\n", pipelinerun.PipelineRef, pipelinerun.TypeRef, pipelinerun.NameRef, pipelinerun.Variables)
	err = pipelinerunsManager.Run(logger, pipelinerun)
	return
}
