package mcp

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/copilot"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/spf13/cobra"
)

var opsServer string
var opsToken string
var verbose string

var McpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "run mcp server",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()
		logger.Info.Println(opsServer)
		pipelinerunsManager, err := copilot.NewPipelineRunsManager(opsServer, opsToken, "ops-system")
		if err != nil {
			logger.Error.Println("request ops server failed " + err.Error())
			return
		}
		pipelines, err := pipelinerunsManager.GetPipelines()
		if err != nil {
			logger.Error.Println("get pipelines failed " + err.Error())
			return
		}
		s := server.NewMCPServer(
			"Ops Mcp Server",
			"1.0.0",
			server.WithResourceCapabilities(true, true),
			server.WithLogging(),
		)
		for _, pipeline := range pipelines {
			var toolOptions = make([]mcp.ToolOption, 0)
			for key, variable := range pipeline.Spec.Variables {
				toolOptions = append(toolOptions, mcp.WithString(key,
					mcp.Required(),
					mcp.Description(variable.Desc),
					mcp.Enum(variable.Enums...),
					mcp.DefaultString(variable.Default),
				))
			}
			mcpTool := mcp.NewTool(pipeline.Name, toolOptions...)
			s.AddTool(mcpTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				logger.Info.Println(request.Params.Name)
				variables := make(map[string]string)
				for key, value := range request.Params.Arguments {
					variables[key] = value.(string)
				}
				pipelines, err := pipelinerunsManager.GetPipelines()
				if err != nil {
					return mcp.NewToolResultText(err.Error()), nil
				}
				output := ""
				for _, pipeline := range pipelines {
					if pipeline.Name == request.Params.Name {
						pipelinerun := opsv1.NewPipelineRun(&pipeline)
						pipelinerun.Spec.Variables = variables
						pipelinerunsManager.Run(logger, pipelinerun)
						output = pipelinerunsManager.PrintMarkdownPipelineRuns(pipelinerun)
					}
				}
				logger.Info.Printf(output)
				return mcp.NewToolResultText(output), nil
			})
		}

		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Server error: %v\n", err)
			return
		}
	},
}

func init() {
	McpCmd.Flags().StringVarP(&opsServer, "opsserver", "", "", "")
	McpCmd.MarkFlagRequired("opsserver")
	McpCmd.Flags().StringVarP(&opsToken, "opstoken", "", "", "")
	McpCmd.MarkFlagRequired("opstoken")
	McpCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
}
