package event

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	opsv1 "github.com/shaowenchen/ops/api/v1"
)

type EventOps struct {
	Cluster    string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Controller string `json:"controller,omitempty" yaml:"controller,omitempty"`
}

type EventHost struct {
	Cluster  string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	opsv1.HostStatus
}

type EventCluster struct {
	Cluster string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Server  string `json:"server,omitempty" yaml:"server,omitempty" `
	opsv1.ClusterStatus
}
type EventTaskRun struct {
	Cluster   string            `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Ref       string            `json:"ref"`
	Desc      string            `json:"desc"`
	Variables map[string]string `json:"variables"`
	opsv1.TaskRunStatus
}

type EventPipelineRun struct {
	Cluster   string            `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Ref       string            `json:"ref"`
	Desc      string            `json:"desc"`
	Variables map[string]string `json:"variables"`
	opsv1.PipelineRunStatus
}

func builderEvent(data interface{}) (cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetSource(opsv1.APIVersion)

	var eventType string
	switch v := data.(type) {
	case *EventOps, EventOps:
		eventType = opsv1.OpsKind
	case *EventHost, EventHost:
		eventType = opsv1.TaskKind
	case *EventCluster, EventCluster:
		eventType = opsv1.ClusterKind
	case *EventTaskRun, EventTaskRun:
		eventType = opsv1.TaskRunKind
	case *EventPipelineRun, EventPipelineRun:
		eventType = opsv1.PipelineRunKind
	default:
		return e, fmt.Errorf("unsupported data type: %T", v)
	}
	e.SetType(eventType)
	err := e.SetData(cloudevents.ApplicationJSON, data)
	return e, err
}
