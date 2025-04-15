package mcp

import (
	"fmt"
	"github.com/mark3labs/mcp-go/server"
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
		mcpServer := server.NewMCPServer(
			"Ops Mcp Server",
			"1.0.0",
			server.WithResourceCapabilities(true, true),
			server.WithLogging(),
		)
		err = pipelinerunsManager.AddMcpTools(logger, mcpServer)
		if err != nil {
			logger.Error.Println("init mcp failed " + err.Error())
			return
		}
		if err := server.ServeStdio(mcpServer); err != nil {
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
