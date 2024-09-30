package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
)

var GlobalCopilotOption *option.CopilotOption

type RoleContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const RoleUser = openai.ChatMessageRoleUser
const RoleSystem = openai.ChatMessageRoleSystem
const RoleAssistant = openai.ChatMessageRoleAssistant

type RoleContentList []RoleContent

func (rcl *RoleContentList) AddSystemContent(content string) *RoleContentList {
	*rcl = append(*rcl, RoleContent{
		Role:    RoleSystem,
		Content: content,
	})
	return rcl
}

func (rcl *RoleContentList) AddUserContent(content string) *RoleContentList {
	*rcl = append(*rcl, RoleContent{
		Role:    RoleUser,
		Content: content,
	})
	return rcl
}

func (rcl *RoleContentList) AddAssistantContent(content string) *RoleContentList {
	*rcl = append(*rcl, RoleContent{
		Role:    RoleAssistant,
		Content: content,
	})
	return rcl
}

func (rcl *RoleContentList) AddChatPairContent(ask, reply string) *RoleContentList {
	return rcl.AddUserContent(ask).AddAssistantContent(reply)
}

func (rcl *RoleContentList) AddRunCodePairContent(code, reply string) *RoleContentList {
	content := fmt.Sprintf("After run code:\n%s\n System output: %s\n", code, reply)
	return rcl.AddUserContent(content)
}

func (rcl *RoleContentList) IsEndWithRunCodePair() bool {
	if len(*rcl) == 0 {
		return false
	}
	return strings.HasSuffix((*rcl)[len(*rcl)-1].Content, "After run code:")
}

func (rcl *RoleContentList) Merge(merge *RoleContentList) *RoleContentList {
	if merge == nil {
		return rcl
	}
	if rcl == nil {
		return merge
	}
	*rcl = append(*rcl, *merge...)
	return rcl
}

func (rcl *RoleContentList) WithHistory(maxHistory int) *RoleContentList {
	if len(*rcl) > maxHistory*2 {
		*rcl = (*rcl)[len(*rcl)-maxHistory:]
	}
	return rcl
}

func (rcl *RoleContentList) GetOpenaiChatCompletionMessages() (messageList []openai.ChatCompletionMessage) {
	if rcl == nil {
		return
	}
	for _, roleContent := range *rcl {
		messageList = append(messageList, openai.ChatCompletionMessage{
			Role:    roleContent.Role,
			Content: roleContent.Content,
		})
	}
	return
}

func (rcl *RoleContentList) GetOpenaiChatCompletionMessagesWithSystem(system string) (messageList []openai.ChatCompletionMessage) {
	messageList = rcl.GetOpenaiChatCompletionMessages()
	messageList = append(messageList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: system,
	})
	return
}

func GetClient(endpoint, key string) *openai.Client {
	config := openai.DefaultConfig(key)
	config.BaseURL = endpoint
	if strings.Contains(endpoint, "azure.com") {
		config = openai.DefaultAzureConfig(key, endpoint)
	}
	return openai.NewClientWithConfig(config)
}

func BuildOpenAIChat(endpoint, key, model string, history *RoleContentList, input, system string, temperature float32) (chat func(string, string, *RoleContentList) (string, error), err error) {
	client := GetClient(endpoint, key)
	if client == nil {
		err = fmt.Errorf("build openai client failed")
		return
	}
	chat = func(input, system string, history *RoleContentList) (string, error) {
		if history == nil {
			history = &RoleContentList{}
		}
		history = history.AddSystemContent(system)
		history = history.AddUserContent(input)
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:       model,
				Messages:    history.GetOpenaiChatCompletionMessages(),
				Temperature: temperature,
			},
		)
		if err != nil {
			return "", err
		}
		return resp.Choices[0].Message.Content, nil
	}
	return
}

func ChatIntention(logger *log.Logger, chat func(string, string, *RoleContentList) (string, error), buildSystem func([]opsv1.Pipeline) string, pipelines []opsv1.Pipeline, history *RoleContentList, input string, maxTryTimes int) (output string, pipeline opsv1.Pipeline, pr *opsv1.PipelineRun, err error) {
Again:
	system := buildSystem(pipelines)
	output, err = chat(input, system, history)
	if err != nil {
		logger.Error.Printf("llm chatIntention error: %v\n", err)
		if maxTryTimes > 0 {
			maxTryTimes--
			goto Again
		}
		return
	}
	logger.Debug.Printf("llm chatIntention output: %s\n", output)
	// check pipelines
	avaliables := make([]string, 0)
	for _, p := range pipelines {
		if strings.Contains(output, p.Name) || strings.Contains(output, p.Spec.Desc) {
			avaliables = append(avaliables, p.Name)
		}
	}
	if len(avaliables) == 0 {
		if maxTryTimes > 0 {
			maxTryTimes--
			logger.Debug.Printf("try again chatIntention, maxTryTimes: %d\n", maxTryTimes)
			goto Again
		}
		return
	}
	// return max length match
	available := avaliables[0]
	for _, name := range avaliables {
		if len(name) > len(available) {
			available = name
		}
	}
	for _, p := range pipelines {
		if p.Name == available {
			pipeline = p
			prT := opsv1.NewPipelineRun(&p)
			pr = &prT
			return
		}
	}
	if maxTryTimes > 0 {
		maxTryTimes--
		goto Again
	}
	return
}

func ChatParameters(logger *log.Logger, chat func(string, string, *RoleContentList) (string, error), buildSystem func(opsv1.Pipeline, []opsv1.Cluster) string, pipelines []opsv1.Pipeline, clusters []opsv1.Cluster, history *RoleContentList, pipeline opsv1.Pipeline, pr *opsv1.PipelineRun, input string, maxTryTimes int) (output string, err error) {
Again:
	system := buildSystem(pipeline, clusters)
	output, err = chat(input, system, history)
	if err != nil {
		logger.Error.Printf("llm chatParameters error: %v\n", err)
		return
	}
	clusterEnums := []string{}
	for _, cluster := range clusters {
		clusterEnums = append(clusterEnums, cluster.Name)
	}
	// clean string
	start := -1
	end := -1
	for i := 0; i < len(output); i++ {
		char := output[i]
		if char == '{' && start == -1 && end == -1 {
			start = i
		} else if char == '}' && start != -1 && end == -1 {
			end = i
		} else if start != -1 && end != -1 {
			break
		}
	}
	if start != -1 && end != -1 && start < end {
		output = output[start : end+1]
	}
	logger.Debug.Printf("llm chatParameters cleaed output: %s\n ", output)
	// validate json
	outputVars := make(map[string]string)
	err1 := json.Unmarshal([]byte(output), &outputVars)
	if err1 != nil {
		logger.Error.Printf("json marshal error: %v\n", err)
		if maxTryTimes > 0 {
			maxTryTimes--
			logger.Debug.Printf("try again chatParameters, maxTryTimes: %d\n", maxTryTimes)
			goto Again
		}
	}

	for k, v := range pipeline.Spec.Variables {
		if k == "host" {
			v.Enums = clusterEnums
		}
		if v.Value != "" {
			pr.Spec.Variables[k] = v.Value
		} else if _, ok := outputVars[k]; ok {
			outV := outputVars[k]
			pr.Spec.Variables[k] = ""
			if len(v.Enums) > 0 {
				found := false
				for _, enum := range v.Enums {
					if outV == enum {
						found = true
						break
					}
				}
				if found {
					pr.Spec.Variables[k] = outV
				}
			} else {
				pr.Spec.Variables[k] = outV
			}
		} else {
			pr.Spec.Variables[k] = ""
		}
	}
	return
}
