package event

import (
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	"time"
)

type EventController struct {
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

func (e EventController) String() string {
	return `kind: ` + e.Kind
}

type EventHost struct {
	Address  string           `json:"address,omitempty" yaml:"address,omitempty"`
	Port     int              `json:"port,omitempty" yaml:"port,omitempty"`
	Username string           `json:"username,omitempty" yaml:"username,omitempty"`
	Status   opsv1.HostStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventHost) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

type EventCluster struct {
	Server string              `json:"server,omitempty" yaml:"server,omitempty" `
	Status opsv1.ClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventCluster) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

type EventTask struct {
	opsv1.TaskSpec
	Status opsv1.TaskStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventTask) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

type EventTaskRun struct {
	TaskRef   string            `json:"taskRef"`
	Desc      string            `json:"desc"`
	Variables map[string]string `json:"variables"`
	opsv1.TaskRunStatus
}

func (e EventTaskRun) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

type EventPipeline struct {
	opsv1.PipelineSpec
	Status opsv1.PipelineStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (e EventPipeline) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

type EventPipelineRun struct {
	PipelineRef string            `json:"pipelineRef"`
	Desc        string            `json:"desc"`
	Variables   map[string]string `json:"variables"`
	opsv1.PipelineRunStatus
}

func (e EventPipelineRun) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

type EventWebhook struct {
	Content    string `json:"content,omitempty" yaml:"content,omitempty"`
	Source     string `json:"source,omitempty" yaml:"source,omitempty"`
	WebhookUrl string `json:"webhookUrl,omitempty" yaml:"webhookUrl,omitempty"`
}

func (e EventWebhook) String() string {
	r, _ := json.Marshal(e)
	return string(r)
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

func (e EventTaskRunReport) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

type EventKube struct {
	Type              string    `json:"type,omitempty" yaml:"type,omitempty"`
	Reason            string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	CreationTimestamp time.Time `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	From              string    `json:"from,omitempty" yaml:"from,omitempty"`
	Message           string    `json:"message,omitempty" yaml:"message,omitempty"`
}

func (e EventKube) String() string {
	r, _ := json.Marshal(e)
	return string(r)
}

func (e EventTaskRunReport) IsAlert() bool {
	return e.Status == "alert"
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
	e.SetType(eventType)
	err := e.SetData(cloudevents.ApplicationJSON, data)
	// add extension
	e.SetExtension(opsconstants.Cluster, cluster)
	return e, err
}

func GetCloudEventString(ce cloudevents.Event) string {
	if ce.Type() == opsconstants.Controller {
		data := &EventController{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.Host {
		data := &EventHost{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.Cluster {
		data := &EventCluster{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.Task {
		data := &EventTask{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.TaskRun {
		data := &EventTaskRun{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.Pipeline {
		data := &EventPipeline{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.PipelineRun {
		data := &EventPipelineRun{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.Webhook {
		data := &EventWebhook{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.TaskRunReport {
		data := &EventTaskRunReport{}
		ce.DataAs(data)
		return data.String()
	} else if ce.Type() == opsconstants.Kube {
		data := &EventKube{}
		ce.DataAs(data)
		return data.String()
	} else {
		return string(ce.Data())
	}
}
