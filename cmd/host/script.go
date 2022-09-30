package host

import (
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/spf13/cobra"
)

var scriptOpt host.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return host.ActionScript(scriptOpt)
	},
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptOpt.Username, "username", "", "root", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Username, "password", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Content, "content", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Hosts, "hosts", "", "", "")
	scriptCmd.Flags().BoolVarP(&scriptOpt.Clear, "clear", "", false, "")
}
