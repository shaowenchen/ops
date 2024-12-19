package event

import (
	"context"
	"errors"
	"strings"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	"github.com/nats-io/nats.go"
)

type EventBus struct {
	Server        string
	Subject       string
	Client        *ceclient.Client
	Protocal      *cenats.Protocol
	Cancel        context.CancelFunc
	ConsumerFuncs []func(ctx context.Context, event cloudevents.Event)
}

func (bus *EventBus) GetClient() (*ceclient.Client, error) {
	// MutexClient.Lock()
	// defer MutexClient.Unlock()
	// key := bus.Server + bus.Subject
	// if client, ok := CacheClient[key]; ok {
	// 	return client, nil
	// }
	if bus.Client != nil {
		return bus.Client, nil
	}
	natsOptions := []nats.Option{}
	p, err := cenats.NewProtocol(endpoint, bus.Subject, bus.Subject, natsOptions)
	if err != nil {
		return nil, err
	}
	bus.Protocal = p
	c, err := ceclient.New(p)
	if err != nil {
		return nil, err
	}
	// CacheClient[key] = c
	return &c, nil
}

func (bus *EventBus) Close(ctx context.Context) {
	if bus.Protocal != nil {
		bus.Protocal.Close(ctx)
		bus.Protocal = nil
	}
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
	client, err := bus.GetClient()
	if err != nil {
		return err
	}
	result := (*client).Send(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.New("failed to publish")
	}
	return nil
}

func (bus *EventBus) Subscribe(ctx context.Context) error {
	if bus.Cancel != nil {
		bus.Cancel()
	}
	client, err := bus.GetClient()
	if err != nil {
		return err
	}
	println("len(bus.ConsumerFuncs):", len(bus.ConsumerFuncs))
	combineFn := func(ctx context.Context, event cloudevents.Event) {
		var fns = bus.ConsumerFuncs
		println("len(fns):", len(fns))
		println("subject:", event.Subject())
		for _, fn := range fns {
			fn(ctx, event)
		}
	}
	newCtx, cancel := context.WithCancel(ctx)
	bus.Cancel = cancel
	return (*client).StartReceiver(newCtx, combineFn)
}
