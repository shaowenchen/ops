package server

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	v1Tasks := r.Group("/api/v1/namespaces/:namespace/tasks").Use(AuthMiddleware())
	{
		v1Tasks.POST("/", CreateTask)
		v1Tasks.DELETE("/:task", DeleteTask)
		v1Tasks.GET("/", ListTask)
		v1Tasks.GET("/:task", GetTask)
		v1Tasks.POST("/:task/taskruns", CreateTaskRun)
	}
}

func SetHealthzRouter(r *gin.Engine) {
	healthz := r.Group("/api/healthz")
	{
		healthz.GET("/", Healthz)
	}
}
