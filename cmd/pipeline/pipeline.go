package pipeline

import (
	"github.com/shaowenchen/opscli/pkg/pipeline"
	"github.com/spf13/cobra"
)

var pipelineOption pipeline.PipelineOption

var PipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "run pipeline with this command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pipeline.ActionPipeline(pipelineOption)
	},
}

func init() {
	PipelineCmd.Flags().BoolVarP(&pipelineOption.Debug, "debug", "", true, "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Hosts, "hosts", "", "", "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.FilePath, "filepath", "f", "", "")
	PipelineCmd.MarkFlagRequired("filepath")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Username, "username", "", "", "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Password, "password", "", "", "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.PrivateKeyPath, "privatekeypath", "", "", "")
}
