package event

import (
	"context"
	"errors"
	"strings"
	"sync"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	"github.com/nats-io/nats.go"
)

var CacheClient = make(map[string]ceclient.Client)
var MutexClient = sync.RWMutex{}

func GetClient(endpoint string, subject string) (ceclient.Client, error) {
	MutexClient.Lock()
	defer MutexClient.Unlock()
	key := endpoint + subject
	if client, ok := CacheClient[key]; ok {
		return client, nil
	}
	natsOptions := []nats.Option{}
	p, err := cenats.NewProtocol(endpoint, subject, subject, natsOptions)
	if err != nil {
		return nil, err
	}

	c, err := ceclient.New(p)
	if err != nil {
		return nil, err
	}
	CacheClient[key] = c
	return c, nil
}

type EventBus struct {
	Server        string
	Subject       string
	ConsumerFuncs []func(ctx context.Context, event cloudevents.Event)
}

func (bus *EventBus) AddConsumerFunc(fn func(ctx context.Context, event cloudevents.Event)) {
	bus.ConsumerFuncs = append(bus.ConsumerFuncs, fn)
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
	client, err := GetClient(bus.Server, bus.Subject)
	if err != nil {
		return err
	}
	result := client.Send(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.New("failed to publish")
	}
	return nil
}

func (bus *EventBus) Subscribe(ctx context.Context) error {
	client, err := GetClient(bus.Server, bus.Subject)
	if err != nil {
		return err
	}
	combineFn := func(ctx context.Context, event cloudevents.Event) {
		for _, fn := range bus.ConsumerFuncs {
			fn(ctx, event)
		}
	}

	return client.StartReceiver(ctx, combineFn)
}
