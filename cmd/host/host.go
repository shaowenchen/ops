package host

import (
	"github.com/spf13/cobra"
)

var HostCmd = &cobra.Command{
	Use:   "host",
	Short: "config host with this command",
}

func init() {
	HostCmd.AddCommand(etcHostsCmd)
	HostCmd.AddCommand(kubeconfigCmd)
}
