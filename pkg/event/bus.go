package event

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

type Event cloudevents.Client

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
	key := fmt.Sprintf("%s-%s", server, subject)
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
	Server  string
	Subject string
}

func (bus *EventBus) WithServer(server string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Server = server
	return bus
}

func (bus *EventBus) WithSubject(subject string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Subject = strings.ToLower(subject)
	if os.Getenv("EVENTBUS_SERVER") != "" {
		bus.Server = os.Getenv("EVENTBUS_SERVER")
	} else {
		bus.Server = opsconstants.DefaultEventBusServer
	}
	return bus
}

func (bus *EventBus) Publish(ctx context.Context, data interface{}) error {
	event, err := builderEvent(data)
	if err != nil {
		return err
	}
	// get client
	client, err := CurrentEventBusClient.GetClient(bus.Server, bus.Subject)
	if err != nil {
		return err
	}
	result := (*client.Producer).Send(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.New("failed to publish")
	}
	return nil
}

func (bus *EventBus) Subscribe(ctx context.Context, fn interface{}) error {
	client, err := CurrentEventBusClient.GetClient(bus.Server, bus.Subject)
	if err != nil {
		return err
	}
	return (*client.Consumer).StartReceiver(ctx, fn)
}
