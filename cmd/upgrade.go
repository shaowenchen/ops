package cmd

import (
	"fmt"
	"os/exec"

	"github.com/shaowenchen/opscli/pkg/script"
	"github.com/spf13/cobra"
)

var url = ""

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade opscli version to latest",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		upgrade := exec.Command("sudo", "bash", "-c", script.GetAvailableUrl("https://raw.githubusercontent.com/shaowenchen/opscli/main/getopscli.sh"))
		_, err = upgrade.Output()
		if err != nil {
			fmt.Println("Upgrade failed!")
		}
		fmt.Println("Upgrade success!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
