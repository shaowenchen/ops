package cmd

import (
	"fmt"
	"os"

	"github.com/shaowenchen/opscli/cmd/docker"
	"github.com/shaowenchen/opscli/cmd/host"
	"github.com/shaowenchen/opscli/cmd/kube"
	"github.com/shaowenchen/opscli/cmd/pipeline"
	"github.com/shaowenchen/opscli/cmd/storage"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd.AddCommand(host.HostCmd)
	rootCmd.AddCommand(kube.KubeCmd)
	rootCmd.AddCommand(storage.StorageCmd)
	rootCmd.AddCommand(docker.DockerCmd)
	rootCmd.AddCommand(pipeline.PipelineCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "opscli",
	Short: "a cli tool",
	Long:  `This is a cli tool for ops.`,
}
