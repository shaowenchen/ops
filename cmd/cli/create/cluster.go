package create

import (
	"fmt"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "create cluster resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()

		err := CreateCluster(logger, clusterOpt, inventory)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
	},
}

func CreateCluster(logger *log.Logger, clusterOpt option.ClusterOption, inventory string) (err error) {
	clusterOpt.Kubeconfig = utils.GetAbsoluteFilePath(clusterOpt.Kubeconfig)
	restConfig, err := utils.GetRestConfig(clusterOpt.Kubeconfig)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	config, err := utils.ReadFile(utils.GetAbsoluteFilePath(inventory))
	if err != nil {
		return
	}
	if clusterSpec.Server == "" {
		clusterSpec.Server, _ = utils.GetServerUrl(inventory)
	}
	clusterSpec.Config = utils.EncodingStringToBase64(config)
	cluster := opsv1.NewCluster(clusterOpt.Namespace, clusterOpt.Name, clusterSpec.Server, clusterSpec.Config, clusterSpec.Token)
	err = kube.CreateCluster(logger, restConfig, cluster, clusterOpt.Clear)
	if err != nil {
		logger.Error.Println(err)
	}
	return
}

func init() {
	clusterCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")
	clusterCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
	clusterCmd.Flags().StringVarP(&clusterOpt.Namespace, "namespace", "", "ops-system", "")
	clusterCmd.Flags().StringVarP(&clusterOpt.Name, "name", "", "", "")
	clusterCmd.MarkFlagRequired("name")
	clusterCmd.Flags().StringVarP(&clusterOpt.Kubeconfig, "kubeconfig", "", constants.GetCurrentUserKubeConfigPath(), "")
	clusterCmd.Flags().BoolVarP(&clusterOpt.Clear, "clear", "", false, "")
}
