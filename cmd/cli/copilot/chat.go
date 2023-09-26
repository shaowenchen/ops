package copilot

import (
	"encoding/json"
	"fmt"

	"github.com/shaowenchen/ops/pkg/copilot"
	"github.com/shaowenchen/ops/pkg/log"

	opsopt "github.com/shaowenchen/ops/pkg/option"
)

func ChatAsk(logger *log.Logger, opt opsopt.CopilotOption, history *copilot.RoleContentList, input string) (message string, err error) {
	client := copilot.GetClient(opt.Key, opt.Endpoint)
	// keep history length
	history.WithHistory(opt.History)
	resp, err := copilot.ChatCompletetionAsk(logger, client, opt.Model, history, input)
	logger.Debug.Printf("llm chat resp: %v\n", resp)
	return resp, err
}

func ChatCode(logger *log.Logger, opt opsopt.CopilotOption, history *copilot.RoleContentList, input string) (langCode copilot.ChatCodeResponse, err error) {
	client := copilot.GetClient(opt.Key, opt.Endpoint)
	resp, err := copilot.ChatCompletetionCode(logger, client, opt.Model, history, input)
	logger.Debug.Printf("llm code resp: %v\n", resp)
	langCode = copilot.ChatCodeResponse{}
	resp = fmt.Sprintf("[%s]", resp)
	err = json.Unmarshal([]byte(resp), &langCode)
	return langCode, err
}
