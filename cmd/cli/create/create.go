package create

import (
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "command about Ops Resource",
}

var createHostOpt option.CreateHostOption
var createClusterOpt option.CreateClusterOption
var createTaskOpt option.CreateTaskOption
var inventory string

func init() {
	CreateCmd.AddCommand(hostCmd)
	CreateCmd.AddCommand(clusterCmd)
	CreateCmd.AddCommand(taskCmd)
}
