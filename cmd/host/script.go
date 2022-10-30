package host

import (
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/spf13/cobra"
)

var scriptOpt host.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		host.ActionBatchScript(scriptOpt)
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
