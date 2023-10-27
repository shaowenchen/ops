package upgrade

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var proxy = ""

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		upgrade := exec.Command("sh", "-c", utils.ShellInstallOpscli(proxy))
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		upgrade.Stdout = &stdout
		upgrade.Stderr = &stdout
		err := upgrade.Run()
		if err != nil {
			fmt.Println(err)
		}
		if len(stderr.Bytes()) > 0 {
			fmt.Println(string(stderr.Bytes()))
		} else if len(stdout.Bytes()) > 0 {
			fmt.Println(string(stdout.Bytes()))
		}
	},
}

func init() {
	UpgradeCmd.Flags().StringVarP(&proxy, "proxy", "", constants.DefaultProxy, "")
}
