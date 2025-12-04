package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shaowenchen/ops/pkg/metrics"
	"github.com/shaowenchen/ops/pkg/server"
	"github.com/shaowenchen/ops/web"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
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
	// Initialize server metrics
	metrics.InitServer()

	// Set server info
	metrics.ServerInfo.WithLabelValues("unknown", "unknown").Set(1)

	// Start uptime tracking
	startTime := time.Now()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics.ServerUptime.Set(time.Since(startTime).Seconds())
		}
	}()

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz", "/readyz", "/metrics"},
	}))

	// Add Prometheus metrics middleware
	r.Use(metrics.PrometheusMiddleware())

	gin.SetMode(server.GlobalConfig.Server.RunMode)

	// Add metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(ctrlmetrics.Registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
	})))

	server.SetupRouter(r)
	server.SetupRouteWithoutAuth(r)
	server.SetHealthzRouter(r)
	web.SetupRouter(r)
	r.Run(":80")
}
