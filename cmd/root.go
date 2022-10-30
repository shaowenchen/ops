package cmd

import (
	"fmt"
	"os"

	"github.com/shaowenchen/opscli/cmd/host"
	"github.com/shaowenchen/opscli/cmd/kube"
	"github.com/shaowenchen/opscli/cmd/pipeline"
	"github.com/shaowenchen/opscli/cmd/storage"
	"github.com/shaowenchen/opscli/pkg/constants"
	"github.com/shaowenchen/opscli/pkg/utils"
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd.AddCommand(host.HostCmd)
	rootCmd.AddCommand(kube.KubeCmd)
	rootCmd.AddCommand(storage.StorageCmd)
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

func init() {
	utils.CreateDir(constants.GetOpscliLogsDir())
}
