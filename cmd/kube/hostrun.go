package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var hostRunOption kube.HostRunOption

var hostRunCmd = &cobra.Command{
	Use:   "hostrun",
	Short: "run script on host",
	RunE: func(cmd *cobra.Command, args []string) error {
		return kube.ActionHostRun(hostRunOption)
	},
}

func init() {
	hostRunCmd.Flags().StringVarP(&hostRunOption.Kubeconfig, "kubeconfig", "", "", "")
	hostRunCmd.Flags().StringVarP(&hostRunOption.NodeName, "nodename", "", "", "")
	hostRunCmd.Flags().StringVarP(&hostRunOption.Script, "script", "", "", "")
	hostRunCmd.MarkFlagRequired("script")
	hostRunCmd.Flags().BoolVarP(&hostRunOption.All, "all", "", false, "")
}
