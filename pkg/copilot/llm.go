package copilot

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"strings"
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

func ChatCompletion(logger *log.Logger, client *openai.Client, model string, history *RoleContentList, input, system string, temperature float32) (output string, err error) {
	logger.Debug.Printf("llm ask history: %v, llm ask input: %v\n ", history, input)
	newHistory := RoleContentList{}
	newHistory.Merge(history)
	newHistory.AddUserContent(input).AddSystemContent(system)
	output, _, err = chatCompletetion(logger, client, model, &newHistory, temperature, nil)
	return
}

func ChatTools(logger *log.Logger, client *openai.Client, model string, history *RoleContentList, input, system string, temperature float32, tools []openai.Tool) (calls []openai.ToolCall, err error) {
	logger.Debug.Printf("llm ask history: %v, llm ask input: %v\n ", history, input)
	newHistory := RoleContentList{}
	newHistory.Merge(history)
	newHistory.AddUserContent(input).AddSystemContent(system)
	_, calls, err = chatCompletetion(logger, client, model, &newHistory, temperature, tools)
	return
}

func chatCompletetion(logger *log.Logger, client *openai.Client, model string, history *RoleContentList, temperature float32, tools []openai.Tool) (output string, calls []openai.ToolCall, err error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    history.GetOpenaiChatCompletionMessages(),
			Temperature: temperature,
			Tools:       tools,
		},
	)
	if err != nil {
		logger.Error.Printf("llm chat error: %v\n", err)
		return
	}
	return resp.Choices[0].Message.Content, resp.Choices[0].Message.ToolCalls, nil
}
