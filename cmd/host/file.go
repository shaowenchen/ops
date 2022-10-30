package host

import (
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/spf13/cobra"
)

var fileOpt host.FileOption

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "transfer file between local and remeote",
	Run: func(cmd *cobra.Command, args []string) {
		host.ActionFile(fileOpt)
	},
}

func init() {
	fileCmd.Flags().StringVarP(&fileOpt.LocalFile, "localfile", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.RemoteFile, "remotefile", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Username, "username", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Password, "password", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Hosts, "hosts", "", "", "")
	fileCmd.Flags().IntVar(&fileOpt.Port, "port", 22, "")
}
