package kube

import (
	"fmt"

	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/shaowenchen/opscli/pkg/log"
	"github.com/spf13/cobra"
)

var scriptOption kube.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewDefaultLogger(true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		kube.ActionScript(logger, scriptOption)
	},
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptOption.Kubeconfig, "kubeconfig", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOption.NodeName, "nodename", "", "", "")
	scriptCmd.Flags().StringVarP(&scriptOption.Image, "image", "", "docker.io/library/alpine:latest", "")
	scriptCmd.Flags().StringVarP(&scriptOption.Content, "content", "", "", "")
	scriptCmd.MarkFlagRequired("content")
	scriptCmd.Flags().BoolVarP(&scriptOption.All, "all", "", false, "")
}
