package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
	"strings"
	"sync"
	"time"
)

type LLMTaskRun struct {
	Desc         string            `json:"desc"`
	Namespace    string            `json:"namespace"`
	TaskRef      string            `json:"taskRef"`
	TypeRef      string            `json:"typeRef"`
	NameRef      string            `json:"nameRef"`
	NodeName     string            `json:"nodeName"`
	Variables    map[string]string `json:"variables"`
	RuntimeImage string            `json:"runtimeImage"`
	Output       string            `json:"output"`
	Always       bool              `json:"always"`
	RunStatus    string            `json:"runStatus"`
}

type LLMTaskRunManager struct {
	endpoint           string
	token              string
	namespace          string
	runtimeImage       string
	tasks              []LLMTask
	tickOnce           sync.Once
	hostsManager       *LLMHostsManager
	clustersManager    *LLMClustersManager
	pipelinerunManager *LLMPipelineRunsManager
}

func NewLLMTaskRunManager(endpoint, token, namespace, runtimeImage string, pipelinerunManager *LLMPipelineRunsManager) *LLMTaskRunManager {
	return &LLMTaskRunManager{
		endpoint:     endpoint,
		token:        token,
		namespace:    namespace,
		runtimeImage: runtimeImage,
		tasks:        make([]LLMTask, 0),
		hostsManager: NewLLMHostsManager(
			endpoint,
			token,
			namespace,
		),
		clustersManager: NewLLMClustersManager(
			endpoint,
			token,
			namespace,
		),
		pipelinerunManager: pipelinerunManager,
	}
}

func (m *LLMTaskRunManager) GetMarkdown() string {
	var b strings.Builder
	b.WriteString("| name | desc | variables |\n")
	b.WriteString("|-|-|-|\n")
	for _, item := range m.tasks {
		var vars string
		for k, v := range item.Variables {
			vars += fmt.Sprintf("%s=%s,", k, v)
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", item.Name, item.Desc, vars))
	}
	return b.String()
}

func (m *LLMTaskRunManager) GetLLMTask(name string) (*LLMTask, error) {
	for _, t := range m.tasks {
		if strings.TrimSpace(t.Name) == strings.TrimSpace(name) {
			return &t, nil
		}
	}
	return nil, errors.New(name + " task not found")
}

func (m *LLMTaskRunManager) Rebuild(t *LLMTask) (task LLMTask, err error) {
	taskTmp, err := m.GetLLMTask(t.Name)
	if err != nil {
		return
	}
	task = LLMTask{
		Desc:         taskTmp.Desc,
		Namespace:    taskTmp.Namespace,
		Name:         taskTmp.Name,
		TypeRef:      taskTmp.TypeRef,
		NameRef:      taskTmp.NameRef,
		NodeName:     taskTmp.NodeName,
		RuntimeImage: taskTmp.RuntimeImage,
		Variables:    taskTmp.Variables,
		CallFunction: taskTmp.CallFunction,
	}
	if t.Always {
		task.Always = true
	}
	return
}

func (m *LLMTaskRunManager) Run(logger *log.Logger, pr *LLMPipelineRun, tr *LLMTaskRun) (err error) {
	if logger == nil {
		return errors.New("logger is nil")
	}
	if pr == nil {
		return errors.New("pipelinerun is nil")
	}
	if tr == nil {
		return errors.New("taskrun is nil")
	}

	task, err := m.GetLLMTask(tr.TaskRef)
	if err != nil {
		return err
	}
	if task.CallFunction != nil {
		tr.Output, err = task.CallFunction(logger, m.pipelinerunManager, pr)
		if tr.Output == "" {
			tr.Output = "no output or not found"
		}
		logger.Debug.Printf("call function for %s, output: %s \n", tr.TaskRef, tr.Output)
		return
	}
	// fill runtime image
	if m.runtimeImage != "" {
		tr.RuntimeImage = m.runtimeImage
	}
	logger.Debug.Printf("create taskrun, taskRef: %v, typeRef: %v, nameRef: %v, nodeName: %v, variables: %v\n", tr.TaskRef, tr.TypeRef, tr.NameRef, tr.NodeName, tr.Variables)
	taskrun, err := m.request(tr)
	if err != nil {
		logger.Debug.Printf("request err: %v\n", err)
		pr.RunStatus = opsv1.StatusFailed
		return err
	} else {
		tr.Output = GetTaskRunMarkdown(taskrun, nil)
		tr.RunStatus = taskrun.Status.RunStatus
	}
	return
}

func (m *LLMTaskRunManager) Find(task *LLMTask) (LLMTask, error) {
	for _, t := range m.tasks {
		if t.Name == task.Name {
			return t, nil
		}
	}
	return LLMTask{}, errors.New("task not found")
}

func (m *LLMTaskRunManager) Update() (ts []LLMTask, err error) {
	uri := "/api/v1/namespaces/" + m.namespace + "/tasks?page_size=999"
	body, err := makeRequest(m.endpoint, m.token, uri, "GET", nil)
	if err != nil || len(string(body)) < 10 {
		return
	}
	type ServerResponseList struct {
		Data struct {
			List []opsv1.Task `json:"list"`
		} `json:"data"`
	}
	var resp ServerResponseList
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	for _, t := range resp.Data.List {
		ts = append(ts, NewLLMTask(&t))
	}
	return
}

func (m *LLMTaskRunManager) request(tr *LLMTaskRun) (opsv1.TaskRun, error) {
	uri := "/api/v1/namespaces/" + tr.Namespace + "/taskruns"
	body, err := makeRequest(m.endpoint, m.token, uri, "POST", tr)
	if err != nil {
		return opsv1.TaskRun{}, err
	}
	type Resp struct {
		Data opsv1.TaskRun `json:"data"`
	}

	var resp Resp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Printf("err: %v, resp: %v\n", err, string(body))
	}
	return resp.Data, err
}

func GetTaskRunMarkdown(tr opsv1.TaskRun, builder func(opsv1.TaskRun) string) string {
	if builder == nil {
		builder = buildtTaskRunMarkdown
	}
	return builder(tr)
}

func (m *LLMTaskRunManager) StartUpdateTimer(interval time.Duration, updateFunc func() ([]LLMTask, error)) {

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.tickOnce.Do(func() {
				res, err := updateFunc()
				if err != nil {
					fmt.Printf("timer update hosts err: %v\n", err)
					return
				}
				m.AddTasks(res...)
			})
		}
	}()
	tasks, _ := updateFunc()
	m.AddTasks(tasks...)
}

func (m *LLMTaskRunManager) AddTasks(ts ...LLMTask) *LLMTaskRunManager {
	existings := make(map[string]struct{})
	for _, task := range m.tasks {
		key := fmt.Sprintf("%s-%s", task.Namespace, task.Name)
		existings[key] = struct{}{}
	}

	for _, new := range ts {
		key := fmt.Sprintf("%s-%s", new.Namespace, new.Name)
		if _, exists := existings[key]; !exists {
			m.tasks = append(m.tasks, new)
			existings[key] = struct{}{}
		}
	}

	return m
}

func buildtTaskRunMarkdown(tr opsv1.TaskRun) string {
	var b strings.Builder
	for _, nodeRunStatus := range tr.Status.TaskRunNodeStatus {
		for _, step := range nodeRunStatus.TaskRunStep {
			b.WriteString(fmt.Sprintf("- Step: %s\n", step.StepName))
			if step.StepOutput == "" {
				step.StepOutput = "no output or not found"
			}
			step.StepOutput = strings.ReplaceAll(step.StepOutput, "\n", "\n\n")
			b.WriteString(fmt.Sprintf("- Output:\n%s\n", step.StepOutput))
		}
	}
	return b.String()
}
