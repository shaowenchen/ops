package host

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
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
		Script(logger, scriptOpt)
	},
}

func Script(logger *log.Logger, option host.ScriptOption) {
	for _, h := range host.GetHosts(logger, option.HostOption) {
		host.Script(logger, h, option)
	}
	return
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptOpt.Content, "content", "", "", "")
	scriptCmd.Flags().BoolVarP(&scriptOpt.Sudo, "sudo", "", false, "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Username, "username", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Password, "password", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOpt.Hosts, "hosts", "", "", "")
	scriptCmd.Flags().IntVar(&scriptOpt.Port, "port", 22, "")
}
