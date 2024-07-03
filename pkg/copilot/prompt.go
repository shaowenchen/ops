package copilot

import (
	"encoding/json"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"strings"
)

func GetIntentionPrompt(tools []openai.Tool) string {
	if len(tools) == 0 {
		return ""
	}
	var b strings.Builder
	for _, tool := range tools {
		b.WriteString(fmt.Sprintf("- %s(%s)\n", tool.Function.Name, tool.Function.Description))
	}
	return `Please select the most appropriate option to finish the task. 
Must be one of the following options: ` +
		b.String()
}

func GetParametersPrompt(tool openai.Tool) string {
	if tool.Function.Parameters == nil {
		return ""
	}
	var funcDesc strings.Builder
	funcDesc.WriteString(fmt.Sprintf("The %s function is used to %s.\n", tool.Function.Name, tool.Function.Description))
	if len(tool.Function.Parameters.(jsonschema.Definition).Properties) >= 0 {
		funcDesc.WriteString("It requires the following parameters:\n")
	}
	for k, v := range tool.Function.Parameters.(jsonschema.Definition).Properties {
		parmDesc := fmt.Sprintf("- %s(%s)\n", k, v.Description)
		if len(v.Enum) > 0 {
			parmDesc = fmt.Sprintf("- %s(%s available options: %s)\n", k, v.Description, strings.Join(v.Enum, ", "))
		}
		funcDesc.WriteString(parmDesc)
	}
	outputScheme := map[string]string{}
	for k, _ := range tool.Function.Parameters.(jsonschema.Definition).Properties {
		if k == "typeRef" {
			outputScheme[k] = "cluster"
		} else if k == "nodeName" {
			outputScheme[k] = "try extract from the input, if not found, use anymaster"
		} else {
			outputScheme[k] = "need to extract from the input"
		}
	}
	outputSchemeBytes, _ := json.Marshal(outputScheme)
	return `You are an expert in parameter extraction, and are good at accurately extracting appropriate slots and parameters from user input according to different function requirements. Please do the following:

-understand < Workflow > and output as required.
-understand the tool description, subfunction description and parameters provided.

< Workflow >
1. According to the following function definition, accurately extract the appropriate parameters from the user input, and pay attention to the data type of the parameters.
two。. Please make sure that the parameters you extract strictly follow the definition of the function and contain only the information explicitly mentioned in the user's input.
2. Output the extracted parameters in JSON format.

Function definition:` + funcDesc.String() + `
The parsing of all necessary parameters is (in JSON format):` +
		string(outputSchemeBytes)
}
