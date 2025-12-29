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
}

func QueryStartTime(client nats.JetStreamContext, subject string, startTime time.Time, maxLen uint, seconds uint) (data []EventData, err error) {
	// Use empty string for ephemeral consumer (auto-generated name)
	// This avoids consumer conflicts and is cleaned up automatically
	sub, err := client.PullSubscribe(
		subject,
		"",
		nats.StartTime(startTime),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull subscription: %w", err)
	}
	defer func() {
		// Clean up subscription
		if sub != nil {
			_ = sub.Unsubscribe()
			_ = sub.Drain()
		}
	}()

	// Use longer timeout to ensure we can fetch messages
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	defer cancel()

	msgs, err := sub.Fetch(int(maxLen), nats.Context(ctx))
	// Handle timeout and no messages as valid cases (not errors)
	if err != nil {
		// Check for timeout errors - these are expected when no messages match the criteria
		errStr := err.Error()
		if err == nats.ErrTimeout ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "no messages") ||
			strings.Contains(errStr, "context deadline exceeded") {
			// No messages found within timeout, return empty result (not an error)
			return []EventData{}, nil
		}
		// For other errors, return them
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	// Process messages in reverse order (newest first)
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		e := event.Event{}
		err := json.Unmarshal(msg.Data, &e)
		if err != nil {
			// Skip messages that can't be unmarshaled, but continue with others
			continue
		}
		data = append(data, EventData{
			Subject: msg.Subject,
			Event:   e,
		})
		// Acknowledge message to prevent redelivery
		_ = msg.Ack()
	}
	return data, nil
}

func ListSubjects(url, streamName, search string) (results []string, err error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
