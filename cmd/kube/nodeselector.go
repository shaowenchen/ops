package kube

import (
	"github.com/shaowenchen/opscli/pkg/kube"
	"github.com/spf13/cobra"
)

var nodeSelectorOption kube.NodeSelectorOption

var nodeSelectorCmd = &cobra.Command{
	Use:   "nodeselector",
	Short: "config nodeSelector for kubernetes deployment",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return kube.ActionNodeSelector(nodeSelectorOption)
	},
}

func init() {
	nodeSelectorCmd.Flags().StringVarP(&nodeSelectorOption.Kubeconfig, "kubeconfig", "", "", "")
	nodeSelectorCmd.Flags().StringVarP(&nodeSelectorOption.Name, "name", "", "", "NamespacedName (required), e.g., default/mydeploy")
	nodeSelectorCmd.MarkFlagRequired("name")
	nodeSelectorCmd.Flags().StringVarP(&nodeSelectorOption.KeyLabels, "keylabel", "", "", "keylabel, e.g., kubernetes.io/hostname=node1")
	nodeSelectorCmd.Flags().BoolVarP(&nodeSelectorOption.Clear, "clear", "", false, "")
}
