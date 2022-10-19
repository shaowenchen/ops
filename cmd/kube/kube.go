package kube

import (
	"github.com/spf13/cobra"
)

var KubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "use kubeconfig to config kubernetes",
}

func init() {
	KubeCmd.AddCommand(etcHostsCmd)
	KubeCmd.AddCommand(imagePulllSecretCmd)
	KubeCmd.AddCommand(annotationCmd)
	KubeCmd.AddCommand(limitRangeCmd)
	KubeCmd.AddCommand(nodeNameCmd)
	KubeCmd.AddCommand(nodeSelectorCmd)
	KubeCmd.AddCommand(clearCmd)
	KubeCmd.AddCommand(deschedulerCmd)
	KubeCmd.AddCommand(hostRunCmd)
}
