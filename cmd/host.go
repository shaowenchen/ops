package cmd

import (
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/spf13/cobra"
)

type HostOption struct {
	Address         string
	Domain string
	IP string
}

var hostOpt HostOption

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "host your cluster smoothly to a newer version with this command",
	RunE: func(cmd *cobra.Command, args []string) error {
		h, _ := host.NewHost("", hostOpt.Address, "", 0, "", "", "", "", 0)
		h.AddHost(hostOpt.Domain, hostOpt.IP)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(hostCmd)
	hostCmd.Flags().StringVarP(&hostOpt.Address, "addr", "", "", "target host")
	hostCmd.Flags().StringVarP(&hostOpt.Domain, "domain", "", "", "/etc/hosts domain")
	hostCmd.Flags().StringVarP(&hostOpt.IP, "ip", "", "", "/etc/hosts ip")
}