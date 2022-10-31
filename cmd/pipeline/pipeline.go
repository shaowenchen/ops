package pipeline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shaowenchen/opscli/pkg/log"
	"github.com/shaowenchen/opscli/pkg/pipeline"
	"github.com/spf13/cobra"
)

var pipelineOption pipeline.PipelineOption

var PipelineCmd = &cobra.Command{
	Use:                "pipeline",
	Short:              "command about pipeline",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewDefaultLogger(true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		pipelineOption = parseArgs(args)
		if len(pipelineOption.FilePath) == 0 {
			fmt.Printf("--filepath is must provided")
			return
		}
		pipeline.ActionPipeline(logger, pipelineOption)
	},
}

func parseArgs(args []string) (pipelineOption pipeline.PipelineOption) {
	pipelineOption.Variables = make(map[string]string)
	for i := 0; i < len(args); i++ {
		fieldName := getArgName(args[i])
		if len(fieldName) > 0 {
			fieldValue := "true"
			if (i + 1) == len(args) {
				// --clear
			} else if (i+1) < len(args) && len(getArgName(args[i+1])) > 0 {
				// --clear --username root
			} else {
				// --username root
				fieldValue = args[i+1]
			}
			if fieldName == "debug" {
				pipelineOption.Debug = fieldValue == "true"
			} else if fieldName == "filepath" {
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
	}
	return ""
}

func init() {
	PipelineCmd.Flags().BoolVarP(&pipelineOption.Debug, "debug", "", false, "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.FilePath, "filepath", "", "", "")
	PipelineCmd.MarkFlagRequired("filepath")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Hosts, "hosts", "", "", "")
	PipelineCmd.Flags().IntVarP(&pipelineOption.Port, "port", "", 22, "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Username, "username", "", "", "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.Password, "password", "", "", "")
	PipelineCmd.Flags().StringVarP(&pipelineOption.PrivateKeyPath, "privatekeypath", "", "", "")
}
