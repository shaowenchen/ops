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
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		err = CreateCluster(logger, createClusterOpt, inventory)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
	},
}

func CreateCluster(logger *log.Logger, clusterOpt option.CreateClusterOption, inventory string) (err error) {
	inventory = utils.GetAbsoluteFilePath(inventory)
	restConfig, err := utils.GetRestConfig(inventory)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	config, err := utils.ReadFile(clusterOpt.Config)
	if err != nil {
		return
	}
	if clusterOpt.ClusterSpec.Server == "" {
		clusterOpt.Server, _ = utils.GetServerUrl(clusterOpt.Config)
	}
	clusterOpt.Config = utils.EncodingStringToBase64(config)
	cluster := opsv1.NewCluster(clusterOpt.Namespace, clusterOpt.Name, clusterOpt.Server, clusterOpt.Config, clusterOpt.Token)
	err = kube.CreateCluster(logger, restConfig, cluster, clusterOpt.Clear)
	if err != nil {
		logger.Error.Println(err)
	}
	return
}

func init() {
	clusterCmd.Flags().StringVarP(&inventory, "invenroty", "i", "", "")
	clusterCmd.Flags().StringVarP(&createHostOpt.Namespace, "namespace", "", "default", "")
	clusterCmd.Flags().StringVarP(&createHostOpt.Name, "name", "", "", "")
	clusterCmd.MarkFlagRequired("name")
	clusterCmd.Flags().StringVarP(&createHostOpt.Kubeconfig, "kubeconfig", "", constants.GetCurrentUserKubeConfigPath(), "")
	clusterCmd.Flags().BoolVarP(&createHostOpt.Clear, "clear", "", false, "")
}
