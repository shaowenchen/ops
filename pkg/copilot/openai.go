package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/shaowenchen/ops/pkg/log"
	opsopt "github.com/shaowenchen/ops/pkg/option"
)

var historyList []openai.ChatCompletionMessage

func chatCompletetion(logger *log.Logger, client *openai.Client, model string, messages []openai.ChatCompletionMessage) (output string, err error) {
	h := append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: system_message})
	logger.Debug.Printf("llm req: %v\n", h)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: h,
		},
	)
	if err != nil {
		return
	}
	return resp.Choices[0].Message.Content, err
}

func Chat(logger *log.Logger, opt opsopt.CopilotOption, input string) (message string, langcodeList []Langcode, err error) {
	config := openai.DefaultConfig(opt.Key)
	config.BaseURL = opt.Endpoint
	client := openai.NewClientWithConfig(config)
	historyList = append(historyList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input,
	})
	if len(historyList) > opt.History {
		historyList = historyList[len(historyList)-opt.History:]
	}
	content, err := chatCompletetion(logger, client, opt.Model, historyList)
	if err != nil {
		return
	}
	logger.Debug.Printf("llm resp: %v\n", content)
	resp := ChatResponse{}
	err = json.Unmarshal([]byte(content), &resp)
	if err != nil {
		return
	}
	return resp.Message, resp.Steps, err
}

func setOpscliRole(historyList []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	return append(historyList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: system_message,
	})
}

func setRunCodeRole(historyList []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	return append(historyList, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: system_message,
	})
}

func runLangCode(language, code string) (outputStr string, err error) {
	language = strings.ToLower(language)

	var output []byte

	if language == "python" {
		output, err = runPythonCode(code)
	} else if language == "bash" {
		output, err = runBashCode(code)
	} else {
		return "", errors.New("not support language")
	}
	outputStr = string(output)
	return
}

func runPythonCode(code string) (output []byte, err error) {
	cmd := exec.Command("python", "-c", code)
	return cmd.Output()
}

func runBashCode(code string) (output []byte, err error) {
	cmd := exec.Command("bash", "-c", code)
	return cmd.Output()
}
