package cmd

import (
	"github.com/spf13/cobra"
)

// initCmd represents the create command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the installation environment",
}

func init() {
	rootCmd.AddCommand(initCmd)
}