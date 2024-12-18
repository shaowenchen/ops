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
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

type EventHost struct {
	Address  string           `json:"address,omitempty" yaml:"address,omitempty"`
	Port     int              `json:"port,omitempty" yaml:"port,omitempty"`
	Username string           `json:"username,omitempty" yaml:"username,omitempty"`
	Status   opsv1.HostStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventHost) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluser(result, ce)
	AppendField(result, "address", e.Address)
	AppendField(result, "hostname", e.Status.Hostname)
	AppendField(result, "diskUsagePercent", e.Status.DiskUsagePercent)
	AppendField(result, "heartStatus", e.Status.HeartStatus)
	return result.String()
}

type EventCluster struct {
	Server string              `json:"server,omitempty" yaml:"server,omitempty" `
	Status opsv1.ClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventCluster) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluser(result, ce)
	AppendField(result, "server", e.Server)
	AppendField(result, "version", e.Status.Version)
	AppendField(result, "certNotAfterDays", fmt.Sprintf("%d", e.Status.CertNotAfterDays))
	AppendField(result, "heartStatus", e.Status.HeartStatus)
	return result.String()
}

type EventTask struct {
	opsv1.TaskSpec
	Status opsv1.TaskStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type EventTaskRun struct {
	TaskRef   string            `json:"taskRef"`
	Desc      string            `json:"desc"`
	Variables map[string]string `json:"variables"`
	opsv1.TaskRunStatus
}

type EventPipeline struct {
	opsv1.PipelineSpec
	Status opsv1.PipelineStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type EventPipelineRun struct {
	PipelineRef string            `json:"pipelineRef"`
	Desc        string            `json:"desc"`
	Variables   map[string]string `json:"variables"`
	opsv1.PipelineRunStatus
}

type EventWebhook struct {
	Content    string `json:"content,omitempty" yaml:"content,omitempty"`
	Source     string `json:"source,omitempty" yaml:"source,omitempty"`
	WebhookUrl string `json:"webhookUrl,omitempty" yaml:"webhookUrl,omitempty"`
}

func (e EventWebhook) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluser(result, ce)
	AppendField(result, "content", e.Content)
	AppendField(result, "source", e.Source)
	AppendField(result, "webhookUrl", e.WebhookUrl)
	AppendField(result, "time", ce.Time().Local().Format("2006-01-02 15:04:05"))
	return result.String()
}

type EventTaskRunReport struct {
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
	AppendCluser(result, ce)
	AppendField(result, `host: `, e.Host)
	AppendField(result, `kind: `, e.Kind)
	AppendField(result, `threshold: `, e.Threshold)
	AppendField(result, `operator: `, e.Operator)
	AppendField(result, `value: `, e.Value)
	AppendField(result, `status: `, e.Status)
	AppendField(result, `message: `, e.Message)
	AppendField(result, "time", ce.Time().Local().Format("2006-01-02 15:04:05"))
	return result.String()
}

type EventKube struct {
	Type              string    `json:"type,omitempty" yaml:"type,omitempty"`
	Reason            string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	CreationTimestamp time.Time `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	From              string    `json:"from,omitempty" yaml:"from,omitempty"`
	Message           string    `json:"message,omitempty" yaml:"message,omitempty"`
}

func (e EventKube) Readable(ce cloudevents.Event) string {
	var result = &strings.Builder{}
	AppendCluser(result, ce)
	subject := ce.Subject()
	// ops.clusters.xxx.namespaces.xxx.resources.xxx.events
	subjectSplits := strings.Split(subject, ".")
	if len(subjectSplits) == 8 {
		resources := subjectSplits[5]
		AppendField(result, "namespace", subjectSplits[4])
		AppendField(result, strings.TrimRight(resources, "s"), subjectSplits[6])
	}
	AppendField(result, "type", e.Type)
	AppendField(result, "reason", e.Reason)
	AppendField(result, "creationTimestamp", e.CreationTimestamp.Format("2006-01-02 15:04:05"))
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
		dataMap := data.(map[string]interface{})
		if _, ok1 := dataMap["kind"]; ok1 {
			if _, ok2 := dataMap["status"]; ok2 {
				eventType = opsconstants.TaskRunReport
			}
		}
	}
	e.SetType(eventType)
	err := e.SetData(cloudevents.ApplicationJSON, data)
	// add extension
	e.SetExtension(opsconstants.Cluster, cluster)
	return e, err
}

func GetCloudEventReadable(ce cloudevents.Event) string {
	if ce.Type() == opsconstants.Host {
		data := &EventHost{}
		ce.DataAs(data)
		return data.Readable(ce)
	} else if ce.Type() == opsconstants.Cluster {
		data := &EventCluster{}
		ce.DataAs(data)
		return data.Readable(ce)
	} else if ce.Type() == opsconstants.Webhook {
		data := &EventWebhook{}
		ce.DataAs(data)
		return data.Readable(ce)
	} else if ce.Type() == opsconstants.TaskRunReport {
		data := &EventTaskRunReport{}
		ce.DataAs(data)
		return data.Readable(ce)
	} else if ce.Type() == opsconstants.Kube {
		data := &EventKube{}
		ce.DataAs(data)
		return data.Readable(ce)
	} else {
		return string(ce.Data())
	}
}
