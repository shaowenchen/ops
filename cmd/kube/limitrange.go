package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var limitRangeOption kube.LimitRangeOption

var limitRangeCmd = &cobra.Command{
	Use:   "limitrange",
	Short: "config limitRange for kubernetes",
	RunE: func(cmd *cobra.Command, args []string)(err error) {
		return kube.ActionLimitRange(limitRangeOption)
	},
}

func init() {
	limitRangeCmd.Flags().StringVarP(&limitRangeOption.Kubeconfig, "kubeconfig", "", "", "")
	limitRangeCmd.Flags().StringVarP(&limitRangeOption.Name, "name", "", "", "default/limit-range")
	limitRangeCmd.MarkFlagRequired("name")
	limitRangeCmd.Flags().StringVarP(&limitRangeOption.ReqMem, "reqmem", "", "", "500Mi")
	limitRangeCmd.Flags().StringVarP(&limitRangeOption.LimitMem, "limitmem", "", "", "2Gi")
	limitRangeCmd.Flags().StringVarP(&limitRangeOption.ReqCPU, "reqcpu", "", "", "0.1")
	limitRangeCmd.Flags().StringVarP(&limitRangeOption.LimitCPU, "limitcpu", "", "", "1")
	limitRangeCmd.Flags().BoolVarP(&limitRangeOption.Clear, "clear", "", false, "")
	limitRangeCmd.Flags().BoolVarP(&limitRangeOption.All, "all", "", false, "")
}
