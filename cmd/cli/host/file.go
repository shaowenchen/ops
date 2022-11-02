package host

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/spf13/cobra"
)

var fileOpt host.FileOption

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "transfer between local and remote file",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewDefaultLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		host.ActionFile(logger, fileOpt)
	},
}

func init() {
	fileCmd.Flags().StringVarP(&fileOpt.LocalFile, "localfile", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.RemoteFile, "remotefile", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Direction, "direction", "d", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Username, "username", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Password, "password", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Hosts, "hosts", "", "", "")
	fileCmd.Flags().IntVar(&fileOpt.Port, "port", 22, "")
}
