package copilot

import (
	"encoding/json"
	"fmt"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
)

func GetChatPrompt() string {
	return `You are a DevOps expert.
You excel at communicating with users and providing practical, actionable suggestions.
Keep responses clear, concise, and to the point.
Whenever possible, reply in the same language as the user's input.
Support both English and Chinese users naturally.
`
}

func GetPlanPrompt(pipelines []opsv1.Pipeline) string {
	if len(pipelines) == 0 {
		return ""
	}
	var b strings.Builder
	for _, pipeline := range pipelines {
		b.WriteString(fmt.Sprintf("- %s(%s)\n", pipeline.Name, pipeline.Spec.Desc))
	}
	return `You are an ops expert, you are good at planning tasks.
Please do the following:
-understand the user's input.
-make a plan for the user's input.
-output the plan in JSON format.

< available standard operating procedures >
` + b.String()
}

func GetIntentionParametersPrompt(clusters []opsv1.Cluster, pipelines []opsv1.Pipeline) string {
	return `{
	"pipeline": "",
	"variables": {}
	}`
}

func GetActionPrompt(pipelines []opsv1.Pipeline) string {
	if len(pipelines) == 0 {
		return ""
	}
	var b strings.Builder
	for _, pipeline := range pipelines {
		b.WriteString(fmt.Sprintf("- %s(%s)\n", pipeline.Name, pipeline.Spec.Desc))
	}
	return `Try to choose one of the following actions to resolve the input issue as effectively as possible.
If input includes keyword action, please select the action option as possible.
If you can't find the appropriate option, please choose the "other" option.
Must be one of the following options:

` + b.String()
}

func GetActionParametersPrompt(pipeline opsv1.Pipeline, clusters []opsv1.Cluster) string {
	// add vars
	var desc strings.Builder
	desc.WriteString(fmt.Sprintf("The %s pipeline is used to %s.\n", pipeline.Name, pipeline.Spec.Desc))
	if len(pipeline.Spec.Variables) >= 0 {
		desc.WriteString("It requires the following parameters(if enum provided, choose one of them):\n")
	}
	clusterEnum := []string{}
	for _, cluster := range clusters {
		clusterEnum = append(clusterEnum, cluster.Name)
	}
	for k, _ := range pipeline.Spec.Variables {
		vt := pipeline.Spec.Variables[k]
		if k == "cluster" {
			vt.Enums = clusterEnum
		}
		vStr, _ := json.Marshal(vt)
		parmDesc := fmt.Sprintf("\t- %s \t %s\n", k, string(vStr))
		desc.WriteString(parmDesc)
	}
	clustersInfo := ""
	for i := 0; i < len(clusters); i++ {
		clustersInfo += fmt.Sprintf("- %s, desc: %s\n", clusters[i].Name, clusters[i].Spec.Desc)
	}
	outputScheme := map[string]string{}
	for key, value := range pipeline.Spec.Variables {
		if value.Value != "" {
			outputScheme[key] = value.Value
		} else {
			outputScheme[key] = ""
		}
	}
	outputSchemeBytes, _ := json.Marshal(outputScheme)
	return `You are an expert in parameter extraction, and are good at accurately extracting appropriate slots and parameters from user input according to different pipeline requirements. Please do the following:

-understand < Workflow > and output as required.
-understand the pipeline description and parameters provided.

< Clusters information >
` + clustersInfo + `

< Workflow >
1. According to the following pipeline definition, accurately extract the appropriate parameters from the user input, and pay attention to the data type of the parameters.
two。. Please make sure that the parameters you extract strictly follow the definition of the pipeline and contain only the information explicitly mentioned in the user's input.
2. Output the extracted parameters in JSON format. 

Pipeline description:` + desc.String() + `
The parsing of all necessary parameters is (in JSON format):` +
		string(outputSchemeBytes)
}
