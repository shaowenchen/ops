package server

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	v1Hosts := r.Group("/api/v1/namespaces/:namespace/hosts").Use(AuthMiddleware())
	{
		v1Hosts.GET("", ListHost)
	}
	v1Clusters := r.Group("/api/v1/namespaces/:namespace/clusters").Use(AuthMiddleware())
	{
		v1Clusters.GET("", ListCluster)
	}
	v1Tasks := r.Group("/api/v1/namespaces/:namespace/tasks").Use(AuthMiddleware())
	{
		v1Tasks.GET("", ListTask)
		v1Tasks.POST("", CreateTask)
		v1Tasks.GET("/:task", GetTask)
		v1Tasks.PUT("/:task", PutTask)
		v1Tasks.DELETE("/:task", DeleteTask)
	}
	v1Taskruns := r.Group("/api/v1/namespaces/:namespace/taskruns").Use(AuthMiddleware())
	{
		v1Taskruns.GET("", ListTaskRun)
		v1Taskruns.POST("", CreateTaskRun)
		v1Taskruns.GET("/:taskrun", GetTaskRun)
	}
}

func SetHealthzRouter(r *gin.Engine) {
	root := r.Group("/")
	{
		root.GET("healthz", Healthz)
		root.GET("readyz", Healthz)
	}
}
