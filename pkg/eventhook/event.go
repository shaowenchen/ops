package eventhook

import (
	"context"
	"os"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	opsevent "github.com/shaowenchen/ops/pkg/event"
)

type EventPost struct{}

func (eventPost *EventPost) Post(event *cloudevents.Event, url string, options map[string]string, data string, addtional string) error {
	// url is the target event subject (NATS subject)
	if url == "" {
		return nil
	}

	// Determine target subject
	targetSubject := url
	// Get source subject from options
	sourceSubject := ""
	if options != nil {
		sourceSubject = options["sourceSubject"]
	}

	// If url contains wildcard "*", replace with values from source subject
	if strings.Contains(url, "*") && sourceSubject != "" {
		targetSubject = replaceWildcards(url, sourceSubject)
	} else if !strings.HasPrefix(url, "ops.") {
		// If url doesn't start with "ops.", treat it as a suffix replacement
		if sourceSubject != "" {
			// Replace the last part of the source subject with the url suffix
			parts := strings.Split(sourceSubject, ".")
			if len(parts) > 0 {
				parts[len(parts)-1] = url
				targetSubject = strings.Join(parts, ".")
			} else {
				// Fallback: if source subject is invalid, use url as-is
				targetSubject = url
			}
		} else {
			// If no source subject, use url as-is
			targetSubject = url
		}
	}

	// Get endpoint
	endpoint := os.Getenv("EVENT_ENDPOINT")
	if endpoint == "" {
		return nil
	}

	// Create a new event copy to avoid modifying the original event
	// Update subject, ID, and time for the forwarded event
	newEvent := event.Clone()
	newEvent.SetSubject(strings.ToLower(targetSubject))
	newEvent.SetID(uuid.New().String())
	newEvent.SetTime(time.Now())

	// Use EventBus.W to directly write event to the new subject
	ctx := context.Background()
	return (&opsevent.EventBus{}).WithEndpoint(endpoint).W(ctx, targetSubject, newEvent)
}
