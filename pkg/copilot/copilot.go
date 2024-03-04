package copilot

import (
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

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
