package upgrade

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/shaowenchen/ops/cmd/cli/config"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var proxy = ""
var manifests = false
var UpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		// Get proxy with priority: CLI > ENV > Config > Default
		proxy = config.GetValueWithPriority(proxy, constants.EnvProxy, "proxy", constants.DefaultProxy)

		bash := ""
		if manifests {
			bash = utils.ShellInstallManifests(proxy)
		} else {
			bash = utils.ShellInstallOpscli(proxy)
		}
		upgrade := exec.Command("sh", "-c", bash)
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
	UpgradeCmd.Flags().BoolVarP(&manifests, "manifests", "", false, "only upgrade manifests")
}
