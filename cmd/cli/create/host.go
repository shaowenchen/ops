package create

import (
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "create host resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()
		err := CreateHost(logger, clusterOpt, hostOpt, inventory)
		if err != nil {
			logger.Error.Println(err)
			return
		}
	},
}

func CreateHost(logger *log.Logger, clusterOpt option.ClusterOption, hostOpt option.HostOption, inventory string) (err error) {
	kubeconfigPath := utils.GetAbsoluteFilePath(clusterOpt.Kubeconfig)
	restConfig, err := utils.GetRestConfig(kubeconfigPath)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if hostOpt.PrivateKey == "" {
		hostOpt.PrivateKey, err = utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(hostOpt.PrivateKey)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
	hs := host.GetHosts(logger, hostOpt, inventory)

	for _, h := range hs {
		h.Namespace = clusterOpt.Namespace
		// one name, one host
		if len(hs) == 1 {
			hs[0].Name = clusterOpt.Name
		}
		// no name, multi host
		err = kube.CreateHost(logger, restConfig, h, clusterOpt.Clear)
		if err != nil {
			logger.Error.Println(err)
		}
	}

	return
}

func init() {
	hostCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
	hostCmd.Flags().StringVarP(&clusterOpt.Kubeconfig, "kubeconfig", "", constants.GetCurrentUserKubeConfigPath(), "")
	hostCmd.Flags().StringVarP(&clusterOpt.Namespace, "namespace", "", "ops-system", "")
	hostCmd.Flags().StringVarP(&clusterOpt.Name, "name", "", "", "")
	hostCmd.MarkFlagRequired("name")
	hostCmd.Flags().BoolVarP(&clusterOpt.Clear, "clear", "", false, "")

	hostCmd.Flags().StringVarP(&hostOpt.Username, "username", "", "root", "")
	hostCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	hostCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", "~/.ssh/id_rsa", "")
	hostCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")
	hostCmd.MarkFlagRequired("inventory")
	hostCmd.Flags().IntVar(&hostOpt.Port, "port", 22, "")
}
