package host

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var scriptOpt host.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		scriptOpt.Password = utils.EncodingStringToBase64(scriptOpt.Password)
		privateKey, _ := utils.ReadFile(scriptOpt.PrivateKeyPath)
		scriptOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		Script(logger, scriptOpt)
	},
}

func Script(logger *log.Logger, option host.ScriptOption) (err error) {
	for _, h := range host.GetHosts(logger, option.HostOption) {
		err = host.Script(logger, h, option)
		if err != nil {
			logger.Error.Println(err)
		} else {
			logger.Info.Println("Successed!")
		}
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
