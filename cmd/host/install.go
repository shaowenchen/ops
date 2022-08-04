package host

import (
	"github.com/shaowenchen/opscli/pkg/host"
	"github.com/spf13/cobra"
)

var installOpt host.InstallOption

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install Component on host",
	RunE: func(cmd *cobra.Command, args []string) error {
		return host.ActionInstall(installOpt)
	},
}

func init() {
	installCmd.Flags().StringVarP(&installOpt.Username, "username", "", "root", "")
	installCmd.Flags().StringVarP(&installOpt.Username, "password", "", "", "")
	installCmd.Flags().StringVarP(&installOpt.PrivateKeyPath, "privatekeypath", "", "", "")
	installCmd.Flags().StringVarP(&installOpt.Name, "name", "", "", "")
	installCmd.Flags().BoolVarP(&installOpt.Clear, "clear", "", false, "")
}
