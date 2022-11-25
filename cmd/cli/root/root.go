package root

import (
	"fmt"
	"os"

	"github.com/shaowenchen/ops/cmd/cli/create"
	"github.com/shaowenchen/ops/cmd/cli/file"
	"github.com/shaowenchen/ops/cmd/cli/script"
	"github.com/shaowenchen/ops/cmd/cli/task"
	"github.com/shaowenchen/ops/cmd/cli/upgrade"
	"github.com/shaowenchen/ops/cmd/cli/version"
	"github.com/spf13/cobra"
)

func Execute() {
	RootCmd.AddCommand(file.FileCmd)
	RootCmd.AddCommand(script.ScriptCmd)
	RootCmd.AddCommand(create.CreateCmd)
	RootCmd.AddCommand(task.TaskCmd)
	RootCmd.AddCommand(version.VersionCmd)
	RootCmd.AddCommand(upgrade.UpgradeCmd)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var RootCmd = &cobra.Command{
	Use:   "opscli",
	Short: "a cli tool",
	Long:  `This is a cli tool for ops.`,
}
