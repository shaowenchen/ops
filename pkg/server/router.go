package server

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
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
	healthz := r.Group("/api/healthz")
	{
		healthz.GET("/", Healthz)
	}
}
