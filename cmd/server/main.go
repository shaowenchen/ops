package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/shaowenchen/ops/pkg/server"
	"github.com/shaowenchen/ops/web"
)

func init() {
	configpath := flag.String("c", "", "")
	flag.Parse()
	server.LoadConfig(*configpath)
}

func main() {
	r := gin.Default()
	gin.SetMode(server.GlobalConfig.Server.RunMode)
	server.SetupRouter(r)
	server.SetupRouteWithoutAuth(r)
	server.SetHealthzRouter(r)
	web.SetupRouter(r)
	r.Run(":80")
}
