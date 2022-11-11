package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shaowenchen/ops/pkg/server"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1/task")
	{
		v1.POST("/", server.CreateTask)
		v1.GET("/", server.GetTask)
	}
	healthz := router.Group("/api/healthz")
	{
		healthz.GET("/", server.Healthz)
	}
	return router
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
