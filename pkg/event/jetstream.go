package event

import (
	"context"
	"encoding/json"
	"time"

	event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
)

type EventData struct {
	Subject string      `json:"subject"`
	Event   event.Event `json:"event"`
}

func QueryStartTime(client nats.JetStreamContext, subject string, startTime time.Time, maxLen uint, seconds uint) (data []EventData, err error) {
	sub, err := client.PullSubscribe(
		subject,
		"",
		nats.StartTime(startTime),
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	defer cancel()
	msgs, err := sub.Fetch(int(maxLen), nats.Context(ctx))
	if err != nil {
		return nil, err
	}
	for _, msg := range msgs {
		e := event.Event{}
		err := json.Unmarshal(msg.Data, &e)
		if err != nil {
			continue
		}
		data = append(data, EventData{
			Subject: msg.Subject,
			Event:   e,
		})
	}
	return data, nil
}
