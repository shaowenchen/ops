package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type EventData struct {
	Subject string      `json:"subject"`
	Event   event.Event `json:"event"`
	Time    string      `json:"time,omitempty"` // Formatted time in local timezone
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
	// Ensure subscription is closed to prevent goroutine leaks
	defer func() {
		if sub != nil {
			_ = sub.Unsubscribe()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	defer cancel()

	// Fetch a larger batch to ensure we get all messages in the time range
	// Then we'll filter and sort to get the latest maxLen messages
	fetchBatch := int(maxLen * 10)
	if fetchBatch > 1000 {
		fetchBatch = 1000 // Cap at 1000 to avoid memory issues
	}
	msgs, err := sub.Fetch(fetchBatch, nats.Context(ctx))
	if err != nil {
		// If timeout, return empty (no messages in time range)
		if err == nats.ErrTimeout {
			return []EventData{}, nil
		}
		return nil, err
	}

	// Parse all messages and collect valid events
	allEvents := make([]EventData, 0, len(msgs))
	for _, msg := range msgs {
		e := event.Event{}
		err := json.Unmarshal(msg.Data, &e)
		if err != nil {
			continue
		}
		allEvents = append(allEvents, EventData{
			Subject: msg.Subject,
			Event:   e,
			Time:    e.Time().Local().Format("2006-01-02 15:04:05"),
		})
	}

	// Sort by event time (newest first)
	// Events are already in chronological order from Fetch, so reverse to get newest first
	for i := len(allEvents) - 1; i >= 0; i-- {
		data = append(data, allEvents[i])
	}

	// Take only the latest maxLen messages
	if len(data) > int(maxLen) {
		data = data[:maxLen]
	}

	return data, nil
}

func ListSubjects(url, streamName, search string, timeoutSeconds uint) (results []string, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	// Drain will gracefully close the connection and wait for pending operations
	defer func() {
		if nc != nil {
			_ = nc.Drain()
		}
	}()

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	stream, err := js.Stream(ctx, streamName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stream: %w", err)
	}

	info, err := stream.Info(
		ctx,
		jetstream.WithSubjectFilter(">"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stream info: %w", err)
	}

	for subject := range info.State.Subjects {
		if search == "" || strings.Contains(subject, search) {
			results = append(results, subject)
		}
	}

	return results, nil
}
