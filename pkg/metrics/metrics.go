/*
Copyright 2022 shaowenchen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ControllerReconcileTotal is a counter for the total number of reconcile operations
	ControllerReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_reconcile_total",
			Help: "Total number of reconcile operations",
		},
		[]string{"controller", "namespace", "result"},
	)

	// ControllerReconcileDuration is a histogram for reconcile operation duration
	ControllerReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_controller_reconcile_duration_seconds",
			Help:    "Duration of reconcile operations in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"controller", "namespace"},
	)

	// ControllerReconcileErrors is a counter for reconcile errors
	ControllerReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_reconcile_errors_total",
			Help: "Total number of reconcile errors",
		},
		[]string{"controller", "namespace", "error_type"},
	)

	// TaskRunTotal is a counter for TaskRun operations
	TaskRunTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_taskrun_total",
			Help: "Total number of TaskRun operations",
		},
		[]string{"namespace", "status"},
	)

	// TaskRunDuration is a histogram for TaskRun execution duration
	TaskRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_taskrun_duration_seconds",
			Help:    "Duration of TaskRun execution in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9 hours
		},
		[]string{"namespace", "task_name"},
	)

	// PipelineRunTotal is a counter for PipelineRun operations
	PipelineRunTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_pipelinerun_total",
			Help: "Total number of PipelineRun operations",
		},
		[]string{"namespace", "status"},
	)

	// PipelineRunDuration is a histogram for PipelineRun execution duration
	PipelineRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_pipelinerun_duration_seconds",
			Help:    "Duration of PipelineRun execution in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9 hours
		},
		[]string{"namespace", "pipeline_name"},
	)

	// HostConnectionStatus is a gauge for host connection status
	HostConnectionStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_host_connection_status",
			Help: "Host connection status (1 = connected, 0 = disconnected)",
		},
		[]string{"namespace", "host_name"},
	)

	// ClusterHealthStatus is a gauge for cluster health status
	ClusterHealthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_cluster_health_status",
			Help: "Cluster health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"namespace", "cluster_name"},
	)

	// HTTPRequestsTotal is a counter for HTTP requests
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestDuration is a histogram for HTTP request duration
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~2s
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestSize is a histogram for HTTP request body size
	HTTPRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_http_request_size_bytes",
			Help:    "Size of HTTP request body in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to ~1GB
		},
		[]string{"method", "path"},
	)

	// HTTPResponseSize is a histogram for HTTP response body size
	HTTPResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_http_response_size_bytes",
			Help:    "Size of HTTP response body in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to ~1GB
		},
		[]string{"method", "path", "status_code"},
	)

	// APIRequestsTotal is a counter for API requests
	APIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint", "namespace", "status"},
	)

	// APIRequestDuration is a histogram for API request duration
	APIRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_api_request_duration_seconds",
			Help:    "Duration of API requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~2s
		},
		[]string{"endpoint", "namespace"},
	)

	// APIErrorsTotal is a counter for API errors
	APIErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_api_errors_total",
			Help: "Total number of API errors",
		},
		[]string{"endpoint", "namespace", "error_type"},
	)

	// ServerInfo is a gauge for server information
	ServerInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_server_info",
			Help: "Server information",
		},
		[]string{"version", "build_date"},
	)

	// ServerUptime is a gauge for server uptime
	ServerUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_server_uptime_seconds",
			Help: "Server uptime in seconds",
		},
	)
)

// InitController initializes and registers controller-specific metrics
// with the controller-runtime metrics registry
func InitController() {
	metrics.Registry.MustRegister(
		ControllerReconcileTotal,
		ControllerReconcileDuration,
		ControllerReconcileErrors,
		TaskRunTotal,
		TaskRunDuration,
		PipelineRunTotal,
		PipelineRunDuration,
		HostConnectionStatus,
		ClusterHealthStatus,
	)
}

// InitServer initializes and registers server-specific metrics
// with the controller-runtime metrics registry
func InitServer() {
	metrics.Registry.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestSize,
		HTTPResponseSize,
		APIRequestsTotal,
		APIRequestDuration,
		APIErrorsTotal,
		ServerInfo,
		ServerUptime,
	)
}
