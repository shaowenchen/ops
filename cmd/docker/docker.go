package docker

import (
	"github.com/spf13/cobra"
)

var DockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "config docker with this command",
}

func init() {
	DockerCmd.AddCommand(clearCmd)
}
