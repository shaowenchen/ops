package kube

import (
	"github.com/spf13/cobra"
)

var KubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "command about Kubernetes",
}

func init() {
	KubeCmd.AddCommand(scriptCmd)
	KubeCmd.AddCommand(fileCmd)
}
