package create

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/create"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var hostOption create.HostOption

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "create host resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = CreateHost(logger, hostOption)
		if err != nil {
			logger.Error.Println(err)
			return
		}
	},
}

func CreateHost(logger *log.Logger, option create.HostOption) (err error) {
	option.Kubeconfig = utils.GetAbsoluteFilePath(option.Kubeconfig)
	restConfig, err := utils.GetRestConfig(option.Kubeconfig)
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
	// one name, one host
	// no name, multi host
	for _, h := range host.GetHosts(logger, option.HostOption) {
		err = create.CreateHost(logger, restConfig, h, option.Clear)
		if err != nil {
			logger.Error.Println(err)
		}
	}

	return
}

func init() {
	hostCmd.Flags().StringVarP(&hostOption.Kubeconfig, "kubeconfig", "", "~/.kube/config", "")
	hostCmd.Flags().StringVarP(&hostOption.Namespace, "namespace", "", "default", "")
	hostCmd.Flags().StringVarP(&hostOption.Name, "name", "", "", "")
	hostCmd.MarkFlagRequired("name")
	hostCmd.Flags().StringVarP(&hostOption.Username, "username", "", "root", "")
	hostCmd.Flags().StringVarP(&hostOption.Password, "password", "", "", "")
	hostCmd.Flags().StringVarP(&hostOption.PrivateKeyPath, "privatekeypath", "", "~/.ssh/id_rsa", "")
	hostCmd.Flags().StringVarP(&hostOption.Hosts, "hosts", "", "", "")
	hostCmd.MarkFlagRequired("hosts")
	hostCmd.Flags().IntVar(&hostOption.Port, "port", 22, "")
	hostCmd.Flags().BoolVarP(&hostOption.Clear, "clear", "", false, "")
}
