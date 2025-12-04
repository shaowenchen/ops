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
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ============================================================================
	// 基本的资源使用指标（内存、CPU、协程、网络等）
	// ============================================================================

	// ControllerMemoryAllocBytes is a gauge for memory allocated in bytes
	ControllerMemoryAllocBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_controller_resource_memory_alloc_bytes",
			Help: "Memory allocated in bytes",
		},
	)

	// ControllerMemoryTotalAllocBytes is a counter for total memory allocated in bytes
	ControllerMemoryTotalAllocBytes = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ops_controller_resource_memory_total_alloc_bytes_total",
			Help: "Total memory allocated in bytes",
		},
	)

	// ControllerMemorySysBytes is a gauge for memory obtained from OS in bytes
	ControllerMemorySysBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_controller_resource_memory_sys_bytes",
			Help: "Memory obtained from OS in bytes",
		},
	)

	// ControllerGoroutines is a gauge for number of goroutines
	ControllerGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_controller_resource_goroutines",
			Help: "Number of goroutines",
		},
	)

	// ControllerGCCPUFraction is a gauge for fraction of CPU time used by GC
	ControllerGCCPUFraction = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_controller_resource_gc_cpu_fraction",
			Help: "Fraction of CPU time used by GC",
		},
	)

	// ControllerNumGC is a counter for number of GC cycles
	ControllerNumGC = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ops_controller_resource_gc_cycles_total",
			Help: "Number of GC cycles",
		},
	)

	// ============================================================================
	// 控制器指标（Task、TaskRun、Pipeline、PipelineRun 等）
	// ============================================================================

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
			Name: "ops_controller_taskrun_total",
			Help: "Total number of TaskRun operations",
		},
		[]string{"namespace", "status"},
	)

	// TaskRunDuration is a histogram for TaskRun execution duration
	TaskRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_controller_taskrun_duration_seconds",
			Help:    "Duration of TaskRun execution in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9 hours
		},
		[]string{"namespace", "task_name"},
	)

	// PipelineRunTotal is a counter for PipelineRun operations
	PipelineRunTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_pipelinerun_total",
			Help: "Total number of PipelineRun operations",
		},
		[]string{"namespace", "status"},
	)

	// PipelineRunDuration is a histogram for PipelineRun execution duration
	PipelineRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_controller_pipelinerun_duration_seconds",
			Help:    "Duration of PipelineRun execution in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9 hours
		},
		[]string{"namespace", "pipeline_name"},
	)

	// HostConnectionStatus is a gauge for host connection status
	HostConnectionStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_host_connection_status",
			Help: "Host connection status (1 = connected, 0 = disconnected)",
		},
		[]string{"namespace", "host_name"},
	)

	// ClusterHealthStatus is a gauge for cluster health status
	ClusterHealthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_cluster_health_status",
			Help: "Cluster health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"namespace", "cluster_name"},
	)

	// ============================================================================
	// EventHooks 指标
	// ============================================================================

	// EventHooksReconcileTotal is a counter for the total number of EventHooks reconcile operations
	EventHooksReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_eventhooks_reconcile_total",
			Help: "Total number of EventHooks reconcile operations",
		},
		[]string{"namespace", "result"},
	)

	// EventHooksReconcileDuration is a histogram for EventHooks reconcile operation duration
	EventHooksReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_eventhooks_reconcile_duration_seconds",
			Help:    "Duration of EventHooks reconcile operations in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"namespace"},
	)

	// EventHooksReconcileErrors is a counter for EventHooks reconcile errors
	EventHooksReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_eventhooks_reconcile_errors_total",
			Help: "Total number of EventHooks reconcile errors",
		},
		[]string{"namespace", "error_type"},
	)

	// EventHooksEventProcessedTotal is a counter for processed events
	EventHooksEventProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_eventhooks_event_processed_total",
			Help: "Total number of events processed by EventHooks",
		},
		[]string{"namespace", "eventhook_name", "status"},
	)

	// EventHooksEventProcessDuration is a histogram for event processing duration
	EventHooksEventProcessDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_eventhooks_event_process_duration_seconds",
			Help:    "Duration of event processing in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~2s
		},
		[]string{"namespace", "eventhook_name"},
	)

	// ============================================================================
	// Server 基础资源指标（CPU、内存、IO、协程等）
	// ============================================================================

	// ServerMemoryAllocBytes is a gauge for server memory allocated in bytes
	ServerMemoryAllocBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_server_resource_memory_alloc_bytes",
			Help: "Server memory allocated in bytes",
		},
	)

	// ServerMemoryTotalAllocBytes is a counter for server total memory allocated in bytes
	ServerMemoryTotalAllocBytes = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ops_server_resource_memory_total_alloc_bytes_total",
			Help: "Server total memory allocated in bytes",
		},
	)

	// ServerMemorySysBytes is a gauge for server memory obtained from OS in bytes
	ServerMemorySysBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_server_resource_memory_sys_bytes",
			Help: "Server memory obtained from OS in bytes",
		},
	)

	// ServerGoroutines is a gauge for server number of goroutines
	ServerGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_server_resource_goroutines",
			Help: "Server number of goroutines",
		},
	)

	// ServerGCCPUFraction is a gauge for server fraction of CPU time used by GC
	ServerGCCPUFraction = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ops_server_resource_gc_cpu_fraction",
			Help: "Server fraction of CPU time used by GC",
		},
	)

	// ServerNumGC is a counter for server number of GC cycles
	ServerNumGC = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ops_server_resource_gc_cycles_total",
			Help: "Server number of GC cycles",
		},
	)

	// ============================================================================
	// Server 服务吞吐指标（状态码、QPS 等经典指标）
	// ============================================================================

	// HTTPRequestsTotal is a counter for HTTP requests
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_server_throughput_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestDuration is a histogram for HTTP request duration
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_server_throughput_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~2s
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestSize is a histogram for HTTP request body size
	HTTPRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_server_throughput_http_request_size_bytes",
			Help:    "Size of HTTP request body in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to ~1GB
		},
		[]string{"method", "path"},
	)

	// HTTPResponseSize is a histogram for HTTP response body size
	HTTPResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_server_throughput_http_response_size_bytes",
			Help:    "Size of HTTP response body in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to ~1GB
		},
		[]string{"method", "path", "status_code"},
	)

	// APIRequestsTotal is a counter for API requests
	APIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_server_throughput_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint", "namespace", "status"},
	)

	// APIRequestDuration is a histogram for API request duration
	APIRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_server_throughput_api_request_duration_seconds",
			Help:    "Duration of API requests in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~2s
		},
		[]string{"endpoint", "namespace"},
	)

	// APIErrorsTotal is a counter for API errors
	APIErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_server_throughput_api_errors_total",
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
	// 基本的资源使用指标
	metrics.Registry.MustRegister(
		ControllerMemoryAllocBytes,
		ControllerMemoryTotalAllocBytes,
		ControllerMemorySysBytes,
		ControllerGoroutines,
		ControllerGCCPUFraction,
		ControllerNumGC,
	)

	// 控制器指标（Task、TaskRun、Pipeline、PipelineRun 等）
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

	// EventHooks 指标
	metrics.Registry.MustRegister(
		EventHooksReconcileTotal,
		EventHooksReconcileDuration,
		EventHooksReconcileErrors,
		EventHooksEventProcessedTotal,
		EventHooksEventProcessDuration,
	)

	// 启动资源使用指标的定期更新
	go updateResourceMetrics()
}

// InitServer initializes and registers server-specific metrics
// with the controller-runtime metrics registry
func InitServer() {
	// Server 基础资源指标
	metrics.Registry.MustRegister(
		ServerMemoryAllocBytes,
		ServerMemoryTotalAllocBytes,
		ServerMemorySysBytes,
		ServerGoroutines,
		ServerGCCPUFraction,
		ServerNumGC,
	)

	// Server 服务吞吐指标
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

	// 启动 server 资源使用指标的定期更新
	go updateServerResourceMetrics()
}

// updateResourceMetrics periodically updates resource usage metrics for controller
func updateResourceMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var lastNumGC uint32
	var lastPauseTotalNs uint64
	var lastTotalAlloc uint64

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Update memory metrics
		ControllerMemoryAllocBytes.Set(float64(m.Alloc))

		// Update total alloc (cumulative counter)
		if m.TotalAlloc > lastTotalAlloc {
			ControllerMemoryTotalAllocBytes.Add(float64(m.TotalAlloc - lastTotalAlloc))
			lastTotalAlloc = m.TotalAlloc
		}

		ControllerMemorySysBytes.Set(float64(m.Sys))

		// Update goroutine count
		ControllerGoroutines.Set(float64(runtime.NumGoroutine()))

		// Update GC metrics
		if m.NumGC > lastNumGC {
			ControllerNumGC.Add(float64(m.NumGC - lastNumGC))
			lastNumGC = m.NumGC
		}

		// Calculate GC CPU fraction
		if m.PauseTotalNs > lastPauseTotalNs {
			pauseDelta := float64(m.PauseTotalNs - lastPauseTotalNs)
			// Approximate CPU fraction: pause time / (10 seconds * 1e9 nanoseconds)
			// This is a rough estimate since we don't have exact CPU time
			cpuFraction := pauseDelta / (10.0 * 1e9)
			ControllerGCCPUFraction.Set(cpuFraction)
			lastPauseTotalNs = m.PauseTotalNs
		}
	}
}

// updateServerResourceMetrics periodically updates resource usage metrics for server
func updateServerResourceMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var lastNumGC uint32
	var lastPauseTotalNs uint64
	var lastTotalAlloc uint64

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Update memory metrics
		ServerMemoryAllocBytes.Set(float64(m.Alloc))

		// Update total alloc (cumulative counter)
		if m.TotalAlloc > lastTotalAlloc {
			ServerMemoryTotalAllocBytes.Add(float64(m.TotalAlloc - lastTotalAlloc))
			lastTotalAlloc = m.TotalAlloc
		}

		ServerMemorySysBytes.Set(float64(m.Sys))

		// Update goroutine count
		ServerGoroutines.Set(float64(runtime.NumGoroutine()))

		// Update GC metrics
		if m.NumGC > lastNumGC {
			ServerNumGC.Add(float64(m.NumGC - lastNumGC))
			lastNumGC = m.NumGC
		}

		// Calculate GC CPU fraction
		if m.PauseTotalNs > lastPauseTotalNs {
			pauseDelta := float64(m.PauseTotalNs - lastPauseTotalNs)
			// Approximate CPU fraction: pause time / (10 seconds * 1e9 nanoseconds)
			// This is a rough estimate since we don't have exact CPU time
			cpuFraction := pauseDelta / (10.0 * 1e9)
			ServerGCCPUFraction.Set(cpuFraction)
			lastPauseTotalNs = m.PauseTotalNs
		}
	}
}
