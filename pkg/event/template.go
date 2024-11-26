package event

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"strings"
)

func (e EventHost) GetDiskUsageAlertMessageWithAction(event cloudevents.Event, action string) string {

	return e.GetDiskUsageAlertMessage(event) + fmt.Sprintf("action: %s  \n", action)
}

func (e EventHost) GetDiskUsageAlertMessage(event cloudevents.Event) string {
	var result strings.Builder
	appendField := func(label, value string) {
		if value != "" {
			result.WriteString(fmt.Sprintf("%s: %s  \n", label, value))
		}
	}
	clusterInterface, _ := event.Context.GetExtension("cluster")
	cluster, _ := clusterInterface.(string)
	if cluster != "" {
		appendField("cluster", cluster)
	}
	appendField("kind", "alert-disk-usage")
	appendField("host", e.Status.Hostname)
	appendField("value", e.Status.DiskUsagePercent)
	appendField("action", "clean disk")
	result.WriteString(fmt.Sprintf("time: %s  \n", event.Time().Local().Format("2006-01-02 15:04:05")))
	return result.String()
}

func (e EventTaskRunReport) GetAlertMessageWithAction(event cloudevents.Event, action string) string {
	return e.GetAlertMessage(event) + fmt.Sprintf("action: %s  \n", action)
}

func (e EventTaskRunReport) GetAlertMessage(event cloudevents.Event) string {
	var result strings.Builder
	appendField := func(label, value string) {
		if value != "" {
			result.WriteString(fmt.Sprintf("%s: %s  \n", label, value))
		}
	}
	clusterInterface, _ := event.Context.GetExtension("cluster")
	cluster, _ := clusterInterface.(string)
	if cluster != "" {
		appendField("cluster", cluster)
	}
	appendField("host", e.Host)
	appendField("kind", e.Kind)
	appendField("threshold", e.Threshold)
	appendField("operator", e.Operator)
	appendField("value", e.Value)
	appendField("status", e.Status)
	appendField("message", e.Message)
	result.WriteString(fmt.Sprintf("time: %s  \n", event.Time().Local().Format("2006-01-02 15:04:05")))
	return result.String()
}
