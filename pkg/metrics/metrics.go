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
	// Controller metrics (Task, TaskRun, Pipeline, PipelineRun, etc.)
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

	// CRDResourceStatusChangeTotal is a counter for CRD resource status changes
	// Records status changes for all CRD resources (TaskRun, PipelineRun, Cluster, Host, etc.)
	CRDResourceStatusChangeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_crd_resource_status_change_total",
			Help: "Total number of CRD resource status changes",
		},
		[]string{"controller", "resource_type", "namespace", "resource_name", "from_status", "to_status"},
	)

	// ScheduledTaskStatusChangeTotal is a counter for scheduled task (TaskRun/PipelineRun with Crontab) status changes
	ScheduledTaskStatusChangeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_scheduled_task_status_change_total",
			Help: "Total number of scheduled task status changes (TaskRun/PipelineRun with Crontab)",
		},
		[]string{"resource_type", "namespace", "resource_name", "crontab", "from_status", "to_status"},
	)

	// ============================================================================
	// EventHooks metrics
	// ============================================================================

	// EventHooksReconcileTotal is a counter for the total number of EventHooks reconcile operations
	EventHooksReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_eventhooks_reconcile_total",
			Help: "Total number of EventHooks reconcile operations",
		},
		[]string{"namespace", "result"},
	)

	// EventHooksReconcileDuration is a histogram for EventHooks reconcile operation duration
	EventHooksReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_controller_eventhooks_reconcile_duration_seconds",
			Help:    "Duration of EventHooks reconcile operations in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"namespace"},
	)

	// EventHooksReconcileErrors is a counter for EventHooks reconcile errors
	EventHooksReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_eventhooks_reconcile_errors_total",
			Help: "Total number of EventHooks reconcile errors",
		},
		[]string{"namespace", "error_type"},
	)

	// EventHooksEventProcessedTotal is a counter for processed events
	EventHooksEventProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_eventhooks_event_processed_total",
			Help: "Total number of events processed by EventHooks",
		},
		[]string{"namespace", "eventhook_name", "status"},
	)

	// EventHooksEventProcessDuration is a histogram for event processing duration
	EventHooksEventProcessDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ops_controller_eventhooks_event_process_duration_seconds",
			Help:    "Duration of event processing in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~2s
		},
		[]string{"namespace", "eventhook_name"},
	)

	// ============================================================================
	// Server basic resource metrics (CPU, memory, IO, goroutines, etc.)
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
	// Server throughput metrics (status codes, QPS, etc.)
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
	// Controller metrics
	metrics.Registry.MustRegister(
		ControllerReconcileTotal,
		ControllerReconcileDuration,
		ControllerReconcileErrors,
		CRDResourceStatusChangeTotal,
		ScheduledTaskStatusChangeTotal,
	)

	// EventHooks metrics
	metrics.Registry.MustRegister(
		EventHooksReconcileTotal,
		EventHooksReconcileDuration,
		EventHooksReconcileErrors,
		EventHooksEventProcessedTotal,
		EventHooksEventProcessDuration,
	)
}

// InitServer initializes and registers server-specific metrics
// with the controller-runtime metrics registry
func InitServer() {
	// Server basic resource metrics
	metrics.Registry.MustRegister(
		ServerMemoryAllocBytes,
		ServerMemoryTotalAllocBytes,
		ServerMemorySysBytes,
		ServerGoroutines,
		ServerGCCPUFraction,
		ServerNumGC,
	)

	// Server throughput metrics
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

	// Start periodic update of server resource usage metrics
	go updateServerResourceMetrics()
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
