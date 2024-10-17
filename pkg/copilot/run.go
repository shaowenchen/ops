package copilot

import (
	"encoding/json"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
)

const ExitCodeDefault = 0
const ExitSystemError = 1
const ExitCodeIntentionEmpty = 2
const ExitCodeParametersNotFound = 3

type PipelineTool struct {
	Pipeline  string            `json:"pipeline"`
	Variables map[string]string `json:"variables"`
}

func (p PipelineTool) String() string {
	data, _ := json.Marshal(p)
	return string(data)
}

func RunPipeline(logger *log.Logger, useTools bool, chat func(string, string, *RoleContentList) (string, error), history *RoleContentList, pipelinerunsManager *PipelineRunsManager, input string, extraVariables map[string]string) (*opsv1.PipelineRun, int, error) {
	var pipeline string
	var variables map[string]string
	var pipelineObj *opsv1.Pipeline
	// get context
	pipelines, err := pipelinerunsManager.GetPipelines()
	if err != nil {
		return nil, ExitSystemError, err
	}
	clusters, err := pipelinerunsManager.GetClusters()
	if err != nil {
		return nil, ExitSystemError, err
	}
	// 1-1 - chat intention
	if useTools {
		outout, err := chat(input, GetIntentionParametersPrompt(clusters, pipelines), history)
		if err != nil {
			return nil, ExitSystemError, err
		}
		pipelineTool := PipelineTool{}
		err = json.Unmarshal([]byte(outout), &pipelineTool)
		if err != nil {
			return nil, ExitSystemError, err
		}
		pipeline = pipelineTool.Pipeline
		variables = pipelineTool.Variables
	} else {
		// chat intention
		history.WithHistory(0)
		_, pipeline, err = ChatIntention(logger, chat, GetIntentionPrompt, pipelines, history, input, 3)
		if err != nil {
			return nil, ExitSystemError, err
		}
	}
	// 1-2 - get pipelineObj
	for _, p := range pipelines {
		if p.Name == pipeline {
			pipelineObj = &p
			break
		}
	}
	if pipelineObj == nil {
		return nil, ExitCodeIntentionEmpty, nil
	}
	// 2-1 - chat parameters
	if !useTools {
		// chat parameters
		history.WithHistory(0)
		_, variables, err = ChatParameters(logger, chat, GetParametersPrompt, pipelines, clusters, history, pipelineObj, input, 3)
		if err != nil {
			return nil, ExitSystemError, err
		}
	}
	// 2-2 - validate parameters
	inValidParameters := false
	if variables != nil {
		for k, _ := range variables {
			if val, ok := variables[k]; ok && val != "" {
				variables[k] = val
			} else if _, ok := extraVariables[k]; !ok {
				inValidParameters = true
				variables[k] = ""
			}
		}
		// merge extra variables
		for k, v := range extraVariables {
			if _, ok := pipelineObj.Spec.Variables[k]; ok {
				variables[k] = v
			}
		}
	}
	if inValidParameters {
		return nil, ExitCodeParametersNotFound, nil
	}
	// 3 - create pipelinerun and return
	pipelinerun := opsv1.NewPipelineRun(pipelineObj)
	pipelinerun.Spec.Variables = variables
	logger.Debug.Printf("> run pipeline %s on %s, variables: %v\n", pipelinerun.Spec.PipelineRef, pipelinerun.Namespace, pipelinerun.Spec.Variables)
	return pipelinerun, ExitCodeDefault, pipelinerunsManager.Run(logger, pipelinerun)
}
