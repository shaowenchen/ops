package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var etcHostsOption kube.EtcHostsOption

var etcHostsCmd = &cobra.Command{
	Use:   "etchosts",
	Short: "config /etc/hosts on host",
	RunE: func(cmd *cobra.Command, args []string) error {
		return kube.ActionEtcHostsOnEachNode(etcHostsOption)
	},
}

func init() {
	etcHostsCmd.Flags().StringVarP(&etcHostsOption.Kubeconfig, "kubeconfig", "", "", "")
	etcHostsCmd.Flags().StringVarP(&etcHostsOption.Domain, "domain", "", "", "domain to /etc/hosts (required), e.g., doamin.com")
	etcHostsCmd.MarkFlagRequired("domain")
	etcHostsCmd.Flags().StringVarP(&etcHostsOption.IP, "ip", "", "", "ip to /etc/hosts (required), e.g., 1.1.1.1")
	etcHostsCmd.Flags().BoolVarP(&etcHostsOption.Clear, "clear", "", false, "")
}
