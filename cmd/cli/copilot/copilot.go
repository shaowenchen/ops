package copilot

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"

	"github.com/shaowenchen/ops/pkg/copilot"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var copilotOpt option.CopilotOption
var verbose string

const welcome = `Welcome to Opscli Copilot. Please type "exit" or "q" to quit.`
const quit = "Goodbye!"
const prompt = "Opscli> "
const maxTryTimes = 5
const defaultEndpoint = "https://api.openai.com/v1"
const defaultModel = "gpt-3.5-turbo"

var CopilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "use llm to assist ops",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetFile().Build()
		fillParameters(&copilotOpt)
		CreateCopilot(logger, copilotOpt)
	},
}

func CreateCopilot(logger *log.Logger, opt option.CopilotOption) {
	fmt.Println(welcome)
	askHistory := copilot.RoleContentList{}
	stdFd := int(os.Stdin.Fd())
	oldState, _ := term.MakeRaw(stdFd)
	defer term.Restore(stdFd, oldState)

	terminal := term.NewTerminal(os.Stdin, prompt)
	confirmTerminal := term.NewTerminal(os.Stdin, "> ")
	rawState, _ := term.GetState(stdFd)
	for {
		ask, err := terminal.ReadLine()
		if err != nil {
			PrintTerm(stdFd, oldState, rawState, err.Error())
			break
		}
		ask = strings.TrimSpace(ask)
		reply, err := ChatRecursion(logger, opt, maxTryTimes, &askHistory, ask, ReplyHasNotStarted, stdFd, oldState, rawState, terminal, confirmTerminal)
		if err != nil {
			PrintTerm(stdFd, oldState, rawState, err.Error())
			continue
		}
		PrintTerm(stdFd, oldState, rawState, reply)
	}
	fmt.Println(quit)
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
}

func init() {
	CopilotCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Endpoint, "endpoint", "e", "", "e.g. https://api.openai.com/v1")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Model, "model", "m", "", "e.g. gpt-3.5-turbo")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Key, "key", "k", "", "e.g. sk-xxx")
	CopilotCmd.Flags().IntVarP(&copilotOpt.History, "history", "", 5, "")
	CopilotCmd.Flags().BoolVarP(&copilotOpt.Silence, "silence", "s", false, "")
}
