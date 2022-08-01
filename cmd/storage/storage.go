package storage

import (
	"github.com/spf13/cobra"
)

var StorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "config storage with this command",
}

func init() {
	StorageCmd.AddCommand(s3FileCmd)
}
