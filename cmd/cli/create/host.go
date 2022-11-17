package create

import (
	"fmt"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/create"
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
			fmt.Printf(err.Error())
			return
		}
		CreateHost(logger, hostOption)
	},
}

func CreateHost(logger *log.Logger, option create.HostOption) (err error) {
	option.Kubeconfig = utils.GetAbsoluteFilePath(option.Kubeconfig)
	restConfig, err := utils.GetRestConfig(option.Kubeconfig)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	host := opsv1.NewHost("", option.Name, option.Address, option.Port, option.Username, option.Password, option.PrivateKey, option.PrivateKeyPath)
	create.CreateHost(logger, restConfig, host, option.Clear)
	return
}

func init() {
	hostCmd.Flags().StringVarP(&hostOption.Kubeconfig, "kubeconfig", "", "~/.kube/config", "")
	hostCmd.Flags().StringVarP(&hostOption.Namespace, "namespace", "", "default", "")
	hostCmd.Flags().StringVarP(&hostOption.Name, "name", "", "", "")
	hostCmd.MarkFlagRequired("name")
	hostCmd.Flags().StringVarP(&hostOption.Username, "username", "", "", "")
	hostCmd.Flags().StringVarP(&hostOption.Password, "password", "", "", "")
	hostCmd.Flags().StringVarP(&hostOption.PrivateKeyPath, "privatekeypath", "", "", "")
	hostCmd.Flags().StringVarP(&hostOption.Hosts, "hosts", "", "", "")
	hostCmd.MarkFlagRequired("hosts")
	hostCmd.Flags().IntVar(&hostOption.Port, "port", 22, "")
	hostCmd.Flags().BoolVarP(&hostOption.Clear, "clear", "", false, "")
}
