package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opshost "github.com/shaowenchen/ops/pkg/host"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsoption "github.com/shaowenchen/ops/pkg/option"
	opstask "github.com/shaowenchen/ops/pkg/task"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
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
	showData(c, paginator(taskList.Items, req.PageSize, req.Page))
}

func CreateTaskRun(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Task      string `uri:"task"`
	}
	type Body struct {
		NameRef   string            `json:"nameref"`
		TypeRef   string            `json:"typeref"`
		All       bool              `json:"all"`
		Variables map[string]string `json:"variables"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	var body = Body{}
	err = c.ShouldBindJSON(&body)
	if err != nil {
		showError(c, "get body error "+err.Error())
		return
	}
	// validate
	if body.NameRef == "" {
		showError(c, "NameRef is required")
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
	// merge variables
	if body.NameRef != "" {
		task.Spec.NameRef = body.NameRef
	}
	if body.TypeRef != "" {
		task.Spec.TypeRef = body.TypeRef
	}
	for k, v := range body.Variables {
		task.Spec.Variables[k] = v
	}
	if body.All {
		task.Spec.All = body.All
	}

	if task.GetSpec().TypeRef == opsv1.TaskTypeRefHost || task.GetSpec().TypeRef == "" {
		h := &opsv1.Host{}
		err := client.Get(context.TODO(), types.NamespacedName{Namespace: task.GetNamespace(), Name: task.GetSpec().NameRef}, h)
		// if hostname is empty, use localhost
		if len(task.GetSpec().NameRef) > 0 && err != nil {
			showError(c, err.Error())
			return
		}
		cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
		hc, err := opshost.NewHostConnBase64(h)
		if err != nil {
			showError(c, err.Error())
			return
		}
		err = opstask.RunTaskOnHost(cliLogger, task, hc, opsoption.TaskOption{})
		if err != nil {
			showError(c, err.Error())
			return
		}
		res := cliLogger.Flush()
		showData(c, res)
		return

	} else if task.GetSpec().TypeRef == opsv1.TaskTypeRefCluster {
		cluster := &opsv1.Cluster{}
		kubeOpt := opsoption.KubeOption{
			NodeName:     task.GetSpec().NodeName,
			All:          task.GetSpec().All,
			RuntimeImage: task.GetSpec().RuntimeImage,
			OpsNamespace: opsconstants.DefaultOpsNamespace,
		}
		err := client.Get(context.TODO(), types.NamespacedName{Namespace: task.GetNamespace(), Name: task.GetSpec().NameRef}, cluster)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
		kc, err := opskube.NewClusterConnection(cluster)
		if err != nil {
			showError(c, err.Error())
			return
		}
		nodes, err := opskube.GetNodes(cliLogger, kc.Client, kubeOpt)
		if err != nil {
			showError(c, err.Error())
			return
		}
		if len(nodes) == 0 {
			showError(c, "no nodes found")
			return
		}
		for _, node := range nodes {
			opstask.RunTaskOnKube(cliLogger, task, kc, &node, opsoption.TaskOption{}, kubeOpt)
		}
		showData(c, cliLogger.Flush())
		return
	} else {
		showError(c, "unsupported task refType")
		return
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
