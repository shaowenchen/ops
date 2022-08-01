package host

import (
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/spf13/cobra"
)

var etcHostsOpt host.EtcHostsOption

var etcHostsCmd = &cobra.Command{
	Use:   "etchosts",
	Short: "config /etc/hosts on host",
	RunE: func(cmd *cobra.Command, args []string) error {
		return host.ActionEtcHosts(etcHostsOpt)
	},
}

func init() {
	etcHostsCmd.Flags().StringVarP(&etcHostsOpt.Username, "username", "", "", "")
	etcHostsCmd.Flags().StringVarP(&etcHostsOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	etcHostsCmd.Flags().StringVarP(&etcHostsOpt.Domain, "domain", "", "", "")
	etcHostsCmd.Flags().StringVarP(&etcHostsOpt.IP, "ip", "", "", "")
	etcHostsCmd.Flags().StringVarP(&etcHostsOpt.Input, "input", "i", "", "resource list")
	etcHostsCmd.Flags().BoolVarP(&etcHostsOpt.Clear, "clear", "", false, "")
}
