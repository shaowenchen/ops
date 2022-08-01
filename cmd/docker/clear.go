package docker

import (
	"github.com/shaowenchen/opscli/pkg/docker"
	"github.com/spf13/cobra"
)

var clearOpt docker.ClearOption

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear docker",
	RunE: func(cmd *cobra.Command, args []string) error {
		return docker.ActionClear(clearOpt)
	},
}

func init() {
	clearCmd.Flags().StringVarP(&clearOpt.Input, "input", "", "", "")
	clearCmd.Flags().StringVarP(&clearOpt.NameRegx, "nameregx", "", "", "")
	clearCmd.Flags().StringVarP(&clearOpt.TagRegx, "tagregex", "", "master|main|gray|test|dev|feature|rc|issue", "")
	clearCmd.Flags().BoolVarP(&clearOpt.Force, "force", "", false, "")
}
