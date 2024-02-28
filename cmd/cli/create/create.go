package create

import (
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "command about Ops Resource",
}

func init() {
	CreateCmd.AddCommand(hostCmd)
	CreateCmd.AddCommand(clusterCmd)
	CreateCmd.AddCommand(taskCmd)
}
