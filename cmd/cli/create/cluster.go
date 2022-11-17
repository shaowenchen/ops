package create

import (
	"fmt"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/create"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var clusterOption create.ClusterOption

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "create cluster resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		CreateCluster(logger, clusterOption)
	},
}

func CreateCluster(logger *log.Logger, option create.ClusterOption) (err error) {
	option.Kubeconfig = utils.GetAbsoluteFilePath(option.Kubeconfig)
	restConfig, err := utils.GetRestConfig(option.Kubeconfig)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	config, err := utils.ReadFile(option.Config)
	if err != nil {
		return
	}
	option.Config = utils.EncodingStringToBase64(config)
	cluster := opsv1.NewCluster(option.Namespace, option.Name, option.Server, option.Config, option.Token)
	err = create.CreateCluster(logger, restConfig, cluster, option.Clear)
	if err != nil {
		logger.Error.Println(err)
	}
	return
}

func init() {
	clusterCmd.Flags().StringVarP(&clusterOption.Kubeconfig, "kubeconfig", "", "~/.kube/config", "")
	clusterCmd.Flags().StringVarP(&clusterOption.Namespace, "namespace", "", "default", "")
	clusterCmd.Flags().StringVarP(&clusterOption.Name, "name", "", "", "")
	clusterCmd.MarkFlagRequired("name")
	clusterCmd.Flags().StringVarP(&clusterOption.Config, "config", "", "", "")
	clusterCmd.MarkFlagRequired("config")
	clusterCmd.Flags().BoolVarP(&clusterOption.Clear, "clear", "", false, "")
}
