package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/shaowenchen/opscli/pkg/utils"
	"github.com/spf13/cobra"
)

var url = ""

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade opscli version to latest",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsSudoUser() {
			fmt.Println("please run with root")
			return
		}
		upgrade := exec.Command("sudo", "sh", "-c", utils.ScriptInstallOpscli())
		var stderr bytes.Buffer
		var stdout bytes.Buffer
		upgrade.Stderr = &stderr
		upgrade.Stdout = &stdout
		err := upgrade.Run()
		fmt.Println(string(stderr.Bytes()))
		fmt.Println(string(stdout.Bytes()))
		if err != nil {
			fmt.Println("Upgrade failed!")
		} else {
			fmt.Println("Upgrade success!")
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
