package pipeline

import (
	"github.com/shaowenchen/opscli/pkg/pipeline"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

var pipelineOption pipeline.PipelineOption

var PipelineCmd = &cobra.Command{
	Use:                "pipeline",
	Short:              "run pipeline with this command",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		pipelineOption = parseArgs(args)
		pipeline.ActionPipeline(pipelineOption)
	},
}

func parseArgs(args []string) (pipelineOption pipeline.PipelineOption) {
	pipelineOption.Variables = make(map[string]string, 0)
	for i := 0; i < len(args); i++ {
		fieldName := getArgName(args[i])
		if len(fieldName) > 0 {
			if fieldName == "debug" {
				pipelineOption.Debug = true
				continue
			}
			fieldValue := args[i+1]
			if fieldName == "filepath" || fieldName == "f" {
				pipelineOption.FilePath = fieldValue
			} else if fieldName == "hosts" {
				pipelineOption.Hosts = fieldValue
			} else if fieldName == "port" {
				pipelineOption.Port, _ = strconv.Atoi(fieldValue)
			} else if fieldName == "username" {
				pipelineOption.Username = fieldValue
			} else if fieldName == "password" {
				pipelineOption.Password = fieldValue
			} else if fieldName == "privatekeypath" {
				pipelineOption.PrivateKeyPath = fieldValue
			} else {
				pipelineOption.Variables[fieldName] = fieldValue
			}
		}
	}
	return
}

func getArgName(arg string) string {
	if strings.HasPrefix(arg, "--") {
		return arg[2:]
	} else if strings.HasPrefix(arg, "-") {
		return arg[1:]
	}
	return ""
}

func init() {
	PipelineCmd.Flags().BoolVarP(&pipelineOption.Debug, "debug", "", true, "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.FilePath, "filepath", "f", "", "")
	PipelineCmd.MarkFlagRequired("filepath")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Hosts, "hosts", "", "", "")
	PipelineCmd.Flags().IntVarP(&pipelineOption.Port, "port", "", 22, "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Username, "username", "", "", "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Password, "password", "", "", "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.PrivateKeyPath, "privatekeypath", "", "", "")
}
