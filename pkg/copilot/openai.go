package copilot

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/shaowenchen/ops/pkg/log"
)

func GetClient(key, endpoint string) *openai.Client {
	config := openai.DefaultConfig(key)
	config.BaseURL = endpoint
	return openai.NewClientWithConfig(config)
}

func ChatCompletetionAsk(logger *log.Logger, client *openai.Client, model string, history *RoleContentList, input string) (output string, err error) {
	history.AddUserContent(input).AddSystemContent(system_aks_message)
	return chatCompletetion(logger, client, model, history)
}

func ChatCompletetionCode(logger *log.Logger, client *openai.Client, model string, history *RoleContentList, input string) (output string, err error) {
	history.AddUserContent(input).AddUserContent(system_code_message)
	return chatCompletetion(logger, client, model, history)
}

func chatCompletetion(logger *log.Logger, client *openai.Client, model string, history *RoleContentList) (output string, err error) {
	logger.Debug.Printf("llm req: %v\n", history)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: history.GetOpenaiChatCompletionMessages(),
		},
	)
	if err != nil {
		return
	}
	return resp.Choices[0].Message.Content, err
}
