package event

import (
	"context"
	"errors"
	"strings"
	"sync"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type GlobalEventBusClients struct {
	Mutex   sync.RWMutex
	Clients map[string]*ProducerConsumerClient
}

type ProducerConsumerClient struct {
	Producer      *cloudevents.Client
	Consumer      *cloudevents.Client
	ConsumerFuncs []func(ctx context.Context, event cloudevents.Event)
}

func (globalClient *GlobalEventBusClients) GetClient(endpoint string, subject string) (*ProducerConsumerClient, error) {
	key := endpoint + subject

	globalClient.Mutex.RLock()
	clientP, ok := globalClient.Clients[key]
	globalClient.Mutex.RUnlock()

	if ok {
		return clientP, nil
	}

	globalClient.Mutex.Lock()
	defer globalClient.Mutex.Unlock()

	clientP, ok = globalClient.Clients[key]
	if ok {
		return clientP, nil
	}

	// build producer
	producerP, err := cenats.NewSender(endpoint, subject, cenats.NatsOptions())
	if err != nil {
		return nil, err
	}
	producerClient, err := cloudevents.NewClient(producerP)
	if err != nil {
		return nil, err
	}

	// build consumer
	consumerP, err := cenats.NewConsumer(endpoint, subject, cenats.NatsOptions())
	if err != nil {
		return nil, err
	}
	consumerClient, err := cloudevents.NewClient(consumerP)
	if err != nil {
		return nil, err
	}

	// update cache with the new client
	globalClient.Clients[key] = &ProducerConsumerClient{
		Producer: &producerClient,
		Consumer: &consumerClient,
	}

	return globalClient.Clients[key], nil
}

var CurrentEventBusClient = &GlobalEventBusClients{
	Clients: make(map[string]*ProducerConsumerClient),
}

type EventBus struct {
	Server  string
	Subject string
}

func (bus *EventBus) WithEndpoint(endpoint string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Server = endpoint
	return bus
}

func (bus *EventBus) WithSubject(subject string) *EventBus {
	if bus == nil {
		bus = &EventBus{}
	}
	bus.Subject = strings.ToLower(subject)
	return bus
}

func (bus *EventBus) Publish(ctx context.Context, data interface{}) error {
	if bus.Server == "" || bus.Subject == "" || data == nil {
		return nil
	}
	event, err := builderEvent(data)
	event.SetSubject(bus.Subject)
	if err != nil {
		return err
	}
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

func (bus *EventBus) Subscribe(ctx context.Context, fn func(ctx context.Context, event cloudevents.Event)) error {
	client, err := CurrentEventBusClient.GetClient(bus.Server, bus.Subject)
	if err != nil {
		return err
	}
	client.ConsumerFuncs = append(client.ConsumerFuncs, fn)
	receiverFn := func(event cloudevents.Event) {
		for _, consumerFunc := range client.ConsumerFuncs {
			consumerFunc(ctx, event)
		}
	}

	return (*client.Consumer).StartReceiver(ctx, receiverFn)
}
