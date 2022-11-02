package storage

import (
	"github.com/spf13/cobra"
)

var StorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "command about remote storage",
}

func init() {
	StorageCmd.AddCommand(s3FileCmd)
}
