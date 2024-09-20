package event

import (
	"context"
	"errors"
	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type EventBus struct {
	client     cloudevents.Client
	NatsServer string
	Subject    string
}

func NewEventBus(natsServer string, subject string) (*EventBus, error) {
	p, err := cenats.NewSender(natsServer, subject, cenats.NatsOptions())
	if err != nil {
		return nil, err
	}

	defer p.Close(context.Background())

	c, err := cloudevents.NewClient(p)
	if err != nil {
		return nil, err
	}
	return &EventBus{client: c, NatsServer: natsServer, Subject: subject}, nil
}

func (bus *EventBus) Publish(ctx context.Context, event cloudevents.Event) error {
	result := bus.client.Send(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.New("failed to publish")
	}
	return nil
}

func (bus *EventBus) Subscribe(ctx context.Context, handler func(ctx context.Context, event cloudevents.Event)) error {
	for {
		bus.client.StartReceiver(ctx, handler)
	}
}
