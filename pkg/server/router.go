package server

import (
	"github.com/gin-gonic/gin"
	_ "github.com/shaowenchen/ops/swagger"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(r *gin.Engine) {
	v1Hosts := r.Group("/api/v1/namespaces/:namespace/hosts").Use(AuthMiddleware())
	{
		v1Hosts.GET("", ListHosts)
		v1Hosts.POST("", CreateHost)
		v1Hosts.GET("/:host", GetHost)
		v1Hosts.PUT("/:host", PutHost)
		v1Hosts.DELETE("/:host", DeleteHost)
	}
	v1Clusters := r.Group("/api/v1/namespaces/:namespace/clusters").Use(AuthMiddleware())
	{
		v1Clusters.GET("", ListClusters)
		v1Clusters.POST("", CreateCluster)
		v1Clusters.GET("/:cluster", GetCluster)
		v1Clusters.PUT("/:cluster", PutCluster)
		v1Clusters.DELETE("/:cluster", DeleteCluster)
		v1Clusters.GET("/:cluster/nodes", GetClusterNodes)
	}
	v1Tasks := r.Group("/api/v1/namespaces/:namespace/tasks").Use(AuthMiddleware())
	{
		v1Tasks.GET("", ListTasks)
		v1Tasks.POST("", CreateTask)
		v1Tasks.GET("/:task", GetTask)
		v1Tasks.PUT("/:task", PutTask)
		v1Tasks.DELETE("/:task", DeleteTask)
	}
	v1Taskruns := r.Group("/api/v1/namespaces/:namespace/taskruns").Use(AuthMiddleware())
	{
		v1Taskruns.GET("", ListTaskRun)
		v1Taskruns.POST("", CreateTaskRun)
		v1Taskruns.POST("/sync", CreateTaskRunSync)
		v1Taskruns.GET("/:taskrun", GetTaskRun)
	}
	v1Pipelines := r.Group("/api/v1/namespaces/:namespace/pipelines").Use(AuthMiddleware())
	{
		v1Pipelines.GET("", ListPipelines)
		v1Pipelines.POST("", CreatePipeline)
		v1Pipelines.GET("/:pipeline", GetPipeline)
		v1Pipelines.PUT("/:pipeline", PutPipeline)
		v1Pipelines.DELETE("/:pipeline", DeletePipeline)
	}
	v1Pipelineruns := r.Group("/api/v1/namespaces/:namespace/pipelineruns").Use(AuthMiddleware())
	{
		v1Pipelineruns.GET("", ListPipelineRuns)
		v1Pipelineruns.POST("", CreatePipelineRun)
		v1Pipelineruns.POST("/sync", CreatePipelineRunSync)
		v1Pipelineruns.GET("/:pipelinerun", GetPipelineRun)
	}
	v1Login := r.Group("/api/v1/login").Use(AuthMiddleware())
	{
		v1Login.GET("/check", LoginCheck)
	}
	v1Summary := r.Group("/api/v1/summary").Use(AuthMiddleware())
	{
		v1Summary.GET("", GetSummary)
	}
	v1Namespaces := r.Group("/api/v1/namespaces").Use(AuthMiddleware())
	{
		v1Namespaces.GET("", ListNamespaces)
	}
	v1Events := r.Group("/api/v1/events").Use()
	{
		v1Events.GET("", ListEvents)
		v1Events.GET("/:event", GetEvents)
	}
	v1EventHooks := r.Group("/api/v1/namespaces/:namespace/eventhooks").Use(AuthMiddleware())
	{
		v1EventHooks.GET("", ListEventHooks)
		v1EventHooks.GET("/:eventhook", GetEventHook)
		v1EventHooks.POST("", CreateEventHook)
		v1EventHooks.PUT("/:eventhook", PutEventHook)
		v1EventHooks.DELETE("/:eventhook", DeleteEventHook)
	}
}

func SetupRouteWithoutAuth(r *gin.Engine) {
	v1Events := r.Group("/api/v1/namespaces/:namespace/events")
	{
		v1Events.POST("/:event", CreateEvent)
	}
	r.GET("/api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

func SetHealthzRouter(r *gin.Engine) {
	root := r.Group("/")
	{
		root.GET("healthz", Healthz)
		root.GET("readyz", Healthz)
	}
}
