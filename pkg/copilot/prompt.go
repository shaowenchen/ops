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
	b.WriteString(fmt.Sprintf("| tool | description |\n"))
	b.WriteString("|-|-|\n")
	for _, tool := range tools {
		b.WriteString(fmt.Sprintf("| %s | %s |\n", tool.Function.Name, tool.Function.Description))
	}
	return `Please select the most appropriate option based on the following list:
Options:
` + b.String() + `
Example: 
Input: ` + tools[0].Function.Description + `
Output: ` + tools[0].Function.Name
}

func GetParametersPrompt(tool openai.Tool) string {
	if tool.Function.Parameters == nil {
		return ""
	}
	var b strings.Builder
	for k, v := range tool.Function.Parameters.(jsonschema.Definition).Properties {
		b.WriteString(fmt.Sprintf("parameter: %s\n", k))
		b.WriteString(fmt.Sprintf("parameter type: string\n"))
		b.WriteString(fmt.Sprintf("description: %s\n", v.Description))
		if len(v.Enum) > 0 {
			b.WriteString(fmt.Sprintf("choice: %v\n", v.Enum))
		}
		b.WriteString("---\n")
	}
	outputScheme := map[string]string{}
	for k, _ := range tool.Function.Parameters.(jsonschema.Definition).Properties {
		outputScheme[k] = "need to extract from the input"
	}
	outputScoutputSchemeBytes, _ := json.Marshal(outputScheme)
	return `#You are an AI assistant specialized in extracting parameters from text. Your task is to analyze the given text, identify and extract key parameters, and then output the results in JSON format.

Please follow these guidelines:

1. Carefully analyze the provided text and identify possible parameters and their values.
3. Organize the extracted parameters into JSON format.
4. If the value of a parameter is uncertain, use "" as its value.
5. If no parameters are found in the text, return an empty JSON object.
6. Do not add any explanations or additional text, only output the JSON object.
7. Do not add any extra parameters excluding the fields in Output Example.

Example:
Input: "Please list pod names in cluster 1 on node 1"
Output:
{
  "nameRef": "cluster1",
  "nodeName": "node1",
  "typeRef": "cluster",
}

Now, please analyze the following text and extract parameters:
` + b.String() + `
# Output Example: ` + string(outputScoutputSchemeBytes)
}
