package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildVersion = ""
	buildDate    = ""
	gicommitID   = ""
)

type BuildInfo struct {
	Version   string `json:"Version,omitempty"`
	BuildDate string `json:"BuildDate,omitempty"`
	GitCommit string `json:"GitCommit,omitempty"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "get current opscli version",
	Run: func(cmd *cobra.Command, args []string) {
		versionBytes, err := json.Marshal(
			BuildInfo{
				Version:   buildVersion,
				BuildDate: buildDate,
				GitCommit: gicommitID,
			})
		if err != nil {
			return
		}
		fmt.Println(string(versionBytes))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
