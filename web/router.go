package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupRouter(r *gin.Engine) {
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")
	r.StaticFile("/logo.png", "./web/dist/logo.png")
	r.Static("/assets", "./web/dist/assets")
	r.LoadHTMLFiles("./web/dist/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
}
