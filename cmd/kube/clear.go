package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var clearOption kube.ClearOption

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return kube.ActionClear(clearOption)
	},
}

func init() {
	clearCmd.Flags().StringVarP(&clearOption.Kubeconfig, "kubeconfig", "", "", "")
	clearCmd.Flags().StringVarP(&clearOption.Namespace, "namespace", "", "default", "")
	clearCmd.Flags().StringVarP(&clearOption.Status, "status", "", "Failed,Unavailable", "")
	clearCmd.Flags().BoolVarP(&clearOption.All, "all", "", false, "")
}
