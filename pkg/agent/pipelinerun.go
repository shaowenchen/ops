package agent

import (
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
	"strings"
	"sync"
	"time"
)

type LLMPipelineRun struct {
	Desc        string            `json:"desc"`
	Creator     string            `json:"creator"`
	Namespace   string            `json:"namespace"`
	NameRef     string            `json:"nameRef"`
	NodeName    string            `json:"nodeName"`
	PipelineRef string            `json:"pipelineRef"`
	TypeRef     string            `json:"typeRef"`
	Variables   map[string]string `json:"variables"`
	TaskRuns    []*LLMTaskRun     `json:"taskRuns"`
	Output      string            `json:"output"`
	RunStatus   string            `json:"runStatus"`
}

func NewLLMPipelineRunsManager(enpoint, token, namespace, runtimeimage string, syncTickerSeconds uint, allPipelines []LLMPipeline, allTasks []LLMTask) (prManager *LLMPipelineRunsManager) {
	prManager = &LLMPipelineRunsManager{}
	prManager.taskrunsManager = NewLLMTaskRunManager(enpoint, token, namespace, runtimeimage, prManager)
	prManager.GetHostManager().StartUpdateTimer(time.Duration(syncTickerSeconds), prManager.GetHostManager().Update)
	prManager.GetClusterManager().StartUpdateTimer(time.Duration(syncTickerSeconds)*time.Second, prManager.GetClusterManager().Update)
	prManager.GetTaskRunManager().StartUpdateTimer(time.Duration(syncTickerSeconds), prManager.GetTaskRunManager().Update)
	prManager.RegisterPipelines(allPipelines...)
	prManager.RegisterTasks(allTasks...)
	return
}

type LLMPipelineRunsManager struct {
	endpoint        string
	token           string
	namespace       string
	runtimeImage    string
	pipelines       []LLMPipeline
	tickOnce        sync.Once
	taskrunsManager *LLMTaskRunManager
}

func (m *LLMPipelineRunsManager) GetHostManager() *LLMHostsManager {
	return m.taskrunsManager.hostsManager
}

func (m *LLMPipelineRunsManager) GetClusterManager() *LLMClustersManager {
	return m.taskrunsManager.clustersManager
}

func (m *LLMPipelineRunsManager) GetTaskRunManager() *LLMTaskRunManager {
	return m.taskrunsManager
}

func (m *LLMPipelineRunsManager) GetPipelineTools() []openai.Tool {
	return m.BuildTools()
}

func (m *LLMPipelineRunsManager) GetPipelineByClearUnavailableChar(name string) (LLMPipeline, error) {
	for _, pipeline := range m.pipelines {
		if ClearUnavailableChar(pipeline.Name) == name {
			return pipeline, nil
		}
	}
	return LLMPipeline{}, errors.New("pipeline not found")
}

func (m *LLMPipelineRunsManager) RegisterPipelines(ps ...LLMPipeline) *LLMPipelineRunsManager {
	return m.AddPipelines(ps...)
}

func (m *LLMPipelineRunsManager) RegisterTasks(ts ...LLMTask) *LLMPipelineRunsManager {
	m.taskrunsManager.AddTasks(ts...)
	return m
}

func (m *LLMPipelineRunsManager) AddPipelines(ps ...LLMPipeline) *LLMPipelineRunsManager {
	existings := make(map[string]struct{})
	for _, task := range m.pipelines {
		if task.Namespace == "" {
			task.Namespace = OpsDefaultNamespace
		}
		key := fmt.Sprintf("%s-%s", task.Namespace, task.Name)
		existings[key] = struct{}{}
	}

	for _, new := range ps {
		key := fmt.Sprintf("%s-%s", new.Namespace, new.Name)
		if _, exists := existings[key]; !exists {
			m.pipelines = append(m.pipelines, new)
			existings[key] = struct{}{}
		}
	}

	return m
}

func (m *LLMPipelineRunsManager) GetClustersMarkdown() string {
	return m.GetClusterManager().GetMarkdown()
}

func (m *LLMPipelineRunsManager) GetTasksMarkdown() string {
	return m.taskrunsManager.GetMarkdown()
}

func (m *LLMPipelineRunsManager) Rebuild(pipelinerun *LLMPipelineRun) error {
	pipeline, err := m.GetPipelineByClearUnavailableChar(pipelinerun.PipelineRef)
	if err != nil {
		return err
	}
	// fill taskruns
	for _, task := range pipeline.LLMTasks {
		task, err = m.taskrunsManager.Rebuild(&task)
		if err != nil {
			return err
		}
		variables := task.Variables
		if variables == nil {
			variables = make(map[string]string)
		}
		for key := range variables {
			// pipeline > pipelinerun > task
			if len(task.Variables) > 0 {
				_, ok := task.Variables[key]
				if ok && task.Variables[key] != "" {
					variables[key] = task.Variables[key]
				}
			}
			if len(pipelinerun.Variables) > 0 {
				_, ok := pipelinerun.Variables[key]
				if ok && pipelinerun.Variables[key] != "" {
					variables[key] = pipelinerun.Variables[key]
				}
			}
		}
		// nodeName, pipeline > pipelinerun
		nodeName := pipeline.NodeName
		if task.NodeName == "anymaster" {
			nodeName = "anymaster"
		} else if pipelinerun.NodeName != "" {
			nodeName = pipelinerun.NodeName
		}
		// nameRef, pipeline > pipelinerun
		nameRef := pipeline.NameRef
		if pipelinerun.NameRef != "" {
			nameRef = pipelinerun.NameRef
		}
		pipelinerun.TaskRuns = append(pipelinerun.TaskRuns, &LLMTaskRun{
			Namespace: pipelinerun.Namespace,
			TaskRef:   task.Name,
			TypeRef:   pipelinerun.TypeRef,
			NameRef:   nameRef,
			NodeName:  nodeName,
			Variables: variables,
			Always:    task.Always,
		})
	}
	return nil
}

func (m *LLMPipelineRunsManager) Run(logger *log.Logger, pipelinerun *LLMPipelineRun) (err error) {
	if logger == nil {
		return errors.New("logger is nil")
	}
	if pipelinerun == nil {
		return errors.New("pipelinerun is nil")
	}
	if pipelinerun.RunStatus != opsv1.StatusRunning {
		pipelinerun.RunStatus = opsv1.StatusRunning
	}
	err = m.Rebuild(pipelinerun)
	if err != nil {
		return err
	}
	onlyAlways := false
	for _, tr := range pipelinerun.TaskRuns {
		logger.Debug.Printf("run taskrun: %s\n", tr.TaskRef)
		if onlyAlways == false || onlyAlways == true && tr.Always == true {
			err = m.taskrunsManager.Run(logger, pipelinerun, tr)
			logger.Debug.Printf("run taskrun: %s, output: %s\n", tr.TaskRef, tr.Output)
			if err != nil || tr.RunStatus != opsv1.StatusSuccessed {
				logger.Error.Printf("run taskrun err: %v, status: %s\n", err, tr.RunStatus)
				onlyAlways = true
			}
		} else {
			tr.Output = "skipped"
			logger.Debug.Printf("skip taskrun: %s\n", tr.TaskRef)
		}
	}
	if pipelinerun.Output == "" && len(pipelinerun.TaskRuns) > 0 {
		pipelinerun.Output = pipelinerun.TaskRuns[len(pipelinerun.TaskRuns)-1].Output
	}
	if pipelinerun.RunStatus == opsv1.StatusRunning {
		pipelinerun.RunStatus = opsv1.StatusSuccessed
	}
	logger.Debug.Printf("run pipeline: %s, output: %s\n", pipelinerun.PipelineRef, pipelinerun.Output)
	return
}

func (p *LLMPipelineRunsManager) GetLLMTasks() []LLMTask {
	return p.taskrunsManager.tasks
}

func (m *LLMPipelineRunsManager) GetLLMPipelines() []LLMPipeline {
	return m.pipelines
}

func (m *LLMPipelineRunsManager) BuildTools() (tools []openai.Tool) {
	for _, pipeline := range m.pipelines {
		tools = append(tools, m.BuildTool(pipeline))
	}
	return
}

func (m *LLMPipelineRunsManager) BuildTool(p LLMPipeline) openai.Tool {
	parmerters := jsonschema.Definition{
		Type: "object",
		Properties: map[string]jsonschema.Definition{
			"typeRef": {
				Type:        "string",
				Description: "just set it to cluster",
				Enum:        []string{"cluster"},
			},
			"nameRef": {
				Type:        "string",
				Description: m.GetClusterManager().GetText(),
				Enum:        m.GetClusterManager().GetList(),
			},
			"nodeName": {
				Type:        "string",
				Description: "if typeRef is cluster, nodeName is a host name. if not found, use anymaster",
			},
		},
		Required: []string{"typeRef", "nameRef", "nodeName"},
	}

	for _, v := range p.Variables {
		if _, ok := parmerters.Properties[v.Key]; ok {
			scheme := parmerters.Properties[v.Key]
			scheme.Description = p.Desc
			parmerters.Properties[v.Key] = scheme
			continue
		}
		parmerters.Properties[v.Key] = jsonschema.Definition{
			Type:        jsonschema.DataType(v.Value.Type),
			Description: v.Desc,
		}
		if v.Required {
			parmerters.Required = append(parmerters.Required, v.Key)
		}
	}
	tool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        ClearUnavailableChar(p.Name),
			Description: p.Desc,
			Parameters:  parmerters,
		},
	}
	return tool
}

func (m *LLMPipelineRunsManager) Update() (ps []LLMPipeline, err error) {
	uri := "/api/v1/namespaces/" + m.namespace + "/pipelines?page_size=999"
	body, err := makeRequest(m.endpoint, m.token, uri, "GET", nil)
	if err != nil || len(string(body)) < 10 {
		return
	}
	//Todo
	return
}

func (m *LLMPipelineRunsManager) StartUpdateTimer(interval time.Duration, updateFunc func() ([]LLMPipeline, error)) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.tickOnce.Do(func() {
				res, err := updateFunc()
				if err != nil {
					fmt.Printf("timer update hosts err: %v\n", err)
					return
				}
				m.AddPipelines(res...)
			})
		}
	}()
	pipelines, _ := updateFunc()
	m.AddPipelines(pipelines...)
}

func (m *LLMPipelineRunsManager) GetMarkdown() string {
	var b strings.Builder
	b.WriteString("| name | desc | variables |\n")
	b.WriteString("|-|-|-|\n")
	for _, item := range m.pipelines {
		var vars string
		for _, v := range item.Variables {
			vars += fmt.Sprintf("%s=%s,", v.Key, v.DefaultValue)
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", item.Name, item.Desc, vars))
	}
	return b.String()

}
