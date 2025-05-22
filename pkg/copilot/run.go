package copilot

import (
	"encoding/json"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
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

func RunPipeline(logger *log.Logger, useTools bool, chat func(string, string, *ChatMessage) (string, error), history *ChatMessage, pipelinerunsManager *PipelineRunsManager, input string, extraVariables map[string]string) (*opsv1.PipelineRun, int, error) {
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
		_, pipeline, err = ChatIntention(logger, chat, GetActionPrompt, pipelines, history, input, 1)
		if err != nil {
			return nil, ExitSystemError, err
		}
	}
	// 1 - get pipelineObj
	for _, p := range pipelines {
		if p.Name == pipeline {
			pipelineObj = &p
			break
		}
	}
	// if pipelineObj == nil {
	// 	history.AddAssistantContent("can not find available actions to run")
	// 	return nil, ExitCodeIntentionEmpty, nil
	// }
	// 2 - if pipeline is default, run chat and return
	if strings.ToLower(pipeline) == "default" || pipelineObj == nil {
		output, err := chat(input, GetChatPrompt(), history)
		if err != nil {
			return nil, ExitSystemError, err
		}
		pipelinerun := opsv1.NewPipelineRun(pipelineObj)
		pipelinerun.Spec.PipelineRef = "default"

		taskRunStatus := &opsv1.TaskRunStatus{
			RunStatus: opsconstants.StatusSuccessed,
			TaskRunNodeStatus: map[string]*opsv1.TaskRunNodeStatus{
				"default": {
					TaskRunStep: []*opsv1.TaskRunStep{
						{
							StepOutput: output,
						},
					},
				},
			},
		}

		pipelinerun.Status.PipelineRunStatus = []opsv1.PipelineRunTaskStatus{
			{
				TaskName:      "default",
				TaskRef:       "default",
				TaskRunStatus: taskRunStatus,
			},
		}
		pipelinerun.Status.RunStatus = opsconstants.StatusSuccessed
		history.AddAssistantContent(output)
		return pipelinerun, ExitCodeDefault, nil
	}
	// 3-1 - chat parameters
	if !useTools {
		// chat parameters
		_, variables, err = ChatParameters(logger, chat, GetActionParametersPrompt, pipelines, clusters, history, pipelineObj, input, 3)
		if err != nil {
			return nil, ExitSystemError, err
		}
	}

	// 3-2 - validate parameters
	inValidParameters := false
	if variables != nil {
		for k, _ := range variables {
			if val, ok := variables[k]; ok && val != "" {
				variables[k] = val
			} else if _, ok := pipelineObj.Spec.Variables[k]; ok && pipelineObj.Spec.Variables[k].Value != "" {
				variables[k] = pipelineObj.Spec.Variables[k].Value
			} else if _, ok := pipelineObj.Spec.Variables[k]; ok && pipelineObj.Spec.Variables[k].Default != "" {
				variables[k] = pipelineObj.Spec.Variables[k].Default
			} else if _, ok := extraVariables[k]; !ok && pipelineObj.Spec.Variables[k].Required {
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
		history.AddAssistantContent("can not find available parameters to run")
		return nil, ExitCodeParametersNotFound, nil
	}
	// 3 - create pipelinerun and return
	pipelinerun := opsv1.NewPipelineRun(pipelineObj)
	pipelinerun.Spec.Variables = variables
	logger.Debug.Printf("> run pipeline %s on %s, variables: %v\n", pipelinerun.Spec.PipelineRef, pipelinerun.Namespace, pipelinerun.Spec.Variables)

	// 4 - run pipeline
	err = pipelinerunsManager.Run(logger, pipelinerun)
	history.AddAssistantContent(input)
	return pipelinerun, ExitCodeDefault, err
}
