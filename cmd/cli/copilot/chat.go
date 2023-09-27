package copilot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shaowenchen/ops/pkg/copilot"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"golang.org/x/term"

	opsopt "github.com/shaowenchen/ops/pkg/option"
)

const ReplyEmpty = "reply is empty"
const ReplyHasNotStarted = "has not started"
const ReplyNotAuthorized = "not authorized"

func chatAsk(logger *log.Logger, opt opsopt.CopilotOption, history *copilot.RoleContentList, input string) (message string, err error) {
	client := copilot.GetClient(opt.Key, opt.Endpoint)
	// keep history length
	history.WithHistory(opt.History)
	resp, err := copilot.ChatCompletetionAsk(logger, client, opt.Model, history, input)
	logger.Debug.Printf("llm chat resp: %v\n", resp)
	return resp, err
}

func chatCode(logger *log.Logger, opt opsopt.CopilotOption, history *copilot.RoleContentList, input string) (langCode copilot.ChatCodeResponse, err error) {
	client := copilot.GetClient(opt.Key, opt.Endpoint)
	resp, err := copilot.ChatCompletetionCode(logger, client, opt.Model, history, input)
	logger.Debug.Printf("llm code resp: %v\n", resp)
	langCode = copilot.ChatCodeResponse{}
	err = json.Unmarshal([]byte(resp), &langCode)
	if err != nil {
		err = fmt.Errorf("llm code resp unmarshal error: %v\n", resp)
		return
	}
	return langCode, err
}

func PrintTerm(stdFd int, oldState, rawState *term.State, log string) {
	term.Restore(stdFd, oldState)
	fmt.Println(log)
	term.Restore(stdFd, rawState)
}

func ChatRecursion(logger *log.Logger, opt opsopt.CopilotOption, maxTryTimes int, history *copilot.RoleContentList, ask, preReply string, stdFd int, oldState, rawState *term.State, terminal, confirmTerminal *term.Terminal) (reply string, err error) {
	if maxTryTimes == 0 {
		reply = ReplyEmpty
		return
	}
	// start first chat or summary reply after run code
	if preReply == ReplyHasNotStarted || history.IsEndWithRunCodePair() {
		reply, err = chatAsk(logger, opt, history, ask)
		if err != nil {
			return
		}
		return ChatRecursion(logger, opt, maxTryTimes-1, history, ask, reply, stdFd, oldState, rawState, terminal, confirmTerminal)
	}
	// don't need to run code
	if !copilot.IsNeedToRunCode(preReply) {
		history.AddChatPairContent(ask, preReply)
		reply = preReply
		return
	}

	preReply = copilot.RemoveExistedRunableCode(preReply)
	// ready to run code
	codeHistory := copilot.RoleContentList{}
	codeHistory.Merge(history)
	langcode, err := chatCode(logger, opt, &codeHistory, preReply)
	if err != nil {
		return
	}
	// check if authorized
	authorized := false
	if opt.Silence {
		authorized = true
	} else {
		PrintTerm(stdFd, oldState, rawState, langcode.Content)
		PrintTerm(stdFd, oldState, rawState, fmt.Sprintf("%s\nCan I run this code? (y/n)", utils.CodeBlock(langcode.Code)))
		confirm, _ := confirmTerminal.ReadLine()
		confirm = strings.TrimSpace(confirm)
		if confirm == "y" {
			authorized = true
		}
	}
	// run code
	if authorized {
		hc, _ := host.NewHostConnBase64(nil)
		reply, err = hc.ExecWithExecutor(false, strings.ToLower(langcode.Language), "-c", langcode.Code)
		codeHistory.AddRunCodePairContent(langcode.Content, reply)

	} else {
		reply = ReplyNotAuthorized
		return
	}
	if err != nil {
		PrintTerm(stdFd, oldState, rawState, "okk")
		PrintTerm(stdFd, oldState, rawState, err.Error())
	}
	return ChatRecursion(logger, opt, maxTryTimes-1, &codeHistory, ask, reply, stdFd, oldState, rawState, terminal, confirmTerminal)
}
