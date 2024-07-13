package copilot

import (
	"encoding/json"
	"fmt"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
)

func GetIntentionPrompt(pipelines []opsv1.Pipeline) string {
	if len(pipelines) == 0 {
		return ""
	}
	var b strings.Builder
	for _, pipeline := range pipelines {
		b.WriteString(fmt.Sprintf("- %s(%s)\n", pipeline.Name, pipeline.Spec.Desc))
	}
	return `Please select the most appropriate option to classify the intention of the user. 
Don't ask any more questions, just select the option.
Must be one of the following options:

` + b.String()
}

func GetParametersPrompt(pipeline opsv1.Pipeline, clusters []opsv1.Cluster) string {
	// add vars
	var desc strings.Builder
	desc.WriteString(fmt.Sprintf("The %s pipeline is used to %s.\n", pipeline.Name, pipeline.Spec.Desc))
	if len(pipeline.Spec.Variables) >= 0 {
		desc.WriteString("It requires the following parameters(if enum provided, choose one of them):\n")
	}
	clusterEnums := []string{}
	for _, cluster := range clusters {
		clusterEnums = append(clusterEnums, cluster.Name)
	}
	for k, _ := range pipeline.Spec.Variables {
		vt := pipeline.Spec.Variables[k]
		if k == "nameRef" {
			vt.Enums = clusterEnums
		}
		vStr, _ := json.Marshal(vt)
		parmDesc := fmt.Sprintf("\t- %s \t %s\n", k, string(vStr))
		desc.WriteString(parmDesc)
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

< Workflow >
1. According to the following pipeline definition, accurately extract the appropriate parameters from the user input, and pay attention to the data type of the parameters.
two。. Please make sure that the parameters you extract strictly follow the definition of the pipeline and contain only the information explicitly mentioned in the user's input.
2. Output the extracted parameters in JSON format. 

Pipeline description:` + desc.String() + `
The parsing of all necessary parameters is (in JSON format):` +
		string(outputSchemeBytes)
}
