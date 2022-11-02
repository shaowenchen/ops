package kubernetes

import (
	"github.com/spf13/cobra"
)

var KubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "command about kubernetes",
}

func init() {
	KubernetesCmd.AddCommand(scriptCmd)
	KubernetesCmd.AddCommand(fileCmd)
}
