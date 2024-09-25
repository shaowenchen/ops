package event

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	opsv1 "github.com/shaowenchen/ops/api/v1"
)

type GlobalEventBusClients struct {
	Mutex   sync.RWMutex
	Clients map[string]*cloudevents.Client
}

func (globalClient *GlobalEventBusClients) GetClient(server string, subject string) (*cloudevents.Client, error) {
	// get from cache
	key := fmt.Sprintf("client-%s-%s", server, subject)
	globalClient.Mutex.RLock()
	clientP, ok := globalClient.Clients[key]
	globalClient.Mutex.RUnlock()
	if !ok {
		// build cache
		p, err := cenats.NewSender(server, subject, cenats.NatsOptions())
		if err != nil {
			return nil, err
		}
		client, err := cloudevents.NewClient(p)
		if err != nil {
			return nil, err
		}
		// update cache
		globalClient.Mutex.Lock()
		defer globalClient.Mutex.Unlock()
		if globalClient.Clients == nil {
			globalClient.Clients = make(map[string]*cloudevents.Client)
		}
		globalClient.Clients[key] = &client
		clientP = &client
	}
	return clientP, nil
}

var CurrentEventBusClient = &GlobalEventBusClients{}

type EventBus struct {
	EventServer string
	Subject     string
}

const DefaultEventServer = "http://nats-headless:4222"
const SubjectOps = "ops.ops"
const SubjectHost = "ops.host"
const SubjectCluster = "ops.cluster"
const SubjectTaskRun = "ops.taskrun"
const SubjectPipelineRun = "ops.pipelinerun"

func NewEventBus() *EventBus {
	return &EventBus{}
}

func (bus *EventBus) BuildWithSubject(subject string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Subject = subject
	if os.Getenv("EVENT_SERVER") != "" {
		bus.EventServer = os.Getenv("EVENT_SERVER")
	} else {
		bus.EventServer = DefaultEventServer
	}
	return bus
}

func (bus *EventBus) publishCloudEvent(ctx context.Context, event cloudevents.Event) error {
	// get client
	client, err := CurrentEventBusClient.GetClient(bus.EventServer, bus.Subject)
	if err != nil {
		return err
	}
	result := (*client).Send(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.New("failed to publish")
	}
	return nil
}

func (bus *EventBus) Publish(ctx context.Context, data interface{}) error {
	event, err := builderEvent(data)
	if err != nil {
		return err
	}
	return bus.publishCloudEvent(ctx, event)
}

func (bus *EventBus) getCloudEvent(ctx context.Context, eventType string) (*cloudevents.Event, error) {
	var receivedEvent *cloudevents.Event
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// get client
	client, err := CurrentEventBusClient.GetClient(bus.EventServer, bus.Subject)
	if err != nil {
		return nil, err
	}
	err = (*client).StartReceiver(ctx, func(event cloudevents.Event) cloudevents.Result {
		if event.Type() == eventType {
			receivedEvent = &event
			cancel()
			return cloudevents.ResultACK
		}
		return cloudevents.ResultNACK
	})
	return receivedEvent, err
}

func (bus *EventBus) GetHost(ctx context.Context) (*EventHost, error) {
	event, err := bus.getCloudEvent(ctx, opsv1.HostKind)
	if err != nil {
		return nil, err
	}
	var data EventHost
	err = event.DataAs(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (bus *EventBus) GetCluster(ctx context.Context) (*EventCluster, error) {
	event, err := bus.getCloudEvent(ctx, opsv1.ClusterKind)
	if err != nil {
		return nil, err
	}
	var data EventCluster
	err = event.DataAs(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (bus *EventBus) GetTaskRun(ctx context.Context) (*EventTaskRun, error) {
	event, err := bus.getCloudEvent(ctx, opsv1.TaskRunKind)
	if err != nil {
		return nil, err
	}
	var data EventTaskRun
	err = event.DataAs(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (bus *EventBus) GetPipelineRun(ctx context.Context) (*EventPipelineRun, error) {
	event, err := bus.getCloudEvent(ctx, opsv1.PipelineRunKind)
	if err != nil {
		return nil, err
	}
	var data EventPipelineRun
	err = event.DataAs(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
