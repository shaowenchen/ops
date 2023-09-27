package copilot

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/shaowenchen/ops/pkg/log"
)

func GetClient(key, endpoint string) *openai.Client {
	config := openai.DefaultConfig(key)
	config.BaseURL = endpoint
	if strings.Contains(endpoint, "openai.azure.com") {
		config = openai.DefaultAzureConfig(key, endpoint)
	}
	return openai.NewClientWithConfig(config)
}

func ChatCompletetionAsk(logger *log.Logger, client *openai.Client, model string, history *RoleContentList, input string) (output string, err error) {
	logger.Debug.Printf("llm ask history: %v\nllm ask input: %v\n ", history, input)
	newHistory := RoleContentList{}
	newHistory.Merge(history)
	newHistory.AddUserContent(input).AddSystemContent(SystemAskMessage)
	return chatCompletetion(logger, client, model, history)
}

func ChatCompletetionCode(logger *log.Logger, client *openai.Client, model string, history *RoleContentList, input string) (output string, err error) {
	logger.Debug.Printf("llm code history: %v\nllm code input: %v\n ", history, input)
	newHistory := RoleContentList{}
	newHistory.Merge(history)
	newHistory.AddUserContent(input).AddUserContent(SystemCodeMessage)
	return chatCompletetion(logger, client, model, history)
}

func chatCompletetion(logger *log.Logger, client *openai.Client, model string, history *RoleContentList) (output string, err error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    history.GetOpenaiChatCompletionMessages(),
			Temperature: 0,
		},
	)
	if err != nil {
		logger.Error.Printf("llm chat error: %v\n", err)
		return
	}
	return resp.Choices[0].Message.Content, err
}
