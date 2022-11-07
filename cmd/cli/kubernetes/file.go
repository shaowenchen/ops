package kubernetes

import (
	"fmt"

	"github.com/shaowenchen/ops/pkg/action"
	"github.com/shaowenchen/ops/pkg/kubernetes"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/spf13/cobra"
)

var fileOption kubernetes.FileOption

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "transfer file between local and remote file in container image",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewDefaultLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		action.KubernetesFile(logger, fileOption)
	},
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
