package event

import (
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

type EventCluster struct {
	Cluster string              `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Server  string              `json:"server,omitempty" yaml:"server,omitempty" `
	Status  opsv1.ClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type EventTask struct {
	Cluster string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	opsv1.TaskSpec
	Status opsv1.TaskStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type EventTaskRun struct {
	Cluster   string            `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	TaskRef   string            `json:"taskRef"`
	Desc      string            `json:"desc"`
	Variables map[string]string `json:"variables"`
	opsv1.TaskRunStatus
}

type EventPipeline struct {
	Cluster string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	opsv1.PipelineSpec
	Status opsv1.PipelineStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type EventPipelineRun struct {
	Cluster     string            `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	PipelineRef string            `json:"pipelineRef"`
	Desc        string            `json:"desc"`
	Variables   map[string]string `json:"variables"`
	opsv1.PipelineRunStatus
}

type EventWebhook struct {
	Content    string `json:"content,omitempty" yaml:"content,omitempty"`
	WebhookUrl string `json:"webhookUrl,omitempty" yaml:"webhookUrl,omitempty"`
}

type EventInspection struct {
	Cluster     string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Host        string `json:"host,omitempty" yaml:"host,omitempty"`
	Kind        string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Threshold   string `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	Operator    string `json:"operator,omitempty" yaml:"operator,omitempty"`
	Value       string `json:"value,omitempty" yaml:"value,omitempty"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty"`
	Reason      string `json:"reason,omitempty" yaml:"reason,omitempty"`
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
		eventType = opsconstants.KindController
	case *EventHost, EventHost:
		eventType = opsconstants.KindHost
	case *EventCluster, EventCluster:
		eventType = opsconstants.KindCluster
	case *EventTask, EventTask:
		eventType = opsconstants.KindTask
	case *EventTaskRun, EventTaskRun:
		eventType = opsconstants.KindTaskRun
	case *EventPipeline, EventPipeline:
		eventType = opsconstants.KindPipeline
	case *EventPipelineRun, EventPipelineRun:
		eventType = opsconstants.KindPipelineRun
	case *EventWebhook, EventWebhook:
		eventType = opsconstants.EventWebhook
	case *EventInspection, EventInspection:
		eventType = opsconstants.EventInspection
	default:
		eventType = opsconstants.EventUnknown
	}
	e.SetType(eventType)
	err := e.SetData(cloudevents.ApplicationJSON, data)
	return e, err
}
