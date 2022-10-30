package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var scriptOption kube.ScriptOption

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		kube.ActionScript(scriptOption)
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
