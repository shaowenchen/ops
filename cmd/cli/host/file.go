package host

import (
	"fmt"

	"errors"

	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var fileOpt host.FileOption

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "transfer between local and remote file",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		fileOpt.Password = utils.EncodingStringToBase64(fileOpt.Password)
		privateKey, _ := utils.ReadFile(fileOpt.PrivateKeyPath)
		fileOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		File(logger, fileOpt)
	},
}

func File(logger *log.Logger, option host.FileOption) (err error) {
	hs := host.GetHosts(logger, option.HostOption)
	if utils.IsDownloadDirection(option.Direction) && len(hs) != 1 {
		errMsg := "need only one host while downloading"
		logger.Error.Println(errMsg)
		return errors.New(errMsg)
	}
	for _, h := range hs {
		err = host.File(logger, h, option)
		if err != nil {
			logger.Error.Println(err)
		} else {
			logger.Info.Println("Successed!")
		}
	}
	return
}

func init() {
	fileCmd.Flags().BoolVarP(&fileOpt.Sudo, "sudo", "", false, "")
	fileCmd.Flags().StringVarP(&fileOpt.LocalFile, "localfile", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.RemoteFile, "remotefile", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Direction, "direction", "d", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Username, "username", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Password, "password", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	fileCmd.Flags().StringVarP(&fileOpt.Hosts, "hosts", "", "", "")
	fileCmd.Flags().IntVar(&fileOpt.Port, "port", 22, "")
}
