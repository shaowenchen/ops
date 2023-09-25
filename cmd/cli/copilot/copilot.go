package copilot

import (
	"bufio"
	"fmt"

	"os"
	"strings"

	"github.com/shaowenchen/ops/pkg/copilot"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var copilotOpt option.CopilotOption
var verbose string

const defaultEndpoint = "https://api.openai.com/v1"
const defaultModel = "gpt-3.5-turbo"

var CopilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "use llm to assist ops",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()
		fillParameters(&copilotOpt)
		err := CreateCopilot(logger, copilotOpt)
		if err != nil {
			logger.Error.Println(err.Error())
			return
		}
	},
}

func CreateCopilot(logger *log.Logger, opt option.CopilotOption) (err error) {
	// Create History List
	askHistory := copilot.RoleContentList{}
	// Start Chat
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		ask, _ := reader.ReadString('\n')
		ask = strings.TrimSpace(ask)

		if ask == "exit" || ask == "q" {
			break
		}
		if ask == "" {
			continue
		}
		reply, err := ChatAsk(logger, opt, &askHistory, ask)
		if err != nil {
			logger.Error.Printf("Chat error: %v\n", err)
			continue
		}
		askHistory.AddChatPairContent(ask, reply).WithHistory(opt.History)
		if copilot.IsCanBeSolvedWithCode(reply) {
			codeHistory := copilot.RoleContentList{}
			codeHistory.Merge(&askHistory)
			langcode, err := ChatCode(logger, opt, &codeHistory, reply)
			if err != nil {
				logger.Error.Printf("Chat error: %v\n", err)
				continue
			}
			needRun := false
			if opt.Silence {
				needRun = true
			} else {
				logger.Info.Println(langcode.Content)
				logger.Info.Printf("Would you like to run this code? (y/n)\n%s\n", langcode.Code)
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(confirm)
				if confirm == "y" {
					needRun = true
				} else if confirm == "n" {
					continue
				}
			}
			if needRun {
				hc, err := host.NewHostConnBase64(nil)
				if err != nil {
					logger.Error.Println(err)
					break
				}
				stdout, err := hc.ExecWithExecutor(false, strings.ToLower(langcode.Language), "-c", langcode.Code)
				if err != nil {
					logger.Error.Println(err)
					break
				}
				codeHistory.AddChatPairContent(langcode.Content, "Code execution returns: "+stdout)
				logger.Info.Println(stdout)
			}
		} else {
			logger.Info.Println(reply)
		}
	}
	logger.Info.Println("Bye!")
	return
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
	CopilotCmd.Flags().StringVarP(&copilotOpt.Endpoint, "endpoint", "", "", "e.g. https://api.openai.com/v1")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Model, "model", "", "", "e.g. gpt-3.5-turbo")
	CopilotCmd.Flags().StringVarP(&copilotOpt.Key, "key", "", "", "e.g. sk-xxx")
	CopilotCmd.Flags().IntVarP(&copilotOpt.History, "history", "", 5, "")
	CopilotCmd.Flags().BoolVarP(&copilotOpt.Silence, "silence", "s", false, "")
}
