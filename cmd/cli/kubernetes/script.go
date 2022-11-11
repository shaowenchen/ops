package kubernetes

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/kubernetes"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/spf13/cobra"
)

var scriptOption kubernetes.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewDefaultLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		Script(logger, scriptOption)
	},
}

func Script(logger *log.Logger, option kubernetes.ScriptOption) (err error) {
	client, nodeList, err := kubernetes.GetClientAndNodes(logger, option.KubeOption)
	if err != nil {
		logger.Error.Println(err)
	}
	for _, node := range nodeList {
		kubernetes.Script(logger, client, node, option)
	}
	return
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptOption.Kubeconfig, "kubeconfig", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOption.NodeName, "nodename", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOption.Image, "image", "", "docker.io/library/alpine:latest", "")
	scriptCmd.Flags().StringVarP(&scriptOption.Content, "content", "", "", "")
	scriptCmd.MarkFlagRequired("content")
	scriptCmd.Flags().BoolVarP(&scriptOption.All, "all", "", false, "")
}
