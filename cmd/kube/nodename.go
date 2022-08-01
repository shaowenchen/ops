package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)


var nodeNameOption kube.NodeNameOption

var nodeNameCmd = &cobra.Command{
	Use:   "nodename",
	Short: "config nodeName for kubernetes deployment",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return kube.ActionNodeName(nodeNameOption)
	},
}

func init() {
	nodeNameCmd.Flags().StringVarP(&nodeNameOption.Kubeconfig, "kubeconfig", "", "", "")
	nodeNameCmd.Flags().StringVarP(&nodeNameOption.NodeName, "nodename", "", "", "e.g., node1")
	nodeNameCmd.MarkFlagRequired("nodeName")
	nodeNameCmd.Flags().StringVarP(&nodeNameOption.Name, "name", "", "", "NamespacedName (required), e.g., default/mydeploy")
	nodeNameCmd.MarkFlagRequired("name")
	nodeNameCmd.Flags().BoolVarP(&nodeNameOption.Clear, "clear", "", false, "")
}
