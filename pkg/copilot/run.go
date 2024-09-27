package copilot

import (
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
)

const ExitCodeDefault = 0
const ExitCodeIntentionEmpty = 1
const ExitCodeParametersNotFound = 2

func RunPipeline(logger *log.Logger, chat func(string, string, *RoleContentList) (string, error), history *RoleContentList, pipelinerunsManager *PipelineRunsManager, input string, extraVariables map[string]string) (prResult *opsv1.PipelineRun, exit int, err error) {
	exit = ExitCodeDefault
	pipelines, err := pipelinerunsManager.GetPipelines()
	if err != nil {
		return
	}
	clusters, err := pipelinerunsManager.GetClusters()
	if err != nil {
		return
	}
	logger.Debug.Println("available pipelines num: ", len(pipelines))
	// chat intention
	history.WithHistory(0)
	_, pipeline, prResult, err := ChatIntention(logger, chat, GetIntentionPrompt, pipelines, history, input, 3)
	if err != nil {
		return
	}
	if prResult == nil {
		exit = ExitCodeIntentionEmpty
		return
	}
	// chat parameters
	history.WithHistory(0)
	ChatParameters(logger, chat, GetParametersPrompt, pipelines, clusters, history, pipeline, prResult, input, 3)
	if pipeline.Spec.Variables != nil {
		variables := map[string]string{}
		for k, _ := range pipeline.Spec.Variables {
			if val, ok := prResult.Spec.Variables[k]; ok && val != "" {
				variables[k] = val
			} else if _, ok := extraVariables[k]; !ok {
				exit = ExitCodeParametersNotFound
				variables[k] = ""
			}
		}
		// merge extra variables
		for k, v := range extraVariables {
			variables[k] = v
		}
		prResult.Spec.Variables = variables
	}
	// skip run pr
	if exit != ExitCodeDefault {
		return
	}
	// validate parameters
	// run pipelinerun
	logger.Debug.Printf("> run pipeline %s on %s, variables: %v\n", prResult.Spec.PipelineRef, prResult.Namespace, prResult.Spec.Variables)
	err = pipelinerunsManager.Run(logger, prResult)
	return
}
