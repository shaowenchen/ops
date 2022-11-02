package upgrade

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var url = ""

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		upgrade := exec.Command("sh", "-c", utils.ScriptInstallOpscli())
		var stdout bytes.Buffer
		upgrade.Stdout = &stdout
		upgrade.Run()
		fmt.Println(string(stdout.Bytes()))
	},
}
