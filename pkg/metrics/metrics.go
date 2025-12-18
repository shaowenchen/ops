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
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// PodName is the name of the current pod, used as a label in metrics
var PodName string

func init() {
	PodName = os.Getenv("HOSTNAME")
	if PodName == "" {
		PodName = "unknown"
	}
}

var (
	// ============================================================================
	// Resource Info metrics - expose all non-time fields for each resource
	// ============================================================================

	// TaskInfo records Task resource info
	TaskInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_task_info",
			Help: "Task resource info",
		},
		[]string{"pod", "namespace", "name", "desc", "host", "runtime_image"},
	)

	// PipelineInfo records Pipeline resource info
	PipelineInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_pipeline_info",
			Help: "Pipeline resource info",
		},
		[]string{"pod", "namespace", "name", "desc"},
	)

	// HostInfo records Host resource info (static fields only)
	HostInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_host_info",
			Help: "Host resource info",
		},
		[]string{"pod", "namespace", "name", "address", "hostname", "distribution", "arch"},
	)

	// ClusterInfo records Cluster resource info (static fields only)
	ClusterInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_cluster_info",
			Help: "Cluster resource info",
		},
		[]string{"pod", "namespace", "name", "server", "version"},
	)

	// EventHooksInfo records EventHooks resource info
	EventHooksInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_eventhooks_info",
			Help: "EventHooks resource info",
		},
		[]string{"pod", "namespace", "name", "type", "subject", "url"},
	)

	// ============================================================================
	// TaskRun/PipelineRun metrics - track running status, start/end time, duration
	// ============================================================================

	// TaskRunInfo records TaskRun resource info with all fields
	TaskRunInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_taskrun_info",
			Help: "TaskRun resource info",
		},
		[]string{"pod", "namespace", "name", "taskref", "crontab", "status"},
	)

	// TaskRunStartTime records TaskRun start time (unix timestamp)
	TaskRunStartTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_taskrun_start_time",
			Help: "TaskRun start time as unix timestamp",
		},
		[]string{"pod", "namespace", "name", "taskref"},
	)

	// TaskRunEndTime records TaskRun end time (unix timestamp)
	TaskRunEndTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_taskrun_end_time",
			Help: "TaskRun end time as unix timestamp",
		},
		[]string{"pod", "namespace", "name", "taskref"},
	)

	// TaskRunDurationSeconds records TaskRun duration in seconds
	TaskRunDurationSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_taskrun_duration_seconds",
			Help: "TaskRun duration in seconds",
		},
		[]string{"pod", "namespace", "name", "taskref", "status"},
	)

	// PipelineRunInfo records PipelineRun resource info with all fields
	PipelineRunInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_pipelinerun_info",
			Help: "PipelineRun resource info",
		},
		[]string{"pod", "namespace", "name", "pipelineref", "crontab", "status"},
	)

	// PipelineRunStartTime records PipelineRun start time (unix timestamp)
	PipelineRunStartTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_pipelinerun_start_time",
			Help: "PipelineRun start time as unix timestamp",
		},
		[]string{"pod", "namespace", "name", "pipelineref"},
	)

	// PipelineRunEndTime records PipelineRun end time (unix timestamp)
	PipelineRunEndTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_pipelinerun_end_time",
			Help: "PipelineRun end time as unix timestamp",
		},
		[]string{"pod", "namespace", "name", "pipelineref"},
	)

	// PipelineRunDurationSeconds records PipelineRun duration in seconds
	PipelineRunDurationSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_controller_pipelinerun_duration_seconds",
			Help: "PipelineRun duration in seconds",
		},
		[]string{"pod", "namespace", "name", "pipelineref", "status"},
	)

	// ============================================================================
	// Task/Pipeline run count metrics
	// ============================================================================

	// TaskRefRunTotal is a counter for TaskRef run count
	TaskRefRunTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_taskref_run_total",
			Help: "Total number of TaskRef runs",
		},
		[]string{"pod", "namespace", "taskref", "status"},
	)

	// PipelineRefRunTotal is a counter for PipelineRef run count
	PipelineRefRunTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_pipelineref_run_total",
			Help: "Total number of PipelineRef runs",
		},
		[]string{"pod", "namespace", "pipelineref", "status"},
	)

	// ============================================================================
	// EventHooks trigger metrics
	// ============================================================================

	// EventHooksTriggerTotal records EventHooks trigger count (only successful triggers)
	EventHooksTriggerTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_eventhooks_trigger_total",
			Help: "Total number of successful EventHooks triggers",
		},
		[]string{"pod", "namespace", "eventhook_name", "keyword", "event_id"},
	)

	// ============================================================================
	// Controller reconcile metrics
	// ============================================================================

	// ControllerReconcileTotal is a counter for the total number of reconcile operations
	ControllerReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_reconcile_total",
			Help: "Total number of reconcile operations",
		},
		[]string{"pod", "controller", "namespace", "result"},
	)

	// ControllerReconcileErrors is a counter for reconcile errors
	ControllerReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_controller_reconcile_errors_total",
			Help: "Total number of reconcile errors",
		},
		[]string{"pod", "controller", "namespace", "error_type"},
	)

	// ============================================================================
	// Server basic resource metrics
	// ============================================================================

	// ServerMemoryAllocBytes is a gauge for server memory allocated in bytes
	ServerMemoryAllocBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_server_resource_memory_alloc_bytes",
			Help: "Server memory allocated in bytes",
		},
		[]string{"pod"},
	)

	// ServerMemorySysBytes is a gauge for server memory obtained from OS in bytes
	ServerMemorySysBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_server_resource_memory_sys_bytes",
			Help: "Server memory obtained from OS in bytes",
		},
		[]string{"pod"},
	)

	// ServerGoroutines is a gauge for server number of goroutines
	ServerGoroutines = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_server_resource_goroutines",
			Help: "Server number of goroutines",
		},
		[]string{"pod"},
	)

	// ============================================================================
	// Server throughput metrics
	// ============================================================================

	// HTTPRequestsTotal is a counter for HTTP requests
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_server_throughput_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"pod", "method", "path", "status_code"},
	)

	// APIRequestsTotal is a counter for API requests
	APIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_server_throughput_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"pod", "endpoint", "namespace", "status"},
	)

	// APIErrorsTotal is a counter for API errors
	APIErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ops_server_throughput_api_errors_total",
			Help: "Total number of API errors",
		},
		[]string{"pod", "endpoint", "namespace", "error_type"},
	)

	// ServerInfo is a gauge for server information
	ServerInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_server_info",
			Help: "Server information",
		},
		[]string{"pod", "version", "build_date"},
	)

	// ServerUptime is a gauge for server uptime
	ServerUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ops_server_uptime_seconds",
			Help: "Server uptime in seconds",
		},
		[]string{"pod"},
	)
)

// InitController initializes and registers controller-specific metrics
func InitController() {
	// Resource info metrics
	metrics.Registry.MustRegister(
		TaskInfo,
		PipelineInfo,
		HostInfo,
		ClusterInfo,
		EventHooksInfo,
	)

	// TaskRun/PipelineRun metrics
	metrics.Registry.MustRegister(
		TaskRunInfo,
		TaskRunStartTime,
		TaskRunEndTime,
		TaskRunDurationSeconds,
		PipelineRunInfo,
		PipelineRunStartTime,
		PipelineRunEndTime,
		PipelineRunDurationSeconds,
	)

	// Run count metrics
	metrics.Registry.MustRegister(
		TaskRefRunTotal,
		PipelineRefRunTotal,
	)

	// EventHooks metrics
	metrics.Registry.MustRegister(
		EventHooksTriggerTotal,
	)

	// Controller reconcile metrics
	metrics.Registry.MustRegister(
		ControllerReconcileTotal,
		ControllerReconcileErrors,
	)
}

// InitServer initializes and registers server-specific metrics
func InitServer() {
	// Server basic resource metrics
	metrics.Registry.MustRegister(
		ServerMemoryAllocBytes,
		ServerMemorySysBytes,
		ServerGoroutines,
	)

	// Server throughput metrics
	metrics.Registry.MustRegister(
		HTTPRequestsTotal,
		APIRequestsTotal,
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

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		ServerMemoryAllocBytes.WithLabelValues(PodName).Set(float64(m.Alloc))
		ServerMemorySysBytes.WithLabelValues(PodName).Set(float64(m.Sys))
		ServerGoroutines.WithLabelValues(PodName).Set(float64(runtime.NumGoroutine()))
	}
}

// ============================================================================
// Resource info recording functions
// ============================================================================

// RecordTaskInfo records Task resource info
func RecordTaskInfo(namespace, name, desc, host, runtimeImage string) {
	TaskInfo.WithLabelValues(PodName, namespace, name, desc, host, runtimeImage).Set(1)
}

// RecordPipelineInfo records Pipeline resource info
func RecordPipelineInfo(namespace, name, desc string) {
	PipelineInfo.WithLabelValues(PodName, namespace, name, desc).Set(1)
}

// RecordHostInfo records Host resource info (static fields only)
func RecordHostInfo(namespace, name, address, hostname, distribution, arch string) {
	HostInfo.WithLabelValues(PodName, namespace, name, address, hostname, distribution, arch).Set(1)
}

// RecordClusterInfo records Cluster resource info (static fields only)
func RecordClusterInfo(namespace, name, server, version string) {
	ClusterInfo.WithLabelValues(PodName, namespace, name, server, version).Set(1)
}

// RecordEventHooksInfo records EventHooks resource info
func RecordEventHooksInfo(namespace, name, eventType, subject, url string) {
	EventHooksInfo.WithLabelValues(PodName, namespace, name, eventType, subject, url).Set(1)
}

// ============================================================================
// TaskRun/PipelineRun recording functions
// ============================================================================

// RecordTaskRunInfo records TaskRun resource info
func RecordTaskRunInfo(namespace, name, taskref, crontab, status string) {
	TaskRunInfo.WithLabelValues(PodName, namespace, name, taskref, crontab, status).Set(1)
}

// RecordTaskRunStart records TaskRun start time
func RecordTaskRunStart(namespace, name, taskref string, startTime float64) {
	TaskRunStartTime.WithLabelValues(PodName, namespace, name, taskref).Set(startTime)
}

// RecordTaskRunEnd records TaskRun end time and duration
func RecordTaskRunEnd(namespace, name, taskref, status string, endTime, durationSeconds float64) {
	TaskRunEndTime.WithLabelValues(PodName, namespace, name, taskref).Set(endTime)
	TaskRunDurationSeconds.WithLabelValues(PodName, namespace, name, taskref, status).Set(durationSeconds)
}

// RecordPipelineRunInfo records PipelineRun resource info
func RecordPipelineRunInfo(namespace, name, pipelineref, crontab, status string) {
	PipelineRunInfo.WithLabelValues(PodName, namespace, name, pipelineref, crontab, status).Set(1)
}

// RecordPipelineRunStart records PipelineRun start time
func RecordPipelineRunStart(namespace, name, pipelineref string, startTime float64) {
	PipelineRunStartTime.WithLabelValues(PodName, namespace, name, pipelineref).Set(startTime)
}

// RecordPipelineRunEnd records PipelineRun end time and duration
func RecordPipelineRunEnd(namespace, name, pipelineref, status string, endTime, durationSeconds float64) {
	PipelineRunEndTime.WithLabelValues(PodName, namespace, name, pipelineref).Set(endTime)
	PipelineRunDurationSeconds.WithLabelValues(PodName, namespace, name, pipelineref, status).Set(durationSeconds)
}

// ============================================================================
// Run count recording functions
// ============================================================================

// RecordTaskRefRun records TaskRef run count
func RecordTaskRefRun(namespace, taskref, status string) {
	TaskRefRunTotal.WithLabelValues(PodName, namespace, taskref, status).Inc()
}

// RecordPipelineRefRun records PipelineRef run count
func RecordPipelineRefRun(namespace, pipelineref, status string) {
	PipelineRefRunTotal.WithLabelValues(PodName, namespace, pipelineref, status).Inc()
}

// ============================================================================
// EventHooks recording functions
// ============================================================================

// RecordEventHooksTrigger records EventHooks trigger (only successful triggers)
func RecordEventHooksTrigger(namespace, eventhookName, keyword, eventID string) {
	EventHooksTriggerTotal.WithLabelValues(PodName, namespace, eventhookName, keyword, eventID).Inc()
}

// ============================================================================
// Controller reconcile recording functions
// ============================================================================

// RecordReconcile records a reconcile operation
func RecordReconcile(controller, namespace, result string) {
	ControllerReconcileTotal.WithLabelValues(PodName, controller, namespace, result).Inc()
}

// RecordReconcileError records a reconcile error
func RecordReconcileError(controller, namespace, errorType string) {
	ControllerReconcileErrors.WithLabelValues(PodName, controller, namespace, errorType).Inc()
}

// ============================================================================
// Server metrics recording functions
// ============================================================================

// RecordServerInfo records server info
func RecordServerInfo(version, buildDate string) {
	ServerInfo.WithLabelValues(PodName, version, buildDate).Set(1)
}

// RecordServerUptime records server uptime
func RecordServerUptime(uptimeSeconds float64) {
	ServerUptime.WithLabelValues(PodName).Set(uptimeSeconds)
}
