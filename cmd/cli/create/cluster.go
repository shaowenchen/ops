package create

import (
	"context"
	"fmt"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var cClusterOpt option.ClusterOption
var cInventory string
var cVerbose string

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "create cluster resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(cVerbose).SetStd().SetFile().Build()
		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultShellTimeoutDuration)
		defer cancel()
		err := CreateCluster(ctx, logger, cClusterOpt, cInventory)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
	},
}

func CreateCluster(ctx context.Context, logger *log.Logger, clusterOpt option.ClusterOption, inventory string) (err error) {
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
	var cClusterSpec opsv1.ClusterSpec
	if cClusterSpec.Server == "" {
		cClusterSpec.Server, _ = utils.GetServerUrl(inventory)
	}
	cClusterSpec.Config = utils.EncodingStringToBase64(config)
	cluster := opsv1.NewCluster(clusterOpt.Namespace, clusterOpt.Name, cClusterSpec.Server, cClusterSpec.Config, cClusterSpec.Token)
	err = kube.CreateCluster(ctx, logger, restConfig, cluster, clusterOpt.Clear)
	if err != nil {
		logger.Error.Println(err)
	}
	return
}

func init() {
	clusterCmd.Flags().StringVarP(&cInventory, "inventory", "i", "", "")
	clusterCmd.Flags().StringVarP(&cVerbose, "verbose", "v", "", "")
	clusterCmd.Flags().StringVarP(&cClusterOpt.Namespace, "namespace", "", constants.DefaultOpsNamespace, "")
	clusterCmd.Flags().StringVarP(&cClusterOpt.Name, "name", "", "", "")
	clusterCmd.MarkFlagRequired("name")
	clusterCmd.Flags().StringVarP(&cClusterOpt.Kubeconfig, "kubeconfig", "", constants.GetCurrentUserKubeConfigPath(), "")
	clusterCmd.Flags().BoolVarP(&cClusterOpt.Clear, "clear", "", false, "")
}
