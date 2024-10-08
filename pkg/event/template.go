package event

import "fmt"

func (e EventHost) GetUsageDiskTemplate() string {
	return fmt.Sprintf("cluster: %s\nhost: %s\ndisk usage: %s", e.Cluster, e.Status.Hostname, e.Status.DiskUsagePercent)
}

func (e EventCheck) GetAlertTemplate() string {
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

	return result
}
