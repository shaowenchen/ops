package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
)

type PipelineRunsManager struct {
	endpoint  string
	token     string
	namespace string
	pipelines []opsv1.Pipeline
	clusters  []opsv1.Cluster
}

func NewPipelineRunsManager(endpoint, token, namespace string) (prManager *PipelineRunsManager, err error) {
	prManager = &PipelineRunsManager{
		endpoint:  endpoint,
		token:     token,
		namespace: namespace,
	}
	err = prManager.Init()
	return
}

func (m *PipelineRunsManager) Init() (err error) {
	m.pipelines, err = m.GetPipelines()
	if err != nil {
		return err
	}
	m.clusters, err = m.GetClusters()
	if err != nil {
		return err
	}
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			if p, e := m.GetPipelines(); e == nil {
				m.pipelines = p
			}
			if c, e := m.GetClusters(); e == nil {
				m.clusters = c
			}
		}
	}()

	return nil
}

func (m *PipelineRunsManager) PrintMarkdownClusters() (output string) {
	output = "### Kubernetes cluster list\n"
	if m.clusters == nil || len(m.clusters) == 0 {
		return "no any available cluster \n"
	}
	for i := 0; i < len(m.clusters); i++ {
		c := m.clusters[i]
		output += fmt.Sprintf("- %s(%s)\n", c.Name, c.Spec.Desc)
	}
	return
}

func (m *PipelineRunsManager) PrintMarkdownPipelines() (output string) {
	output = "### pipelines list\n"
	if m.pipelines == nil || len(m.pipelines) == 0 {
		return "no any available pipelines \n"
	}
	for i := 0; i < len(m.pipelines); i++ {
		p := m.pipelines[i]
		output += fmt.Sprintf("- %s(%s)\n", p.Name, p.Spec.Desc)
	}
	return
}

func (m *PipelineRunsManager) PrintMarkdownPipelineRuns(pr *opsv1.PipelineRun) (output string) {
	// if pipeline is default, print chat result
	if pr != nil && pr.Spec.PipelineRef == "default" {
		if len(pr.Status.PipelineRunStatus) > 0 &&
			len(pr.Status.PipelineRunStatus[0].TaskRunStatus.TaskRunNodeStatus) > 0 {
			for _, nodeStatus := range pr.Status.PipelineRunStatus[0].TaskRunStatus.TaskRunNodeStatus {
				if len(nodeStatus.TaskRunStep) > 0 {
					return nodeStatus.TaskRunStep[0].StepOutput
				}
			}
		}
		return "chat result is empty"
	}
	output = "###" + pr.Name + " Run Details\n"
	if pr == nil || len(pr.Status.PipelineRunStatus) == 0 {
		return "not run any task\n"
	}
	for i := 0; i < len(pr.Status.PipelineRunStatus); i++ {
		t := pr.Status.PipelineRunStatus[i]
		output += "#### " + t.TaskName + "\n"
		var b strings.Builder
		for _, nodeRunStatus := range t.TaskRunStatus.TaskRunNodeStatus {
			for _, step := range nodeRunStatus.TaskRunStep {
				if step.StepOutput == "" || step.StepOutput == opsconstants.NoOutput {
					continue
				}
				b.WriteString(fmt.Sprintf("- Step: %s\n", step.StepName))
				step.StepOutput = strings.ReplaceAll(step.StepOutput, "\n", "\n\n")
				b.WriteString(fmt.Sprintf("- Output:\n\n%s\n", step.StepOutput))
			}
		}
		output += b.String()
	}
	return
}

func (m *PipelineRunsManager) GetForIntent() string {
	var b strings.Builder
	for i := 0; i < len(m.pipelines); i++ {
		b.WriteString(fmt.Sprintf("- %s(%s)\n", m.pipelines[i].Name, m.pipelines[i].Spec.Desc))
	}
	return b.String()
}

func (m *PipelineRunsManager) GetForVariables(pr *opsv1.PipelineRun) string {
	if pr == nil || len(pr.Spec.Variables) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range pr.Spec.Variables {
		b.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
	}
	return b.String()
}

func (m *PipelineRunsManager) Run(logger *opslog.Logger, pipelinerun *opsv1.PipelineRun) (err error) {

	// patch pipelinerun
	uri := "/api/v1/namespaces/" + m.namespace + "/pipelineruns/sync"
	respBody, err := m.makeRequest(m.endpoint, m.token, uri, "POST", pipelinerun.Spec)
	if err != nil {
		logger.Error.Println("error", err)
		return
	}
	type Resp struct {
		Data opsv1.PipelineRun `json:"data"`
	}

	var resp Resp
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return
	}
	pipelinerun.Status = resp.Data.Status
	return
}

func (m *PipelineRunsManager) GetPipelines() (ps []opsv1.Pipeline, err error) {
	uri := "/api/v1/namespaces/" + m.namespace + "/pipelines?labels_selector=ops/copilot=enabled&page_size=999"
	body, err := m.makeRequest(m.endpoint, m.token, uri, "GET", nil)
	if err != nil || len(string(body)) < 10 {
		return
	}
	type ServerResponseList struct {
		Data struct {
			List []opsv1.Pipeline `json:"list"`
		} `json:"data"`
	}
	var resp ServerResponseList
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	ps = resp.Data.List
	return
}

func (m *PipelineRunsManager) GetClusters() (cs []opsv1.Cluster, err error) {
	uri := "/api/v1/namespaces/" + m.namespace + "/clusters?page_size=999"
	body, err := m.makeRequest(m.endpoint, m.token, uri, "GET", nil)
	if err != nil || len(string(body)) < 10 {
		return
	}
	type ServerResponseList struct {
		Data struct {
			List []opsv1.Cluster `json:"list"`
		} `json:"data"`
	}
	var resp ServerResponseList
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	cs = resp.Data.List
	return

}

func (m *PipelineRunsManager) makeRequest(endpoint, token, uri, method string, payload interface{}) ([]byte, error) {
	url := endpoint + uri

	client := &http.Client{Timeout: 600 * time.Second}

	var req *http.Request
	var err error

	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(payloadBytes)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (m *PipelineRunsManager) AddMcpTools(logger *opslog.Logger, s *server.MCPServer) error {
	pipelines, err := m.GetPipelines()
	if err != nil {
		return err
	}
	for _, pipeline := range pipelines {
		var toolOptions = make([]mcp.ToolOption, 0)
		for key, variable := range pipeline.Spec.Variables {
			if variable.Value != "" {
				variable.Default = variable.Value
			}
			toolOptions = append(toolOptions, mcp.WithString(key,
				mcp.Required(),
				mcp.Description(variable.Desc),
				mcp.Enum(variable.Enums...),
				mcp.DefaultString(variable.Default),
				mcp.Pattern(variable.Regex),
			))
		}
		mcpTool := mcp.NewTool(pipeline.Name, toolOptions...)
		s.AddTool(mcpTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			variables := make(map[string]string)
			// Todo: fix
			// for key, value := range request.Params.Arguments {
			// 	variables[key] = value.(string)
			// }
			pipelines, err := m.GetPipelines()
			if err != nil {
				return mcp.NewToolResultText(err.Error()), nil
			}
			output := ""
			for _, pipeline := range pipelines {
				if pipeline.Name == request.Params.Name {
					pipelinerun := opsv1.NewPipelineRun(&pipeline)
					pipelinerun.Spec.Variables = variables
					err := m.Run(logger, pipelinerun)
					if err != nil {
						logger.Error.Println(err)
						return mcp.NewToolResultText(err.Error()), nil
					}
					output = m.PrintMarkdownPipelineRuns(pipelinerun)
				}
			}
			return mcp.NewToolResultText(output), nil
		})
	}
	return nil
}

func (m *PipelineRunsManager) AddMcpResources(logger *opslog.Logger, s *server.MCPServer) error {
	clusterRes := mcp.NewResource(
		"clusters://all",
		"all available clusters",
		mcp.WithResourceDescription("clusters are the kubernetes clusters"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.AddResource(clusterRes, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "clusters://all",
				MIMEType: "text/markdown",
				Text:     string(m.PrintMarkdownClusters()),
			},
		}, nil
	})
	pipelineRes := mcp.NewResource(
		"pipelines://all",
		"all available pipelines",
		mcp.WithResourceDescription("pipelines provide standard ops procedures"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.AddResource(pipelineRes, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "pipelines://all",
				MIMEType: "text/markdown",
				Text:     string(m.PrintMarkdownPipelines()),
			},
		}, nil
	})
	return nil
}
