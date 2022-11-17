package kube

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var scriptOption kube.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		Script(logger, scriptOption)
	},
}

func Script(logger *log.Logger, option kube.ScriptOption) (err error) {
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
		kube.Script(logger, client, node, option)
	}
	return
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptOption.Kubeconfig, "kubeconfig", "", "~/.kube/config", "")
	scriptCmd.Flags().StringVarP(&scriptOption.NodeName, "nodename", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOption.Image, "image", "", "docker.io/library/alpine:latest", "runtime image")
	scriptCmd.Flags().StringVarP(&scriptOption.Content, "content", "", "", "")
	scriptCmd.MarkFlagRequired("content")
	scriptCmd.Flags().BoolVarP(&scriptOption.All, "all", "", false, "")
}
