package event

import (
	"fmt"
	"time"
)

func (e EventHost) GetUsageDiskTemplate(t time.Time) string {
	return fmt.Sprintf("cluster: %s\nhost: %s\ndisk usage: %s\ntime: %s\n", e.Cluster, e.Status.Hostname, e.Status.DiskUsagePercent, t.Local())
}

func (e EventCheck) GetAlertTemplate(t time.Time) string {
	var result string
	if e.Cluster != "" {
		result += fmt.Sprintf("cluster: %s\n", e.Cluster)
	}
	if e.Host != "" {
		result += fmt.Sprintf("host: %s\n", e.Host)
	}
	if e.Kind != "" {
		result += fmt.Sprintf("kind: %s\n", e.Kind)
	}
	if e.Threshold != "" {
		result += fmt.Sprintf("threshold: %s\n", e.Threshold)
	}
	if e.Operator != "" {
		result += fmt.Sprintf("operator: %s\n", e.Operator)
	}
	if e.Value != "" {
		result += fmt.Sprintf("value: %s\n", e.Value)
	}
	if e.Status != "" {
		result += fmt.Sprintf("status: %s\n", e.Status)
	}
	if e.Reason != "" {
		result += fmt.Sprintf("reason: %s\n", e.Reason)
	}

	result += fmt.Sprintf("time: %s\n", t.Local())

	return result
}
