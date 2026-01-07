package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// @Summary Health Check
// @Tags Health
// @Accept json
// @Produce json
// @Success 200
// @Router /healthz [get]
func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

// @Summary List Namespaces
// @Tags Namespaces
// @Accept json
// @Produce json
// @Success 200
// @Router /api/v1/namespaces [get]
func ListNamespaces(c *gin.Context) {
	client, err := opsutils.NewKubernetesClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	namespaceList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		showError(c, err.Error())
		return
	}
	namespaces := make([]string, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}
	sort.Strings(namespaces)
	showData(c, namespaces)
}

// @Summary List Hosts
// @Tags Hosts
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/hosts [get]
func ListHosts(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
		Search    string `form:"search"`
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
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	hostList := &opsv1.HostList{}
	err = client.List(context.TODO(), hostList, runtimeClient.InNamespace(req.Namespace))
	if err != nil {
		return
	}
	newHosts := make([]opsv1.Host, 0)
	// search
	if req.Search != "" {
		for i := range hostList.Items {
			searchField := []string{hostList.Items[i].Name, hostList.Items[i].Spec.Address, hostList.Items[i].Status.Hostname, hostList.Items[i].Status.AcceleratorModel, hostList.Items[i].Status.AcceleratorVendor}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newHosts = append(newHosts, hostList.Items[i])
					break
				}
			}
		}
	} else {
		newHosts = hostList.Items
	}
	// clear info
	for i := range newHosts {
		newHosts[i].Cleaned()
	}
	showData(c, paginator[opsv1.Host](newHosts, req.PageSize, req.Page))
}

// @Summary Get Host
// @Tags Hosts
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param host path string true "host"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/hosts/{host} [get]
func GetHost(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Host      string `uri:"host"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	host := &opsv1.Host{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Host,
	}, host)
	if err != nil {
		showError(c, err.Error())
		return
	}
	host.Cleaned()
	showData(c, host)
}

// @Summary Update Host
// @Tags Hosts
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param host path string true "host"
// @Param host body opsv1.Host true "host"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/hosts/{host} [put]
func PutHost(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Host      string `uri:"host"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
		return
	}
	host := &opsv1.Host{}
	err = json.Unmarshal(dataBytes, host)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Get existing host to preserve ResourceVersion
	oldHost := &opsv1.Host{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Host,
	}, oldHost)
	if err != nil {
		showError(c, err.Error())
		return
	}
	host.ObjectMeta.ResourceVersion = oldHost.ObjectMeta.ResourceVersion
	err = client.Update(context.TODO(), host)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Delete Host
// @Tags Hosts
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param host path string true "host"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/hosts/{host} [delete]
func DeleteHost(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Host      string `uri:"host"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	host := &opsv1.Host{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Host,
	}, host)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Delete(context.TODO(), host)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Create Host
// @Tags Hosts
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param host body opsv1.Host true "host"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/hosts [post]
func CreateHost(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
		return
	}
	host := &opsv1.Host{}
	err = json.Unmarshal(dataBytes, host)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Set namespace from URI if not set in body
	if host.Namespace == "" {
		host.Namespace = req.Namespace
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Create(context.TODO(), host)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary List Clusters
// @Tags Clusters
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/clusters [get]
func ListClusters(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
		Search    string `form:"search"`
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
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	clusterList := &opsv1.ClusterList{}
	err = client.List(context.TODO(), clusterList, runtimeClient.InNamespace(req.Namespace))
	if err != nil {
		return
	}
	newCluster := make([]opsv1.Cluster, 0)
	// search
	if req.Search != "" {
		for i := range clusterList.Items {
			searchField := []string{clusterList.Items[i].Name, clusterList.Items[i].Spec.Server, clusterList.Items[i].Spec.Desc}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newCluster = append(newCluster, clusterList.Items[i])
					break
				}
			}
		}
	} else {
		newCluster = clusterList.Items
	}
	// clear info
	for i := range newCluster {
		newCluster[i].Cleaned()
	}
	showData(c, paginator[opsv1.Cluster](newCluster, req.PageSize, req.Page))
}

// @Summary Get Cluster
// @Tags Clusters
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param cluster path string true "cluster"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/clusters/{cluster} [get]
func GetCluster(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Cluster   string `uri:"cluster"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	cluster := &opsv1.Cluster{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Cluster,
	}, cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// hide info
	cluster.Cleaned()
	showData(c, cluster)
}

// @Summary Update Cluster
// @Tags Clusters
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param cluster path string true "cluster"
// @Param cluster body opsv1.Cluster true "cluster"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/clusters/{cluster} [put]
func PutCluster(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Cluster   string `uri:"cluster"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
		return
	}
	cluster := &opsv1.Cluster{}
	err = json.Unmarshal(dataBytes, cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Get existing cluster to preserve ResourceVersion
	oldCluster := &opsv1.Cluster{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Cluster,
	}, oldCluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	cluster.ObjectMeta.ResourceVersion = oldCluster.ObjectMeta.ResourceVersion
	err = client.Update(context.TODO(), cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Delete Cluster
// @Tags Clusters
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param cluster path string true "cluster"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/clusters/{cluster} [delete]
func DeleteCluster(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Cluster   string `uri:"cluster"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	cluster := &opsv1.Cluster{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Cluster,
	}, cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Delete(context.TODO(), cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Create Cluster
// @Tags Clusters
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param cluster body opsv1.Cluster true "cluster"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/clusters [post]
func CreateCluster(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
		return
	}
	cluster := &opsv1.Cluster{}
	err = json.Unmarshal(dataBytes, cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Set namespace from URI if not set in body
	if cluster.Namespace == "" {
		cluster.Namespace = req.Namespace
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Create(context.TODO(), cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Get Cluster Nodes
// @Tags Clusters
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param cluster path string true "cluster"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/clusters/{cluster}/nodes [get]
func GetClusterNodes(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Cluster   string `uri:"cluster"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
		Search    string `form:"search"`
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
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	cluster := &opsv1.Cluster{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Cluster,
	}, cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	kc, err := opskube.NewClusterConnection(cluster)
	objs, err := kc.GetNodes()
	if err != nil {
		showError(c, err.Error())
		return
	}
	newObjs := make([]corev1.Node, 0)
	// search
	if req.Search != "" {
		for i := range objs.Items {
			searchField := []string{objs.Items[i].Name, opsutils.GetNodeInternalIp(&objs.Items[i])}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newObjs = append(newObjs, objs.Items[i])
					break
				}
			}
		}
	} else {
		newObjs = objs.Items
	}

	showData(c, paginator[corev1.Node](newObjs, req.PageSize, req.Page))
}

// @Summary Get Task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param task path string true "task"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/tasks/{task} [get]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Get Pipeline
// @Tags Pipelines
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param pipeline path string true "pipeline"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelines/{pipeline} [get]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Create Task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param task body opsv1.Task true "task"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/tasks [post]
func CreateTask(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
		return
	}
	task := &opsv1.Task{}
	err = json.Unmarshal(dataBytes, task)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Set namespace from URI if not set in body
	if task.Namespace == "" {
		task.Namespace = req.Namespace
	}
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Create Pipeline
// @Tags Pipelines
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param pipeline body opsv1.Pipeline true "pipeline"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelines [post]
func CreatePipeline(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipeline := &opsv1.Pipeline{}
	err = json.Unmarshal(dataBytes, pipeline)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Set namespace from URI if not set in body
	if pipeline.Namespace == "" {
		pipeline.Namespace = req.Namespace
	}
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Update Task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param task body opsv1.Task true "task"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/tasks [put]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Update Pipeline
// @Tags Pipelines
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param pipeline body opsv1.Pipeline true "pipeline"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelines [put]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Delete Task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param task path string true "task"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/tasks/{task} [delete]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Delete Pipeline
// @Tags Pipelines
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param pipeline path string true "pipeline"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelines/{pipeline} [delete]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary List Tasks
// @Tags Tasks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param search query string false "search"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/tasks [get]
func ListTasks(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
		Search    string `form:"search"`
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
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskList := &opsv1.TaskList{}
	err = client.List(context.TODO(), taskList, runtimeClient.InNamespace(req.Namespace))
	if err != nil {
		return
	}
	newTaskList := make([]opsv1.Task, 0)
	// search
	if req.Search != "" {
		for i := range taskList.Items {
			searchField := []string{taskList.Items[i].Name, taskList.Items[i].Spec.Desc}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newTaskList = append(newTaskList, taskList.Items[i])
					break
				}
			}
		}
	} else {
		newTaskList = taskList.Items
	}
	showData(c, paginator[opsv1.Task](newTaskList, req.PageSize, req.Page))
}

// @Summary List Pipelines
// @Tags Pipelines
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param search query string false "search"
// @Param labels_selector query string false "labels_selector"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelines [get]
func ListPipelines(c *gin.Context) {
	type Params struct {
		Namespace      string `uri:"namespace"`
		Page           uint   `form:"page"`
		PageSize       uint   `form:"page_size"`
		Search         string `form:"search"`
		LabelsSelector string `form:"labels_selector"`
	}
	var req = Params{
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindUri(&req)
	if err != nil {
		return
	}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		return
	}
	// labels
	labels := make(map[string]string)
	labelSelectorKeyValues := strings.Split(req.LabelsSelector, ",")
	for i := range labelSelectorKeyValues {
		keyValue := strings.Split(labelSelectorKeyValues[i], "=")
		if len(keyValue) == 2 {
			labels[keyValue[0]] = keyValue[1]
		}
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		return
	}
	pipelineList := &opsv1.PipelineList{}
	err = client.List(context.TODO(), pipelineList, runtimeClient.InNamespace(req.Namespace), runtimeClient.MatchingLabels(labels))
	if err != nil {
		return
	}
	newPipelineList := make([]opsv1.Pipeline, 0)
	// search
	if req.Search != "" {
		for i := range pipelineList.Items {
			searchField := []string{pipelineList.Items[i].Name, pipelineList.Items[i].Spec.Desc}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newPipelineList = append(newPipelineList, pipelineList.Items[i])
					break
				}
			}
		}
	} else {
		newPipelineList = pipelineList.Items
	}
	showData(c, paginator[opsv1.Pipeline](newPipelineList, req.PageSize, req.Page))
}

// @Summary Get TaskRun
// @Tags TaskRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param taskrun path string true "taskrun"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/taskruns/{taskrun} [get]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary Get PipelineRun
// @Tags PipelineRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param pipelinerun path string true "pipelinerun"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelineruns/{pipelinerun} [get]
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
	client, err := opskube.GetRuntimeClient("")
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

// @Summary List TaskRun
// @Tags TaskRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/taskruns [get]
func ListTaskRun(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
		Search    string `form:"search"`
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
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskRunList := &opsv1.TaskRunList{}
	err = client.List(context.TODO(), taskRunList, runtimeClient.InNamespace(req.Namespace))
	if err != nil {
		return
	}

	sort.Slice(taskRunList.Items, func(i, j int) bool {
		return taskRunList.Items[i].ObjectMeta.CreationTimestamp.Time.After(taskRunList.Items[j].ObjectMeta.CreationTimestamp.Time)
	})
	newTaskRunList := make([]opsv1.TaskRun, 0)
	// search
	if req.Search != "" {
		for i := range taskRunList.Items {
			searchField := []string{taskRunList.Items[i].Name, taskRunList.Items[i].Spec.TaskRef, taskRunList.Items[i].Status.RunStatus}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newTaskRunList = append(newTaskRunList, taskRunList.Items[i])
					break
				}
			}
		}
	} else {
		newTaskRunList = taskRunList.Items
	}
	// item.status.startTime
	showData(c, paginator[opsv1.TaskRun](newTaskRunList, req.PageSize, req.Page))
}

// @Summary List PipelineRun
// @Tags PipelineRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelineruns [get]
func ListPipelineRuns(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
		Search    string `form:"search"`
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
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	objs := &opsv1.PipelineRunList{}
	err = client.List(context.TODO(), objs, runtimeClient.InNamespace(req.Namespace))
	if err != nil {
		return
	}

	sort.Slice(objs.Items, func(i, j int) bool {
		return objs.Items[i].ObjectMeta.CreationTimestamp.Time.After(objs.Items[j].ObjectMeta.CreationTimestamp.Time)
	})
	newPipelineRunList := make([]opsv1.PipelineRun, 0)
	// search
	if req.Search != "" {
		for i := range objs.Items {
			searchField := []string{objs.Items[i].Name, objs.Items[i].Spec.PipelineRef, objs.Items[i].Status.RunStatus}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newPipelineRunList = append(newPipelineRunList, objs.Items[i])
					break
				}
			}
		}
	} else {
		newPipelineRunList = objs.Items
	}
	showData(c, paginator[opsv1.PipelineRun](newPipelineRunList, req.PageSize, req.Page))
}

// @Summary Create TaskRun
// @Tags TaskRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param taskRef body string true "taskRef"
// @Param variables body map[string]string true "variables"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/taskruns [post]
func CreateTaskRun(c *gin.Context) {
	latest, err := createTaskRun(c, false)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, latest.CopyWithOutVersion())
}

// @Summary Create TaskRun Sync
// @Tags TaskRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param taskRef body string true "taskRef"
// @Param variables body map[string]string true "variables"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/taskruns/sync [post]
func CreateTaskRunSync(c *gin.Context) {
	latest, err := createTaskRun(c, true)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, latest.CopyWithOutVersion())
}

func createTaskRun(c *gin.Context, sync bool) (latest opsv1.TaskRun, err error) {
	type Params struct {
		Namespace string            `uri:"namespace"`
		TaskRef   string            `json:"taskRef"`
		Variables map[string]string `json:"variables"`
	}
	var req = Params{}
	err = c.ShouldBindUri(&req)
	if err != nil {
		return
	}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}
	if req.TaskRef == "" {
		err = errors.New("taskRef is required")
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		return
	}
	// get task
	task := &opsv1.Task{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.TaskRef,
	}, task)
	if err != nil {
		return
	}
	taskRun := opsv1.NewTaskRun(task)
	if req.Variables != nil {
		if taskRun.Spec.Variables == nil {
			taskRun.Spec.Variables = make(map[string]string)
		}
		for k, v := range req.Variables {
			taskRun.Spec.Variables[k] = v
		}
	}
	taskRun.Namespace = req.Namespace
	err = client.Create(context.TODO(), &taskRun)
	if err != nil {
		return
	}
	if !sync {
		latest = taskRun
		return
	}
	// wait done
	// Use request context so we can cancel if client disconnects
	reqCtx := c.Request.Context()
	ctx, cancel := context.WithTimeout(reqCtx, 600*time.Second)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			latest = opsv1.TaskRun{}
			// Use the timeout context instead of context.TODO() so the request can be cancelled
			err = client.Get(ctx, runtimeClient.ObjectKey{
				Namespace: taskRun.Namespace,
				Name:      taskRun.Name,
			}, &latest)
			if err != nil {
				return
			}
			if latest.Status.RunStatus == opsconstants.StatusSuccessed || latest.Status.RunStatus == opsconstants.StatusFailed || latest.Status.RunStatus == opsconstants.StatusAborted || latest.Status.RunStatus == opsconstants.StatusDataInValid {
				return
			}

		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				err = errors.New("request cancelled")
			} else {
				err = errors.New("timeout")
			}
			return
		}
	}
}

// @Summary Create PipelineRun
// @Tags PipelineRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param pipelineRef body string true "pipelineRef"
// @Param variables body map[string]string true "variables"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelineruns [post]
func CreatePipelineRun(c *gin.Context) {
	latest, err := createPipelineRun(c, false)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, latest.CopyWithOutVersion())
}

// @Summary Create PipelineRun Sync
// @Tags PipelineRuns
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param pipelineRef body string true "pipelineRef"
// @Param variables body map[string]string true "variables"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/pipelineruns/sync [post]
func CreatePipelineRunSync(c *gin.Context) {
	latest, err := createPipelineRun(c, true)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, latest.CopyWithOutVersion())
}

func createPipelineRun(c *gin.Context, sync bool) (latest opsv1.PipelineRun, err error) {
	type Params struct {
		Namespace   string            `uri:"namespace"`
		PipelineRef string            `json:"pipelineRef"`
		Variables   map[string]string `json:"variables"`
	}
	var req = Params{}
	err = c.ShouldBindUri(&req)
	if err != nil {
		return
	}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		return
	}
	// merge pipeline variables
	pipeline := &opsv1.Pipeline{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.PipelineRef,
	}, pipeline)
	if err != nil {
		return
	}
	// create pipelinerun
	pipelinerun := opsv1.NewPipelineRun(pipeline)
	if req.Variables != nil {
		pipelinerun.Spec.Variables = req.Variables
	}

	err = client.Create(context.TODO(), pipelinerun)
	if err != nil {
		return
	}

	if !sync {
		latest = *pipelinerun
		return
	}
	// wait done
	// Use request context so we can cancel if client disconnects
	reqCtx := c.Request.Context()
	ctx, cancel := context.WithTimeout(reqCtx, 600*time.Second)
	defer cancel()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			latest = opsv1.PipelineRun{}
			// Use the timeout context instead of context.TODO() so the request can be cancelled
			err = client.Get(ctx, runtimeClient.ObjectKey{
				Namespace: pipelinerun.Namespace,
				Name:      pipelinerun.Name,
			}, &latest)
			if err != nil {
				return
			}
			if latest.Status.RunStatus == opsconstants.StatusSuccessed || latest.Status.RunStatus == opsconstants.StatusFailed || latest.Status.RunStatus == opsconstants.StatusAborted {
				return
			}

		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				err = errors.New("request cancelled")
			} else {
				err = errors.New("timeout")
			}
			return
		}
	}
}

// @Summary Create Event
// @Tags Events
// @Accept json
// @Produce json
// @Param event path string true "event"
// @Param        body   body      map[string]interface{}          true  "Event payload"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/events/{event} [post]
func CreateEvent(c *gin.Context) {
	type Params struct {
		Event     string `uri:"event"`
		Namespace string `uri:"namespace"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}

	body := make(map[string]interface{})
	err = c.ShouldBindJSON(&body)

	if err != nil {
		showError(c, err.Error())
		return
	}

	// Publish event asynchronously to avoid blocking the HTTP response
	// Use a goroutine with timeout to prevent goroutine leaks
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var eventBus *opsevent.EventBus
		// If event path starts with "ops.", use it directly as subject (e.g., notification events)
		// Format: ops.notifications.providers.{provider}.channels.{channel}.severities.{severity}
		if strings.HasPrefix(req.Event, "ops.") {
			// Use the event path directly as subject without any transformation
			eventBus = (&opsevent.EventBus{}).WithEndpoint(GlobalConfig.Event.Endpoint).WithSubject(req.Event)
		} else if strings.HasPrefix(req.Event, "nodes.") {
			// Handle node events specially: nodes.{nodeName}.{observation} format
			// Node events use special format without namespaces: ops.clusters.{cluster}.nodes.{nodeName}.{observation}
			// Split event path: "nodes.{nodeName}.findings" -> ["nodes", "{nodeName}", "findings"]
			// Use FactoryKube with empty namespace for node events
			parts := strings.Split(req.Event, ".")
			eventBus = opsevent.FactoryKube("", parts...)
		} else {
			// Standard format: ops.clusters.{cluster}.namespaces.{namespace}.{event}
			eventBus = opsevent.Factory(GlobalConfig.Event.Endpoint, GlobalConfig.Event.Cluster, req.Namespace, req.Event)
		}
		// Publish already has defer Close inside, so we don't need to call Close again
		// The context timeout ensures this goroutine will exit even if Publish blocks
		_ = eventBus.Publish(ctx, body)
		// Goroutine will exit here, and context will be cancelled by defer
	}()

	showSuccess(c)
}

// @Summary Get Events
// @Tags Event
// @Accept json
// @Produce json
// @Param event path string true "event"
// @Parm max_length query int false "max_length"
// @Param TimeOut query int false "timeout"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Success 200
// @Router /api/v1/events/{event} [get]
func GetEvents(c *gin.Context) {
	type Params struct {
		Event     string `uri:"event"`
		StartTime int64  `form:"start_time"`
		MaxLength uint   `form:"max_length"`
		TimeOut   uint   `form:"timeout"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
	}
	// Get timeout with priority: env var > config file > default
	defaultTimeout := opsconstants.GetEventQueryTimeout(GlobalConfig.Event.QueryTimeout)
	var req = Params{
		PageSize:  10,
		Page:      1,
		MaxLength: 1000,
		Event:     "ops.>",
		TimeOut:   defaultTimeout,
		StartTime: time.Now().Add(-time.Minute).UnixMicro(), // Default to 1 minute ago
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
	// If timeout is not specified or is 0, use default from config
	if req.TimeOut == 0 {
		req.TimeOut = defaultTimeout
	}
	sec := req.StartTime / 1000

	t := time.Unix(sec, 0).UTC()
	location := time.FixedZone("CST", 8*60*60)
	startTime := t.In(location)
	if err != nil {
		startTime = time.Now().Add(-time.Hour * 1)
	}
	client, err := opsevent.FactoryJetStreamClient(GlobalConfig.Event.Endpoint, GlobalConfig.Event.Cluster)
	if err != nil {
		showError(c, err.Error())
		return
	}
	data, err := opsevent.QueryStartTime(client, req.Event, startTime, req.MaxLength, req.TimeOut)
	showData(c, paginator[opsevent.EventData](data, req.PageSize, req.Page))
}

// @Summary List Events
// @Tags Event
// @Accept json
// @Produce json
// @Param search query string false "search"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Success 200
// @Router /api/v1/events [get]
func ListEvents(c *gin.Context) {
	type Params struct {
		Search   string `form:"search"`
		Page     uint   `form:"page"`
		PageSize uint   `form:"page_size"`
	}
	var req = Params{
		Search:   "",
		PageSize: 10,
		Page:     1,
	}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Get timeout with priority: env var > config file > default
	timeout := opsconstants.GetEventListSubjectsTimeout(GlobalConfig.Event.ListSubjectsTimeout)
	data, err := opsevent.ListSubjects(GlobalConfig.Event.Endpoint, opsconstants.OpsStreamName, req.Search, timeout)
	showData(c, paginator[string](data, req.PageSize, req.Page))
}

// @Summary List EventHooks
// @Tags EventHooks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param search query string false "search"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/eventhooks [get]
func ListEventHooks(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Page      uint   `form:"page"`
		PageSize  uint   `form:"page_size"`
		Search    string `form:"search"`
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
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	eventhooks := &opsv1.EventHooksList{}
	err = client.List(context.TODO(), eventhooks, runtimeClient.InNamespace(req.Namespace))
	if err != nil {
		showError(c, err.Error())
		return
	}
	newEventHooks := make([]opsv1.EventHooks, 0)
	// search
	if req.Search != "" {
		for i := range eventhooks.Items {
			searchField := []string{eventhooks.Items[i].Name, eventhooks.Items[i].Spec.Type, eventhooks.Items[i].Spec.Subject, eventhooks.Items[i].Spec.URL}
			for j := range searchField {
				if opsutils.Contains(searchField[j], req.Search) {
					newEventHooks = append(newEventHooks, eventhooks.Items[i])
					break
				}
			}
		}
	} else {
		newEventHooks = eventhooks.Items
	}
	showData(c, paginator[opsv1.EventHooks](newEventHooks, req.PageSize, req.Page))
}

// @Summary Get EventHook
// @Tags EventHooks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param eventhook path string true "eventhook"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/eventhooks/{eventhook} [get]
func GetEventHook(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Eventhook string `uri:"eventhook"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	eventhook := &opsv1.EventHooks{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Eventhook,
	}, eventhook)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, eventhook)
}

// @Summary Create EventHook
// @Tags EventHooks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param eventhook body opsv1.EventHooks true "eventhook"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/eventhooks [post]
func CreateEventHook(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	dataBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		showError(c, err.Error())
		return
	}
	eventhook := &opsv1.EventHooks{}
	err = json.Unmarshal(dataBytes, eventhook)
	if err != nil {
		showError(c, err.Error())
		return
	}
	// Set namespace from URI if not set in body
	if eventhook.Namespace == "" {
		eventhook.Namespace = req.Namespace
	}
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Create(context.TODO(), eventhook)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Update EventHook
// @Tags EventHooks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param eventhook path string true "eventhook"
// @Param eventhook body opsv1.EventHooks true "eventhook"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/eventhooks/{eventhook} [put]
func PutEventHook(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Eventhook string `uri:"eventhook"`
	}
	var eventhook = opsv1.EventHooks{}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = c.ShouldBindJSON(&eventhook)
	if err != nil {
		showError(c, err.Error())
		return
	}
	eventhook.ObjectMeta.Namespace = req.Namespace
	eventhook.ObjectMeta.Name = req.Eventhook
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	oldEventhook := &opsv1.EventHooks{}
	err = client.Get(context.TODO(), runtimeClient.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Eventhook,
	}, oldEventhook)
	if err != nil {
		showError(c, err.Error())
		return
	}
	eventhook.ObjectMeta.ResourceVersion = oldEventhook.ObjectMeta.ResourceVersion
	err = client.Update(context.TODO(), &eventhook)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Delete EventHook
// @Tags EventHooks
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param eventhook path string true "eventhook"
// @Success 200
// @Router /api/v1/namespaces/{namespace}/eventhooks/{eventhook} [delete]
func DeleteEventHook(c *gin.Context) {
	type Params struct {
		Namespace string `uri:"namespace"`
		Eventhook string `uri:"eventhook"`
	}
	var req = Params{}
	err := c.ShouldBindUri(&req)
	if err != nil {
		showError(c, err.Error())
		return
	}
	var eventhook = opsv1.EventHooks{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      req.Eventhook,
		},
	}

	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	err = client.Delete(context.TODO(), &eventhook)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showSuccess(c)
}

// @Summary Login Check
// @Tags Login
// @Accept json
// @Produce json
// @Success 200
// @Router /api/v1/login/check [get]
func LoginCheck(c *gin.Context) {
	showSuccess(c)
}

// @Summary Get Summary
// @Tags Summary
// @Accept json
// @Produce json
// @Success 200
// @Router /api/v1/summary [get]
func GetSummary(c *gin.Context) {
	client, err := opskube.GetRuntimeClient("")
	if err != nil {
		showError(c, err.Error())
		return
	}
	clusterList := &opsv1.ClusterList{}
	err = client.List(context.TODO(), clusterList)
	if err != nil {
		showError(c, err.Error())
		return
	}
	hostList := &opsv1.HostList{}
	err = client.List(context.TODO(), hostList)
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipelineList := &opsv1.PipelineList{}
	err = client.List(context.TODO(), pipelineList)
	if err != nil {
		showError(c, err.Error())
		return
	}
	pipelinerunList := &opsv1.PipelineRunList{}
	err = client.List(context.TODO(), pipelinerunList)
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskList := &opsv1.TaskList{}
	err = client.List(context.TODO(), taskList)
	if err != nil {
		showError(c, err.Error())
		return
	}
	taskrunList := &opsv1.TaskRunList{}
	err = client.List(context.TODO(), taskrunList)
	if err != nil {
		showError(c, err.Error())
		return
	}
	eventhooksList := &opsv1.EventHooksList{}
	err = client.List(context.TODO(), eventhooksList)
	if err != nil {
		showError(c, err.Error())
		return
	}
	showData(c, gin.H{
		"clusters":     len(clusterList.Items),
		"hosts":        len(hostList.Items),
		"pipelines":    len(pipelineList.Items),
		"pipelineruns": len(pipelinerunList.Items),
		"tasks":        len(taskList.Items),
		"taskruns":     len(taskrunList.Items),
		"eventhooks":   len(eventhooksList.Items),
	})
}
