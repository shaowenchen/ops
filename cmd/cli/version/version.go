package version

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BuildVersion = ""
	BuildDate    = ""
	GicommitID   = ""
)

type BuildInfo struct {
	Version   string `json:"Version,omitempty"`
	BuildDate string `json:"BuildDate,omitempty"`
	GitCommit string `json:"GitCommit,omitempty"`
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "get current version",
	Run: func(cmd *cobra.Command, args []string) {
		versionBytes, err := json.Marshal(
			BuildInfo{
				Version:   BuildVersion,
				BuildDate: BuildVersion,
				GitCommit: BuildVersion,
			})
		if err != nil {
			return
		}
		fmt.Println(string(versionBytes))
	},
}
