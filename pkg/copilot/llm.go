package copilot

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
)

var GlobalCopilotOption *option.CopilotOption

type ChatCodeResponse Langcode

type Langcode struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Content  string `json:"content"`
}

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

func ChatTools(logger *log.Logger, input string, buildIntentionSystem func([]openai.Tool) string, buildParametersSystem func(openai.Tool) string, chat func(string, string, *RoleContentList) (string, error), history *RoleContentList, tools []openai.Tool) (call *openai.ToolCall, err error) {
	intentMaxTry := 3
	parametersMaxTry := 3
	intentionSystem := buildIntentionSystem(tools)
	// 1/2, try to get intention
IntentMaxAgain:
	output, tool, err := chatIntention(logger, input, intentionSystem, chat, history, tools)
	logger.Debug.Printf("llm intention output: %s\n ", output)
	if err != nil {
		time.Sleep(1 * time.Second)
		if intentMaxTry > 0 {
			intentMaxTry--
			goto IntentMaxAgain
		}
		return
	}
	if tool.Function == nil || tool.Function.Name == "" {
		logger.Info.Printf("llm intent function not found: %v\n", err)
		time.Sleep(1 * time.Second)
		if intentMaxTry > 0 {
			intentMaxTry--
			goto IntentMaxAgain
		}
		return
	}
	// 2/2, try to get parameters
	parametersSystem := buildParametersSystem(tool)
parametersMaxTryAgain:
	output, call, err = chatParameters(logger, input, parametersSystem, chat, history, tool)
	logger.Debug.Printf("llm chatParameters output: %v\n ", output)
	if err != nil {
		time.Sleep(1 * time.Second)
		if parametersMaxTry > 0 {
			parametersMaxTry--
			goto parametersMaxTryAgain
		}
		return
	}
	return
}

func chatIntention(logger *log.Logger, input string, system string, chat func(string, string, *RoleContentList) (string, error), history *RoleContentList, tools []openai.Tool) (output string, tool openai.Tool, err error) {
	output, err = chat(input, system, history)
	if err != nil {
		logger.Error.Printf("llm chatIntention error: %v\n", err)
		return
	}
	// not openai
	for _, t := range tools {
		if strings.Contains(output, t.Function.Name) {
			tool = t
			return
		}
	}
	return
}

func chatParameters(logger *log.Logger, input, system string, chat func(string, string, *RoleContentList) (string, error), history *RoleContentList, tool openai.Tool) (output string, call *openai.ToolCall, err error) {
	output, err = chat(input, system, history)
	if err != nil {
		logger.Error.Printf("llm chatParameters error: %v\n", err)
		return
	}
	// clean string
	start := -1
	end := -1

	for i, char := range output {
		if char == '{' {
			start = i
			break
		}
	}
	for i := len(output) - 1; i >= 0; i-- {
		if output[i] == '}' {
			end = i
			break
		}
	}
	if start != -1 && end != -1 && start < end {
		output = output[start : end+1]
	} else {
		output = "{}"
	}

	call = &openai.ToolCall{
		Function: openai.FunctionCall{
			Name:      tool.Function.Name,
			Arguments: output,
		},
	}
	return
}
