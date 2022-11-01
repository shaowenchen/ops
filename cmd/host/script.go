package host

import (
	"fmt"

	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/shaowenchen/opscli/pkg/log"
	"github.com/spf13/cobra"
)

var scriptOpt host.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewDefaultLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		host.ActionBatchScript(logger, scriptOpt)
	},
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptOpt.Content, "content", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Username, "username", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Password, "password", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Hosts, "hosts", "", "", "")
	scriptCmd.Flags().IntVar(&scriptOpt.Port, "port", 22, "")
}
