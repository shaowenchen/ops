package kube

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var fileOption kube.FileOption

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "transfer file between local and remote file in container image",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		File(logger, fileOption)
	},
}

func File(logger *log.Logger, option kube.FileOption) (err error) {
	client, err := utils.NewKubernetesClient(option.Kubeconfig)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	nodeList, err := kube.GetNodes(logger, client, option.KubeOption)
	if err != nil {
		logger.Error.Println(err)
	}
	if len(nodeList) == 0 {
		logger.Info.Println("Please provide a node at least")
	}
	for _, node := range nodeList {
		kube.File(logger, client, node, option)
	}
	return
}

func init() {
	fileCmd.Flags().StringVarP(&fileOption.Kubeconfig, "kubeconfig", "", "", "")
	fileCmd.Flags().StringVarP(&fileOption.NodeName, "nodename", "", "", "")
	fileCmd.Flags().StringVarP(&fileOption.Image, "image", "", "", "")
	fileCmd.MarkFlagRequired("image")
	fileCmd.Flags().StringVarP(&fileOption.LocalFile, "localfile", "", "", "")
	fileCmd.MarkFlagRequired("localfile")
	fileCmd.Flags().StringVarP(&fileOption.RemoteFile, "remotefile", "", "", "")
	fileCmd.MarkFlagRequired("remotefile")
	fileCmd.Flags().BoolVarP(&fileOption.All, "all", "", false, "")
}
