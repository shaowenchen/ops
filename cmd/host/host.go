package host

import (
	"github.com/spf13/cobra"
)

var HostCmd = &cobra.Command{
	Use:   "host",
	Short: "command about host",
}

func init() {
	HostCmd.AddCommand(fileCmd)
	HostCmd.AddCommand(scriptCmd)
}
