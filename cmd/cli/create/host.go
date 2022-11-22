package create

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/create"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "create host resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = CreateHost(logger, createHostOpt, inventory)
		if err != nil {
			logger.Error.Println(err)
			return
		}
	},
}

func CreateHost(logger *log.Logger, option option.CreateHostOption, inventory string) (err error) {
	inventory = utils.GetAbsoluteFilePath(inventory)
	restConfig, err := utils.GetRestConfig(inventory)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if option.PrivateKey == "" {
		option.PrivateKey, err = utils.ReadFile(option.PrivateKeyPath)
		option.PrivateKey = utils.EncodingStringToBase64(option.PrivateKey)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	option.Password = utils.EncodingStringToBase64(option.Password)
	hs := host.GetHosts(logger, option.HostOption, inventory)

	for _, h := range hs {
		// one name, one host
		if len(hs) == 1 {
			hs[0].Name = option.Name
		}
		// no name, multi host
		err = create.CreateHost(logger, restConfig, h, option.Clear)
		if err != nil {
			logger.Error.Println(err)
		}
	}

	return
}

func init() {
	hostCmd.Flags().StringVarP(&createHostOpt.Kubeconfig, "kubeconfig", "", constants.GetCurrentUserKubeConfigPath(), "")
	hostCmd.Flags().StringVarP(&createHostOpt.Namespace, "namespace", "", "default", "")
	hostCmd.Flags().StringVarP(&createHostOpt.Name, "name", "", "", "")
	hostCmd.MarkFlagRequired("name")
	hostCmd.Flags().StringVarP(&createHostOpt.Username, "username", "", "root", "")
	hostCmd.Flags().StringVarP(&createHostOpt.Password, "password", "", "", "")
	hostCmd.Flags().StringVarP(&createHostOpt.PrivateKeyPath, "privatekeypath", "", "~/.ssh/id_rsa", "")
	hostCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")
	hostCmd.MarkFlagRequired("inventory")
	hostCmd.Flags().IntVar(&createHostOpt.Port, "port", 22, "")
	hostCmd.Flags().BoolVarP(&createHostOpt.Clear, "clear", "", false, "")
}
