package copilot

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/shaowenchen/ops/pkg/agent"
	"github.com/shaowenchen/ops/pkg/copilot"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/spf13/cobra"
)

var copilotOpt option.CopilotOption
var verbose string

const welcome = `Welcome to Opscli Copilot. Please type "exit" or "q" to quit.`
const quit = "Goodbye!"
const prompt = "Opscli> "
const defaultEndpoint = "https://api.openai.com/v1"
const defaultModel = "gpt-3.5-turbo-16k"

var CopilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "use llm to assist ops",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetFile().SetFlag().Build()
		fillParameters(&copilotOpt)
		copilot.GlobalCopilotOption = &copilotOpt
		CreateCopilot(logger, copilotOpt)
	},
}

func CreateCopilot(logger *log.Logger, opt option.CopilotOption) {
	fmt.Println(welcome)
	defer fmt.Println(quit)
	history := copilot.RoleContentList{}
	stdFd := int(os.Stdin.Fd())
	oldState, _ := term.MakeRaw(stdFd)
	defer term.Restore(stdFd, oldState)

	terminal := term.NewTerminal(os.Stdin, prompt)
	rawState, _ := term.GetState(stdFd)

	pipelinerunsManager := agent.NewLLMPipelineRunsManager(copilotOpt.OpsServer, copilotOpt.OpsToken, "ops-system", copilotOpt.RuntimeImage, 10, copilot.AllPipelines, copilot.AllTasks)
	for {
		input, err := terminal.ReadLine()
		if err != nil {
			printTerm(stdFd, oldState, rawState, err.Error())
			break
		}
		input = strings.TrimSpace(input)
		pr, err := copilot.RunPipeline(logger, &history, pipelinerunsManager, input, 3, "copilot")
		if err != nil {
			printTerm(stdFd, oldState, rawState, err.Error())
			continue
		}
		printTerm(stdFd, oldState, rawState, pr.Output)
	}
}

func fillParameters(opt *option.CopilotOption) {
	if opt.Endpoint == "" {
		opt.Endpoint = utils.GetMultiEnvDefault([]string{"OPENAI_API_HOST", "OPENAI_API_BASE", "endpoint"}, defaultEndpoint)
	}
	if opt.Model == "" {
		opt.Model = utils.GetMultiEnvDefault([]string{"OPENAI_API_MODEL", "model"}, defaultModel)
	}
	if opt.Key == "" {
		opt.Key = utils.GetMultiEnvDefault([]string{"OPENAI_API_KEY", "key"}, "")
	}
	if opt.OpsServer == "" {
		opt.OpsServer = utils.GetMultiEnvDefault([]string{"OPS_SERVER", "opsserver"}, "")
	}
	if opt.OpsToken == "" {
		opt.OpsToken = utils.GetMultiEnvDefault([]string{"OPS_TOKEN", "opstoken"}, "")
	}
}

func printTerm(stdFd int, oldState, rawState *term.State, log string) {
	term.Restore(stdFd, oldState)
	fmt.Println(log)
	term.Restore(stdFd, rawState)
}

func init() {
	CopilotCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Endpoint, "endpoint", "e", "", "e.g. https://api.openai.com/v1")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Model, "model", "m", "", "e.g. gpt-3.5-turbo")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Key, "key", "k", "", "e.g. sk-xxx")
	CopilotCmd.Flags().IntVarP(&copilotOpt.History, "history", "", 5, "")
	CopilotCmd.Flags().BoolVarP(&copilotOpt.Silence, "silence", "s", false, "")
	CopilotCmd.Flags().StringVarP(&copilotOpt.OpsServer, "opsserver", "", "", "")
	CopilotCmd.Flags().StringVarP(&copilotOpt.OpsToken, "opstoken", "", "", "")
	CopilotCmd.Flags().StringVarP(&copilotOpt.RuntimeImage, "runtimeimage", "", constants.DefaultRuntimeImage, "e.g. ubuntu:22.04")
}
