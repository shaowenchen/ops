package upgrade

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/shaowenchen/ops/cmd/cli/root"
	"github.com/spf13/cobra"
)

var url = ""

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade opscli version to latest",
	Run: func(cmd *cobra.Command, args []string) {
		upgrade := exec.Command("sh", "-c", utils.ScriptInstallOpscli())
		var stdout bytes.Buffer
		upgrade.Stdout = &stdout
		upgrade.Run()
		fmt.Println(string(stdout.Bytes()))
	},
}

func init() {
	root.RootCmd.AddCommand(upgradeCmd)
}