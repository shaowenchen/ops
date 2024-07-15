package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}
func ListHost(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
	}
	var req = Params{
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	hostList := &opsv1.HostList{}
	if req.Namespace == "all" {
		err = client.List(context.TODO(), hostList)
	} else {
		err = client.List(context.TODO(), hostList, runtimeClient.InNamespace(req.Namespace))
	}
	if err != nil {
		return
	}
	// clear sensitive info
	for i := range hostList.Items {
		hostList.Items[i].Spec.PrivateKey = ""
		hostList.Items[i].Spec.Password = ""
	}
	showData(c, paginator[opsv1.Host](hostList.Items, req.PageSize, req.Page))
}
func ListCluster(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
	}
	var req = Params{
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	clusterList := &opsv1.ClusterList{}
	if req.Namespace == "all" {
		err = client.List(context.TODO(), clusterList)
	} else {
		err = client.List(context.TODO(), clusterList, runtimeClient.InNamespace(req.Namespace))
	}
	if err != nil {
		return
	}
	// clear sensitive info
	for i := range clusterList.Items {
		clusterList.Items[i].Spec.Token = ""
		clusterList.Items[i].Spec.Config = ""
	}
	showData(c, paginator[opsv1.Cluster](clusterList.Items, req.PageSize, req.Page))
}
func GetTask(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Task      string `uri:"task"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	task := &opsv1.Task{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Task,
	}, task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, task)
	return
}

func GetPipeline(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Pipeline  string `uri:"pipeline"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipeline := &opsv1.Pipeline{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Pipeline,
	}, pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, pipeline)
}

func CreateTask(c *gin.Context) {
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
	}
	task := &opsv1.Task{}
	err = json.Unmarshal(dataBytes, task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Create(context.TODO(), task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

func CreatePipeline(c *gin.Context) {
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
	}
	pipeline := &opsv1.Pipeline{}
	err = json.Unmarshal(dataBytes, pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Create(context.TODO(), pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

func PutTask(c *gin.Context) {
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
	}
	task := &opsv1.Task{}
	err = json.Unmarshal(dataBytes, task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Update(context.TODO(), task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

func PutPipeline(c *gin.Context) {
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
	}
	pipeline := &opsv1.Pipeline{}
	err = json.Unmarshal(dataBytes, pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Update(context.TODO(), pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

func DeleteTask(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Task      string `uri:"task"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	task := &opsv1.Task{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Task,
	}, task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Delete(context.TODO(), task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
	return
}

func DeletePipeline(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Pipeline  string `uri:"pipeline"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipeline := &opsv1.Pipeline{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Pipeline,
	}, pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Delete(context.TODO(), pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
	return
}

func ListTask(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
	}
	var req = Params{
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskList := &opsv1.TaskList{}
	if req.Namespace == "all" {
		err = client.List(context.TODO(), taskList)
	} else {
		err = client.List(context.TODO(), taskList, runtimeClient.InNamespace(req.Namespace))
	}
	if err != nil {
		return
	}
	showData(c, paginator[opsv1.Task](taskList.Items, req.PageSize, req.Page))
}

func ListPipeline(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
	}
	var req = Params{
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipelineList := &opsv1.PipelineList{}
	if req.Namespace == "all" {
		err = client.List(context.TODO(), pipelineList)
	} else {
		err = client.List(context.TODO(), pipelineList, runtimeClient.InNamespace(req.Namespace))
	}
	if err != nil {
		return
	}
	showData(c, paginator[opsv1.Pipeline](pipelineList.Items, req.PageSize, req.Page))
}

func GetTaskRun(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Taskrun   string `uri:"taskrun"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskRun := &opsv1.TaskRun{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Taskrun,
	}, taskRun)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, taskRun)
}

func GetPipelineRun(c *gin.Context) {
	type Params struct {
		Namespace   string `uri:"namespace"`
		Pipelinerun string `uri:"pipelinerun"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipelineRun := &opsv1.PipelineRun{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Pipelinerun,
	}, pipelineRun)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, pipelineRun)
}

func ListTaskRun(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
	}
	var req = Params{
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskRunList := &opsv1.TaskRunList{}
	if req.Namespace == "all" {
		err = client.List(context.TODO(), taskRunList)
	} else {
		err = client.List(context.TODO(), taskRunList, runtimeClient.InNamespace(req.Namespace))
	}
	if err != nil {
		return
	}
	sort.Slice(taskRunList.Items, func(i, j int) bool {
		return taskRunList.Items[i].ObjectMeta.CreationTimestamp.Time.After(taskRunList.Items[j].ObjectMeta.CreationTimestamp.Time)
	})
	// item.status.startTime
	showData(c, paginator[opsv1.TaskRun](taskRunList.Items, req.PageSize, req.Page))
}

func ListPipelineRun(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
	}
	var req = Params{
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipelineRunList := &opsv1.PipelineRunList{}
	if req.Namespace == "all" {
		err = client.List(context.TODO(), pipelineRunList)
	} else {
		err = client.List(context.TODO(), pipelineRunList, runtimeClient.InNamespace(req.Namespace))
	}
	if err != nil {
		return
	}
	showData(c, paginator[opsv1.PipelineRun](pipelineRunList.Items, req.PageSize, req.Page))
}

func CreateTaskRun(c *gin.Context) {
	type Params struct {
		Namespace    string            `uri:"namespace"`
		TaskRef      string            `json:"taskRef"`
		TypeRef      string            `json:"typeRef"`
		NameRef      string            `json:"nameRef"`
		NodeName     string            `json:"nodeName"`
		All          bool              `json:"all"`
		RuntimeImage string            `json:"runtimeImage"`
		Variables    map[string]string `json:"variables"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		showError(c, "get body error "+err.Error())
		return
	}
	// validate
	if req.NameRef == "" {
		showError(c, "nameRef is required")
		return
	}
	if req.TaskRef == "" {
		showError(c, "taskRef is required")
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	// get task
	task := &opsv1.Task{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.TaskRef,
	}, task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskRun := opsv1.NewTaskRun(task)
	taskRun.Spec.NameRef = req.NameRef
	if req.Variables != nil {
		if taskRun.Spec.Variables == nil {
			taskRun.Spec.Variables = make(map[string]string)
		}
		for k, v := range req.Variables {
			taskRun.Spec.Variables[k] = v
		}
	}
	if req.TypeRef != "" {
		taskRun.Spec.TypeRef = req.TypeRef
	}
	if req.NodeName != "" {
		taskRun.Spec.NodeName = req.NodeName
	}
	taskRun.Namespace = req.Namespace
	err = client.Create(context.TODO(), &taskRun)
	if err != nil {
		showError(c, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			latest := &opsv1.TaskRun{}
			err = client.Get(context.TODO(), runtimeClient.ObjectKey{
				Namespace: taskRun.Namespace,
				Name:      taskRun.Name,
			}, latest)
			if err != nil {
				showError(c, err.Error())
				return
			}
			if latest.Status.RunStatus == opsv1.StatusSuccessed || latest.Status.RunStatus == opsv1.StatusFailed || latest.Status.RunStatus == opsv1.StatusAborted {
				showData(c, latest)
				return
			}

		case <-ctx.Done():
			showError(c, "timeout")
			return
		}
	}
}

func CreatePipelineRun(c *gin.Context) {
	type Params struct {
		Namespace   string            `uri:"namespace"`
		PipelineRef string            `json:"pipelineRef"`
		TypeRef     string            `json:"typeRef"`
		NameRef     string            `json:"nameRef"`
		NodeName    string            `json:"nodeName"`
		Variables   map[string]string `json:"variables"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		showError(c, "get body error "+err.Error())
		return
	}
	client, err := getRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	// merge pipeline variables
	pipeline := &opsv1.Pipeline{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.PipelineRef,
	}, pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// create pipelinerun
	pipelinerun := opsv1.NewPipelineRun(pipeline)

	// patch
	if req.NameRef != "" {
		pipelinerun.Spec.NameRef = req.NameRef
	}
	if req.TypeRef != "" {
		pipelinerun.Spec.TypeRef = req.TypeRef
	}
	if req.NodeName != "" {
		pipelinerun.Spec.NodeName = req.NodeName
	}
	if req.Variables != nil {
		pipelinerun.Spec.Variables = req.Variables
	}

	err = client.Create(context.TODO(), &pipelinerun)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// wait
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			latest := &opsv1.PipelineRun{}
			err = client.Get(context.TODO(), runtimeClient.ObjectKey{
				Namespace: pipelinerun.Namespace,
				Name:      pipelinerun.Name,
			}, latest)
			if err != nil {
				showError(c, err.Error())
				return
			}
			if latest.Status.RunStatus == opsv1.StatusSuccessed || latest.Status.RunStatus == opsv1.StatusFailed || latest.Status.RunStatus == opsv1.StatusAborted {
				showData(c, latest)
				return
			}

		case <-ctx.Done():
			showError(c, "timeout")
			return
		}
	}
}

func getRuntimeClient(kubeconfigPath string) (client runtimeClient.Client, err error) {
	scheme, err := opsv1.SchemeBuilder.Build()
	if err != nil {
		return
	}
	restConfig, err := opsutils.GetRestConfig(kubeconfigPath)

	if err != nil {
		return
	}

	return runtimeClient.New(restConfig, runtimeClient.Options{Scheme: scheme})
}
