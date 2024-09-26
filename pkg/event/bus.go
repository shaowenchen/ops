package event

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

type GlobalEventBusClients struct {
	Mutex   sync.RWMutex
	Clients map[string]*ProducerConsumerClient
}

type ProducerConsumerClient struct {
	Producer *cloudevents.Client
	Consumer *cloudevents.Client
}

func (globalClient *GlobalEventBusClients) GetClient(server string, subject string) (*ProducerConsumerClient, error) {
	// get from cache
	key := fmt.Sprintf("client-%s-%s", server, subject)
	globalClient.Mutex.RLock()
	clientP, ok := globalClient.Clients[key]
	globalClient.Mutex.RUnlock()
	if !ok {
		// build producer
		producerP, err := cenats.NewSender(server, subject, cenats.NatsOptions())
		if err != nil {
			return nil, err
		}
		producerClient, err := cloudevents.NewClient(producerP)
		if err != nil {
			return nil, err
		}
		// build consumer
		consumerP, err := cenats.NewConsumer(server, subject, cenats.NatsOptions())
		if err != nil {
			return nil, err
		}
		consumerClient, err := cloudevents.NewClient(consumerP)
		if err != nil {
			return nil, err
		}
		// update cache
		globalClient.Mutex.Lock()
		defer globalClient.Mutex.Unlock()
		if globalClient.Clients == nil {
			globalClient.Clients = make(map[string]*ProducerConsumerClient)
		}
		globalClient.Clients[key] = &ProducerConsumerClient{Producer: &producerClient, Consumer: &consumerClient}
		clientP = globalClient.Clients[key]
	}
	return clientP, nil
}

var CurrentEventBusClient = &GlobalEventBusClients{}

type EventBus struct {
	EventServer string
	Subject     string
}

func (bus *EventBus) WithServer(server string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.EventServer = server
	return bus
}

func (bus *EventBus) WithSubject(subject string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Subject = subject
	if os.Getenv("EVENT_SERVER") != "" {
		bus.EventServer = os.Getenv("EVENT_SERVER")
	} else {
		bus.EventServer = opsconstants.DefaultEventServer
	}
	return bus
}

func (bus *EventBus) publishCloudEvent(ctx context.Context, event cloudevents.Event) error {
	// get client
	client, err := CurrentEventBusClient.GetClient(bus.EventServer, bus.Subject)
	if err != nil {
		return err
	}
	result := (*client.Producer).Send(ctx, event)
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

func (bus *EventBus) getCloudEvent(ctx context.Context) (*cloudevents.Event, error) {
	var receivedEvent *cloudevents.Event
	// get client
	client, err := CurrentEventBusClient.GetClient(bus.EventServer, bus.Subject)
	if err != nil {
		return nil, err
	}
	err = (*client.Consumer).StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event) error {
		receivedEvent = &event
		println("received event: " + event.Type())
		return nil
	})
	return receivedEvent, err
}

func (bus *EventBus) GetController(ctx context.Context) (*EventController, error) {
	event, err := bus.getCloudEvent(ctx)
	if err != nil {
		return nil, err
	}
	var data EventController
	err = event.DataAs(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (bus *EventBus) GetHost(ctx context.Context) (*EventHost, error) {
	event, err := bus.getCloudEvent(ctx)
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
	event, err := bus.getCloudEvent(ctx)
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

func (bus *EventBus) GetTask(ctx context.Context) (*EventTask, error) {
	event, err := bus.getCloudEvent(ctx)
	if err != nil {
		return nil, err
	}
	var data EventTask
	err = event.DataAs(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (bus *EventBus) GetTaskRun(ctx context.Context) (*EventTaskRun, error) {
	event, err := bus.getCloudEvent(ctx)
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

func (bus *EventBus) GetPipeline(ctx context.Context) (*EventPipeline, error) {
	event, err := bus.getCloudEvent(ctx)
	if err != nil {
		return nil, err
	}
	var data EventPipeline
	err = event.DataAs(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (bus *EventBus) GetPipelineRun(ctx context.Context) (*EventPipelineRun, error) {
	event, err := bus.getCloudEvent(ctx)
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

func (bus *EventBus) Get(ctx context.Context, dataPointer interface{}) error {
	event, err := bus.getCloudEvent(ctx)
	if err != nil {
		return err
	}
	err = event.DataAs(dataPointer)
	if err != nil {
		return err
	}
	return nil
}
