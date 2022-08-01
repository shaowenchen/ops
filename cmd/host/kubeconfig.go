package host

import (
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/spf13/cobra"
)

var kubeconfigOpt host.KubeconfigOption

var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "get kubeconfig from remote host",
	RunE: func(cmd *cobra.Command, args []string) error {
		return host.ActionGetKubeconfig(kubeconfigOpt)
	},
}

func init() {
	kubeconfigCmd.Flags().StringVarP(&kubeconfigOpt.Username, "username", "", "", "")
	kubeconfigCmd.Flags().StringVarP(&kubeconfigOpt.Password, "password", "", "", "")
	kubeconfigCmd.Flags().StringVarP(&kubeconfigOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	kubeconfigCmd.Flags().StringVarP(&kubeconfigOpt.Input, "input", "i", "", "host addr")
	kubeconfigCmd.Flags().BoolVarP(&kubeconfigOpt.Clear, "clear", "", false, "")
}
