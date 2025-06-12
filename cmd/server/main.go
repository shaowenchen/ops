package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/shaowenchen/ops/pkg/server"
	"github.com/shaowenchen/ops/web"
)

func init() {
	configpath := flag.String("c", "", "")
	flag.Parse()
	server.LoadConfig(*configpath)
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()
}

func main() {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz", "/readyz"},
	}))

	gin.SetMode(server.GlobalConfig.Server.RunMode)
	server.SetupRouter(r)
	server.SetupRouteWithoutAuth(r)
	server.SetHealthzRouter(r)
	web.SetupRouter(r)
	r.Run(":80")
}
