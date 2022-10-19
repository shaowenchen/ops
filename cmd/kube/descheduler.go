package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var deschedulerOption kube.DeschedulerOption

var deschedulerCmd = &cobra.Command{
	Use:   "descheduler",
	Short: "descheduler resource",
	Run: func(cmd *cobra.Command, args []string) {
		kube.ActionDescheduler(deschedulerOption)
	},
}

func init() {
	deschedulerCmd.Flags().StringVarP(&deschedulerOption.Kubeconfig, "kubeconfig", "", "", "")
	deschedulerCmd.Flags().StringVarP(&deschedulerOption.Namespace, "namespace", "", "default", "")
	deschedulerCmd.Flags().BoolVarP(&deschedulerOption.RemoveDuplicates, "removeduplicates", "d", true, "")
	deschedulerCmd.Flags().BoolVarP(&deschedulerOption.NodeUtilization, "nodeutilization", "n", true, "")
	deschedulerCmd.Flags().Int16VarP(&deschedulerOption.HighPercent, "highpercent", "", 80, "")
	deschedulerCmd.Flags().BoolVarP(&deschedulerOption.All, "all", "", false, "")
}
