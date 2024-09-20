package event

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type EventPipelineRun struct {
	Ref       string `json:"ref"`
	Desc      string `json:"desc"`
	Variables string `json:"variables"`
}

type EventTaskRun struct {
	Ref       string `json:"ref"`
	Desc      string `json:"desc"`
	Variables string `json:"variables"`
}

type EventInspection struct {
	TypeRef        string `json:"typeRef"`
	NameRef        string `json:"nameRef"`
	NodeName       string `json:"nodeName"`
	Variables      string `json:"variables"`
	ThresholdValue string `json:"thresholdValue"`
	Comparator     string `json:"comparator"`
	CurrentValue   string `json:"currentValue"`
	Status         string `json:"status"`
	Priority       string `json:"priority"`
}

func BuilderEvent(data interface{}) (cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetSource("https://www.chenshaowen.com/ops/")

	var eventType string
	switch v := data.(type) {
	case EventInspection:
		eventType = "ops.inspection"
	case *EventInspection:
		eventType = "ops.inspection"
	default:
		eventType = "ops.unknown"
		return e, fmt.Errorf("unsupported data type: %T", v)
	}
	e.SetType(eventType)
	err := e.SetData("application/json", data)
	return e, err
}
