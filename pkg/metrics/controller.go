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

// ============================================================================
// CRD resource status change metrics recording functions
// ============================================================================

// RecordCRDResourceStatusChange records a CRD resource status change
func RecordCRDResourceStatusChange(controller, resourceType, namespace, resourceName, fromStatus, toStatus string) {
	if fromStatus == "" {
		fromStatus = "Empty"
	}
	if toStatus == "" {
		toStatus = "Empty"
	}
	CRDResourceStatusChangeTotal.WithLabelValues(controller, resourceType, namespace, resourceName, fromStatus, toStatus).Inc()
}

// ============================================================================
// Scheduled task status change metrics recording functions
// ============================================================================

// RecordScheduledTaskStatusChange records a scheduled task (TaskRun/PipelineRun with Crontab) status change
func RecordScheduledTaskStatusChange(resourceType, namespace, resourceName, crontab, fromStatus, toStatus string) {
	if fromStatus == "" {
		fromStatus = "Empty"
	}
	if toStatus == "" {
		toStatus = "Empty"
	}
	if crontab == "" {
		crontab = "N/A"
	}
	ScheduledTaskStatusChangeTotal.WithLabelValues(resourceType, namespace, resourceName, crontab, fromStatus, toStatus).Inc()
}

// ============================================================================
// TaskRef and PipelineRef usage metrics recording functions
// ============================================================================

// RecordTaskRefUsage records TaskRef usage in TaskRun
func RecordTaskRefUsage(namespace, taskRef, status string) {
	if status == "" {
		status = "Empty"
	}
	TaskRefUsageTotal.WithLabelValues(namespace, taskRef, status).Inc()
}

// RecordPipelineRefUsage records PipelineRef usage in PipelineRun
func RecordPipelineRefUsage(namespace, pipelineRef, status string) {
	if status == "" {
		status = "Empty"
	}
	PipelineRefUsageTotal.WithLabelValues(namespace, pipelineRef, status).Inc()
}

// ============================================================================
// EventHooks metrics recording functions
// ============================================================================

// RecordEventHooksReconcile records an EventHooks reconcile operation
func RecordEventHooksReconcile(namespace, result string, duration time.Duration) {
	EventHooksReconcileTotal.WithLabelValues(namespace, result).Inc()
	EventHooksReconcileDuration.WithLabelValues(namespace).Observe(duration.Seconds())
}

// RecordEventHooksReconcileError records an EventHooks reconcile error
func RecordEventHooksReconcileError(namespace, errorType string) {
	EventHooksReconcileErrors.WithLabelValues(namespace, errorType).Inc()
}

// RecordEventHooksEventProcessed records a processed event
func RecordEventHooksEventProcessed(namespace, eventhookName, status string) {
	EventHooksEventProcessedTotal.WithLabelValues(namespace, eventhookName, status).Inc()
}

// RecordEventHooksEventProcessDuration records event processing duration
func RecordEventHooksEventProcessDuration(namespace, eventhookName string, duration time.Duration) {
	EventHooksEventProcessDuration.WithLabelValues(namespace, eventhookName).Observe(duration.Seconds())
}
