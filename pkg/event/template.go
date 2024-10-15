package event

import (
	"fmt"
	"strings"
	"time"
)

func (e EventHost) GetDiskUsageAlertMessage(t time.Time) string {
	formattedTime := t.Local().Format("2006-01-02 15:04:05")

	return fmt.Sprintf(
		`kind: disk usage alert
cluster: %s
host: %s
value: %s
action: clean disk
time: %s
`,
		e.Cluster,
		e.Status.Hostname,
		e.Status.DiskUsagePercent,
		formattedTime,
	)
}

func (e EventCheck) GetAlertMessage(t time.Time) string {
	var result strings.Builder
	appendField := func(label, value string) {
		if value != "" {
			result.WriteString(fmt.Sprintf("%s: %s\n", label, value))
		}
	}
	appendField("cluster", e.Cluster)
	appendField("host", e.Host)
	appendField("kind", e.Kind)
	appendField("threshold", e.Threshold)
	appendField("operator", e.Operator)
	appendField("value", e.Value)
	appendField("status", e.Status)
	appendField("message", e.Message)
	result.WriteString(fmt.Sprintf("time: %s\n", t.Local().Format("2006-01-02 15:04:05")))
	return result.String()
}
