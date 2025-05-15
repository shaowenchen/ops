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
const DefaultMaxHistory = 5

func NewChatMessage(maxHistory int) *ChatMessage {
	roleContentList := make([]RoleContent, 0)
	return &ChatMessage{
		RoleContentList: roleContentList,
		MaxHistory:      maxHistory,
	}
}

func NewDefaultChatMessage() *ChatMessage {
	return NewChatMessage(DefaultMaxHistory)
}

type ChatMessage struct {
	RoleContentList []RoleContent `json:"role_content_list"`
	MaxHistory      int           `json:"max_history"`
}

func (msg *ChatMessage) AddSystemContent(content string) *ChatMessage {
	// First try to find and replace existing system message
	systemFound := false
	for i, rc := range msg.RoleContentList {
		if rc.Role == RoleSystem {
			msg.RoleContentList[i].Content = content
			systemFound = true
			break
		}
	}

	// If no system message found, add a new one
	if !systemFound {
		msg.RoleContentList = append(msg.RoleContentList, RoleContent{
			Role:    RoleSystem,
			Content: content,
		})
	}

	msg.TrimHistory()
	return msg
}

func (msg *ChatMessage) AddUserContent(content string) *ChatMessage {
	msg.RoleContentList = append(msg.RoleContentList, RoleContent{
		Role:    RoleUser,
		Content: content,
	})
	msg.TrimHistory()
	return msg
}

func (msg *ChatMessage) AddAssistantContent(content string) *ChatMessage {
	msg.RoleContentList = append(msg.RoleContentList, RoleContent{
		Role:    RoleAssistant,
		Content: content,
	})
	msg.TrimHistory()
	return msg
}

func (msg *ChatMessage) AddChatPairContent(ask, reply string) *ChatMessage {
	return msg.AddUserContent(ask).AddAssistantContent(reply)
}

func (msg *ChatMessage) Merge(merge *ChatMessage) *ChatMessage {
	if merge == nil {
		return msg
	}
	if msg == nil {
		return merge
	}
	msg.RoleContentList = append(msg.RoleContentList, merge.RoleContentList...)
	msg.TrimHistory()
	return msg
}

func (msg *ChatMessage) TrimHistory() *ChatMessage {
	if msg == nil || len(msg.RoleContentList) <= msg.MaxHistory {
		return msg
	}

	// Check if there's a system message
	var systemContent RoleContent
	hasSystem := false
	for _, rc := range msg.RoleContentList {
		if rc.Role == RoleSystem {
			systemContent = rc
			hasSystem = true
			break
		}
	}

	// Filter out system messages, keep only the latest (MaxHistory-1) non-system messages
	userAssistantMessages := make([]RoleContent, 0)
	for _, rc := range msg.RoleContentList {
		if rc.Role != RoleSystem {
			userAssistantMessages = append(userAssistantMessages, rc)
		}
	}

	// Keep only the most recent messages
	if len(userAssistantMessages) > msg.MaxHistory-1 && msg.MaxHistory > 1 {
		userAssistantMessages = userAssistantMessages[len(userAssistantMessages)-(msg.MaxHistory-1):]
	}

	// Reassemble the message list
	if hasSystem {
		msg.RoleContentList = []RoleContent{systemContent}
		msg.RoleContentList = append(msg.RoleContentList, userAssistantMessages...)
	} else {
		msg.RoleContentList = userAssistantMessages
	}

	return msg
}

func (msg *ChatMessage) GetOpenaiChatMessages() (messageList []openai.ChatCompletionMessage) {
	if msg == nil {
		return
	}
	for _, roleContent := range msg.RoleContentList {
		messageList = append(messageList, openai.ChatCompletionMessage{
			Role:    roleContent.Role,
			Content: roleContent.Content,
		})
	}
	return
}

func (msg *ChatMessage) GetOpenaiChatMessagesWithSystem(system string) (messageList []openai.ChatCompletionMessage) {
	messageList = msg.GetOpenaiChatMessages()
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

func BuildOpenAIChat(endpoint, key, model string, history *ChatMessage, input, system string, temperature float32) (chat func(string, string, *ChatMessage) (string, error), err error) {
	client := GetClient(endpoint, key)
	if client == nil {
		err = fmt.Errorf("build openai client failed")
		return
	}
	chat = func(input, system string, history *ChatMessage) (string, error) {
		if history == nil {
			history = NewDefaultChatMessage()
		}
		history = history.AddSystemContent(system)
		history = history.AddUserContent(input)
		// for _, roleContent := range history.RoleContentList {
		// 	println("llm chat role: ", roleContent.Role, " content: ", roleContent.Content)
		// }
		// println("---------------------------")
		println("length: ", len(history.RoleContentList))
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:       model,
				Messages:    history.GetOpenaiChatMessages(),
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

func ChatIntention(logger *log.Logger, chat func(string, string, *ChatMessage) (string, error), buildSystem func([]opsv1.Pipeline) string, pipelines []opsv1.Pipeline, history *ChatMessage, input string, maxTryTimes int) (output string, pipeline string, err error) {
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
			pipeline = p.Name
			return
		}
	}
	if maxTryTimes > 0 {
		maxTryTimes--
		goto Again
	}
	return
}

func ChatParameters(logger *log.Logger, chat func(string, string, *ChatMessage) (string, error), buildSystem func(opsv1.Pipeline, []opsv1.Cluster) string, pipelines []opsv1.Pipeline, clusters []opsv1.Cluster, history *ChatMessage, pipeline *opsv1.Pipeline, input string, maxTryTimes int) (output string, variables map[string]string, err error) {
Again:
	system := buildSystem(*pipeline, clusters)
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

	variables = make(map[string]string)

	for k, v := range pipeline.Spec.Variables {
		if k == "cluster" {
			v.Enums = clusterEnums
		}
		if v.Value != "" {
			variables[k] = v.Value
		} else if _, ok := outputVars[k]; ok {
			outV := outputVars[k]
			variables[k] = ""
			if len(v.Enums) > 0 {
				found := false
				for _, enum := range v.Enums {
					if outV == enum {
						found = true
						break
					}
				}
				if found {
					variables[k] = outV
				}
			} else {
				variables[k] = outV
			}
		} else {
			variables[k] = ""
		}
	}
	return
}
