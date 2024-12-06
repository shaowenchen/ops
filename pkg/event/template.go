package event

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"strings"
)

func (e EventHost) GetDiskUsageAlertMessageWithAction(event cloudevents.Event, action string) string {

	return e.GetDiskUsageAlertMessage(event) + fmt.Sprintf("action: %s  \n", action)
}

func AppendField(result *strings.Builder, label, value string) {
	result.WriteString(fmt.Sprintf("%s: %s  \n", label, value))
}

func AppendCluser(result *strings.Builder, ce cloudevents.Event) {
	clusterInterface, _ := ce.Context.GetExtension("cluster")
	cluster, _ := clusterInterface.(string)
	if cluster != "" {
		AppendField(result, "cluster", cluster)
	}
}

func (e EventHost) GetDiskUsageAlertMessage(event cloudevents.Event) string {
	var result = &strings.Builder{}

	AppendCluser(result, event)
	AppendField(result, "kind", "alert-disk-usage")
	AppendField(result, "host", e.Status.Hostname)
	AppendField(result, "value", e.Status.DiskUsagePercent)
	AppendField(result, "action", "clean disk")
	AppendField(result, "time", event.Time().Local().Format("2006-01-02 15:04:05"))
	return result.String()
}

func (e EventTaskRunReport) GetAlertMessageWithAction(event cloudevents.Event, action string) string {
	return e.GetAlertMessage(event) + fmt.Sprintf("action: %s  \n", action)
}

func (e EventTaskRunReport) GetAlertMessage(event cloudevents.Event) string {
	var result = &strings.Builder{}

	AppendCluser(result, event)
	AppendField(result, "host", e.Host)
	AppendField(result, "kind", e.Kind)
	AppendField(result, "threshold", e.Threshold)
	AppendField(result, "operator", e.Operator)
	AppendField(result, "value", e.Value)
	AppendField(result, "status", e.Status)
	AppendField(result, "message", e.Message)
	AppendField(result, "time", event.Time().Local().Format("2006-01-02 15:04:05"))
	return result.String()
}
