package event

import (
	"context"
	"github.com/nats-io/nats.go"
	"time"
)

func QueryStartTime(client nats.JetStreamContext, subject string, startTime time.Time, maxLen uint) (data []string, err error) {
	sub, err := client.PullSubscribe(
		subject,
		"",
		nats.StartTime(startTime),
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	msgs, err := sub.Fetch(int(maxLen), nats.Context(ctx))
	if err != nil {
		return nil, err
	}
	for _, msg := range msgs {
		data = append(data, string(msg.Data))
	}
	return data, nil
}
