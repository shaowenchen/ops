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
	b.WriteString(fmt.Sprintf("| tool | description | parameters |\n"))
	b.WriteString("|-|-|-|\n")
	for _, tool := range tools {
		parametersBytes, _ := json.Marshal(tool.Function.Parameters)
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", tool.Function.Name, tool.Function.Description, string(parametersBytes)))
		b.WriteString(fmt.Sprintf("| %s | %s |\n", tool.Function.Name, tool.Function.Description))
	}
	return `# do not give answer directly, just choose one tool to solve the problem
# use the following tools to solve\diagnose the problem
` + b.String()
}

func GetParametersPrompt(tool openai.Tool) string {
	parametersBytes, _ := json.Marshal(tool.Function.Parameters)
	if tool.Function.Parameters == nil {
		return ""
	}
	outputScheme := map[string]string{}
	for k, _ := range tool.Function.Parameters.(jsonschema.Definition).Properties {
		outputScheme[k] = "need to extract from the input"
	}
	outputScoutputSchemeBytes, _ := json.Marshal(outputScheme)
	return `#You are an AI assistant specialized in extracting parameters from text. Your task is to analyze the given text, identify and extract key parameters, and then output the results in JSON format.

Please follow these guidelines:

1. Carefully analyze the provided text and identify possible parameters and their values.
2. Parameters may include, but are not limited to: dates, times, locations, names, quantities, amounts, etc.
3. Organize the extracted parameters into JSON format, using meaningful key names.
4. If the value of a parameter is uncertain, use null as its value.
5. If no parameters are found in the text, return an empty JSON object.
6. Do not add any explanations or additional text, only output the JSON object.

Example:
Input: "Please list pod names in cluster 1 on node 1"
Output:
{
  "nameRef": "cluster1",
  "nodeName": "node1",
  "typeRef": "cluster",
}

Now, please analyze the following text and extract parameters: ` + string(parametersBytes) + `# Output Example: ` + string(outputScoutputSchemeBytes)
}
