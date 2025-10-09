package event

import (
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

type EventController struct {
	Cluster string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

type EventHost struct {
	Cluster  string           `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Address  string           `json:"address,omitempty" yaml:"address,omitempty"`
	Port     int              `json:"port,omitempty" yaml:"port,omitempty"`
	Username string           `json:"username,omitempty" yaml:"username,omitempty"`
	Status   opsv1.HostStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventHost) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluster(result, ce, e.Cluster)
	AppendField(result, "address", e.Address)
	AppendField(result, "hostname", e.Status.Hostname)
	AppendField(result, "diskUsagePercent", e.Status.DiskUsagePercent)
	AppendField(result, "heartStatus", e.Status.HeartStatus)
	return result.String()
}

type EventCluster struct {
	Cluster string              `json:"cluster,omitempty" yaml:"cluster,omitempty" `
	Server  string              `json:"server,omitempty" yaml:"server,omitempty" `
	Status  opsv1.ClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventCluster) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluster(result, ce, e.Cluster)
	AppendField(result, "server", e.Server)
	AppendField(result, "version", e.Status.Version)
	AppendField(result, "certNotAfterDays", fmt.Sprintf("%d", e.Status.CertNotAfterDays))
	AppendField(result, "heartStatus", e.Status.HeartStatus)
	return result.String()
}

type EventTask struct {
	opsv1.TaskSpec
	Cluster string           `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Status  opsv1.TaskStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type EventTaskRun struct {
	Cluster   string            `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	TaskRef   string            `json:"taskRef"`
	Desc      string            `json:"desc"`
	Variables map[string]string `json:"variables"`
	opsv1.TaskRunStatus
}

type EventPipeline struct {
	opsv1.PipelineSpec
	Cluster string               `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Status  opsv1.PipelineStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type EventPipelineRun struct {
	Cluster     string            `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	PipelineRef string            `json:"pipelineRef"`
	Desc        string            `json:"desc"`
	Variables   map[string]string `json:"variables"`
	opsv1.PipelineRunStatus
}

type EventWebhook struct {
	Cluster    string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Content    string `json:"content,omitempty" yaml:"content,omitempty"`
	Source     string `json:"source,omitempty" yaml:"source,omitempty"`
	WebhookUrl string `json:"webhookUrl,omitempty" yaml:"webhookUrl,omitempty"`
	Type       string `json:"type,omitempty" yaml:"type,omitempty"`
}

func (e EventWebhook) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluster(result, ce, e.Cluster)
	AppendField(result, "content", e.Content)
	AppendField(result, "source", e.Source)
	AppendField(result, "webhookUrl", e.WebhookUrl)
	AppendField(result, "time", ce.Time().Local().Format("2006-01-02 15:04:05"))
	return result.String()
}

type EventTaskRunReport struct {
	Cluster   string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Host      string `json:"host,omitempty" yaml:"host,omitempty"`
	Kind      string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Threshold string `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	Operator  string `json:"operator,omitempty" yaml:"operator,omitempty"`
	Value     string `json:"value,omitempty" yaml:"value,omitempty"`
	Status    string `json:"status,omitempty" yaml:"status,omitempty"`
	Message   string `json:"message,omitempty" yaml:"message,omitempty"`
}

func (e EventTaskRunReport) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluster(result, ce, e.Cluster)
	AppendField(result, `host`, e.Host)
	AppendField(result, `kind`, e.Kind)
	AppendField(result, `threshold`, e.Threshold)
	AppendField(result, `operator`, e.Operator)
	AppendField(result, `value`, e.Value)
	AppendField(result, `status`, e.Status)
	AppendField(result, `message`, e.Message)
	AppendField(result, "time", ce.Time().Local().Format("2006-01-02 15:04:05"))
	return result.String()
}

type EventKube struct {
	Cluster   string    `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Type      string    `json:"type,omitempty" yaml:"type,omitempty"`
	Reason    string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	EventTime time.Time `json:"eventTime,omitempty" yaml:"eventTime,omitempty"`
	From      string    `json:"from,omitempty" yaml:"from,omitempty"`
	Message   string    `json:"message,omitempty" yaml:"message,omitempty"`
}

func (e EventKube) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	subject := ce.Subject()
	// ops.clusters.xxx.namespaces.xxx.resources.xxx.event
	// ops.clusters.xxx.resources.xxx.event
	subject = strings.TrimPrefix(subject, "ops.")
	subject = strings.TrimSuffix(subject, ".event")
	parts := strings.Split(subject, ".")
	for i := 0; i < len(parts)-1; i += 2 {
		key := strings.TrimRight(parts[i], "s")
		value := parts[i+1]
		AppendField(result, key, value)
	}
	AppendField(result, "type", e.Type)
	AppendField(result, "reason", e.Reason)
	AppendField(result, "timestamp", e.EventTime.Local().Format("2006-01-02 15:04:05"))
	AppendField(result, "from", e.From)
	AppendField(result, "message", e.Message)
	return result.String()
}

func (e EventTaskRunReport) IsAlerting() bool {
	return e.Status == "alerting"
}

func builderEvent(data interface{}) (cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetSource(opsconstants.Source)
	e.SetSpecVersion(cloudevents.VersionV1)
	e.SetTime(time.Now())
	var eventType string
	switch data.(type) {
	case *EventController, EventController:
		eventType = opsconstants.Controller
	case *EventHost, EventHost:
		eventType = opsconstants.Host
	case *EventCluster, EventCluster:
		eventType = opsconstants.Cluster
	case *EventTask, EventTask:
		eventType = opsconstants.Task
	case *EventTaskRun, EventTaskRun:
		eventType = opsconstants.TaskRun
	case *EventPipeline, EventPipeline:
		eventType = opsconstants.Pipeline
	case *EventPipelineRun, EventPipelineRun:
		eventType = opsconstants.PipelineRun
	case *EventWebhook, EventWebhook:
		eventType = opsconstants.Webhook
	case *EventTaskRunReport, EventTaskRunReport:
		eventType = opsconstants.TaskRunReport
	case *EventKube, EventKube:
		eventType = opsconstants.Kube
	default:
		eventType = opsconstants.Default
	}
	if eventType == opsconstants.Default {
		dataMap, ok := data.(map[string]interface{})
		if ok {
			if _, ok1 := dataMap["type"]; ok1 {
				dataType, ok2 := dataMap["type"].(string)
				if ok2 {
					eventType = dataType
				}
			}
		}
	}
	e.SetType(eventType)
	err := e.SetData(cloudevents.ApplicationJSON, data)
	// add extension
	e.SetExtension(opsconstants.ClusterLower, cluster)
	return e, err
}

func GetCloudEventReadable(ce cloudevents.Event) string {
	ceType := ce.Type()
	switch ceType {
	case opsconstants.Host:
		data := &EventHost{}
		ce.DataAs(data)
		return data.Readable(ce)
	case opsconstants.Cluster:
		data := &EventCluster{}
		ce.DataAs(data)
		return data.Readable(ce)
	case opsconstants.Webhook:
		data := &EventWebhook{}
		ce.DataAs(data)
		return data.Readable(ce)
	case opsconstants.TaskRunReport:
		data := &EventTaskRunReport{}
		ce.DataAs(data)
		return data.Readable(ce)
	case opsconstants.Kube:
		data := &EventKube{}
		ce.DataAs(data)
		return data.Readable(ce)
	default:
		return string(ce.Data())
	}
}
