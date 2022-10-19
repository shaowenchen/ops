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
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		upgrade := exec.Command("sh", "-c", utils.InstallOpscli())
		var stderr bytes.Buffer
		upgrade.Stderr = &stderr
		err = upgrade.Run()
		if err != nil {
			fmt.Println("Upgrade failed! %s", stderr.String())
			return err
		}
		fmt.Println("Upgrade success!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
