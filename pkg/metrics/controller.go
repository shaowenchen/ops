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
	"time"
)

// RecordReconcile records a reconcile operation
func RecordReconcile(controller, namespace, result string, duration time.Duration) {
	ControllerReconcileTotal.WithLabelValues(controller, namespace, result).Inc()
	ControllerReconcileDuration.WithLabelValues(controller, namespace).Observe(duration.Seconds())
}

// RecordReconcileError records a reconcile error
func RecordReconcileError(controller, namespace, errorType string) {
	ControllerReconcileErrors.WithLabelValues(controller, namespace, errorType).Inc()
}

// RecordTaskRun records a TaskRun operation
func RecordTaskRun(namespace, status string) {
	TaskRunTotal.WithLabelValues(namespace, status).Inc()
}

// RecordTaskRunDuration records TaskRun execution duration
func RecordTaskRunDuration(namespace, taskName string, duration time.Duration) {
	TaskRunDuration.WithLabelValues(namespace, taskName).Observe(duration.Seconds())
}

// RecordPipelineRun records a PipelineRun operation
func RecordPipelineRun(namespace, status string) {
	PipelineRunTotal.WithLabelValues(namespace, status).Inc()
}

// RecordPipelineRunDuration records PipelineRun execution duration
func RecordPipelineRunDuration(namespace, pipelineName string, duration time.Duration) {
	PipelineRunDuration.WithLabelValues(namespace, pipelineName).Observe(duration.Seconds())
}

// SetHostConnectionStatus sets the host connection status
func SetHostConnectionStatus(namespace, hostName string, connected bool) {
	value := 0.0
	if connected {
		value = 1.0
	}
	HostConnectionStatus.WithLabelValues(namespace, hostName).Set(value)
}

// SetClusterHealthStatus sets the cluster health status
func SetClusterHealthStatus(namespace, clusterName string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	ClusterHealthStatus.WithLabelValues(namespace, clusterName).Set(value)
}
