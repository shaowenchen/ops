package copilot

import (
	"encoding/json"
	"fmt"
	"github.com/shaowenchen/ops/pkg/agent"
	"strings"
)

func GetIntentionPrompt(pipelines []agent.LLMPipeline) string {
	if len(pipelines) == 0 {
		return ""
	}
	var b strings.Builder
	for _, pipeline := range pipelines {
		b.WriteString(fmt.Sprintf("- %s(%s)\n", pipeline.Name, pipeline.Desc))
	}
	return `Please select the most appropriate option to classify the intention of the user. 
Don't ask any more questions, just select the option.
Must be one of the following options:

` + b.String()
}

func GetParametersPrompt(pipeline agent.LLMPipeline) string {
	// add vars
	var desc strings.Builder
	desc.WriteString(fmt.Sprintf("The %s pipeline is used to %s.\n", pipeline.Name, pipeline.Desc))
	if len(pipeline.Variables) >= 0 {
		desc.WriteString("It requires the following parameters:\n")
	}
	for k, v := range pipeline.Variables {
		parmDesc := fmt.Sprintf("- %s(%s)\n", k, v.Desc)
		if len(v.Enum) > 0 {
			parmDesc = fmt.Sprintf("- %s(%s available options: %s)\n", k, v.Desc, strings.Join(v.Enum, ", "))
		}
		desc.WriteString(parmDesc)
	}
	outputScheme := map[string]string{}
	for _, k := range pipeline.GetFullVariables() {
		if k.Key == "typeRef" {
			outputScheme[k.Key] = "cluster"
		} else if k.Key == "nodeName" {
			outputScheme[k.Key] = "try extract from the input, if not found, use anymaster"
		} else {
			outputScheme[k.Key] = "need to extract from the input"
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
