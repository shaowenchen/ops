package create

import (
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "command about Ops Resource",
}

var hostOpt option.HostOption
var clusterOpt option.ClusterOption
var clusterSpec opsv1.ClusterSpec
var taskOpt option.TaskOption
var inventory string
var verbose string

func init() {
	CreateCmd.AddCommand(hostCmd)
	CreateCmd.AddCommand(clusterCmd)
	CreateCmd.AddCommand(taskCmd)
}
