package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	opscopilot "github.com/shaowenchen/ops/pkg/copilot"
	opslog "github.com/shaowenchen/ops/pkg/log"
)

func PostCopilotPlain(c *gin.Context) {
	// 1. Get input from body
	readBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, "failed to read request body: "+err.Error())
		return
	}
	input := string(readBody)

	// 2. Forward input to PostCopilot
	builder := struct {
		Input string `json:"input"`
	}{
		Input: input,
	}
	builderStr, err := json.Marshal(builder)
	if err != nil {
		showError(c, "marshal error: "+err.Error())
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(builderStr))
	PostCopilot(c)
}

func PostCopilot(c *gin.Context) {
	type Params struct {
		Input string `json:"input"`
	}
	var req = Params{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		showError(c, "get body error "+err.Error())
		return
	}
	if req.Input == "" {
		showError(c, "input is required")
		return
	}
	input := req.Input
	// build chat
	history := opscopilot.NewDefaultChatMessage()
	chat, err := opscopilot.BuildOpenAIChat(GlobalConfig.Copilot.Endpoint, GlobalConfig.Copilot.Key, GlobalConfig.Copilot.Model, history, "", "copilot", 0.1)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// init pr manager
	logger := opslog.NewLogger().SetVerbose("debug").SetStd().SetFlag().Build()
	pipelinerunsManager, _ := opscopilot.NewPipelineRunsManager(GlobalConfig.Copilot.OpsServer, GlobalConfig.Copilot.OpsToken, "ops-system")
	prHistory := opscopilot.NewDefaultChatMessage()
	pr, exitCode, err := opscopilot.RunPipeline(logger, false, chat, prHistory, pipelinerunsManager, input, nil)

	var output string
	if exitCode == opscopilot.ExitCodeIntentionEmpty {
		output = "I can not understand your input:" + input + ", please help me to solve it, use following intention:\n " + pipelinerunsManager.GetForIntent()
	} else if exitCode == opscopilot.ExitCodeParametersNotFound {
		output = "I can not get the parameters, please help me to solve it:\n " + pipelinerunsManager.GetForVariables(pr)
	} else {
		output = fmt.Sprintf("%s", pipelinerunsManager.PrintMarkdownPipelineRuns(pr))
	}
	if output == "" {
		output = "It's bug, please contact chenshaowen to fix it"
	}
	if err != nil {
		output = err.Error()
	}
	showData(c, output)
}
