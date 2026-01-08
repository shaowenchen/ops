package eventhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	opsevent "github.com/shaowenchen/ops/pkg/event"
)

const (
	Xiezuo  = "xiezuo"
	Webhook = "webhook"
	Event   = "event"
)

type PostInterface interface {
	Post(event *cloudevents.Event, url string, options map[string]string, data string, addtional string) error
}

var NotificationMap = map[string]PostInterface{
	Xiezuo:  &XiezuoPost{},
	Webhook: &WebhookPost{},
	Event:   &EventPost{},
}

type XiezuoPost struct {
	URL string
}

type XiezuoBody struct {
	Msgtype  string `json:"msgtype"`
	Markdown struct {
		Text string `json:"text"`
	} `json:"markdown"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

func (xiezuo *XiezuoPost) Post(event *cloudevents.Event, url string, options map[string]string, data string, addtional string) error {
	// if data is XiezuoBody, just post it
	sendJson := "{}"
	var tryXiezuoBody XiezuoBody
	err := json.Unmarshal([]byte(data), &tryXiezuoBody)
	if err == nil && tryXiezuoBody.Msgtype != "" {
		sendJson = data
	} else {
		data = data + addtional
		msg := XiezuoBody{}
		msg.Msgtype = "text"
		msg.Text.Content = data
		sendJsonBytes, _ := json.Marshal(msg)
		sendJson = string(sendJsonBytes)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(sendJson)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

type WebhookPost struct {
	URL string
}

func (webhook *WebhookPost) Post(event *cloudevents.Event, url string, options map[string]string, data string, addtional string) error {
	data = data + addtional
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

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

// replaceWildcards replaces wildcard "*" in urlTemplate with corresponding values from sourceSubject
// Examples:
//
//	urlTemplate: "ops.clusters.*.namespaces.*.pods.*.alerts"
//	sourceSubject: "ops.clusters.cluster1.namespaces.ns1.pods.pod1.events"
//	result: "ops.clusters.cluster1.namespaces.ns1.pods.pod1.alerts"
//
//	urlTemplate: "ops.clusters.*.nodes.*.alerts"
//	sourceSubject: "ops.clusters.cluster1.nodes.node1.events"
//	result: "ops.clusters.cluster1.nodes.node1.alerts"
func replaceWildcards(urlTemplate, sourceSubject string) string {
	// Split both strings by "."
	templateParts := strings.Split(urlTemplate, ".")
	sourceParts := strings.Split(sourceSubject, ".")

	// Build result by matching template parts with source parts
	result := make([]string, len(templateParts))
	sourceIndex := 0

	for i, templatePart := range templateParts {
		if templatePart == "*" {
			// Replace wildcard with corresponding value from source
			if sourceIndex < len(sourceParts) {
				result[i] = sourceParts[sourceIndex]
				sourceIndex++
			} else {
				// If source is shorter, keep the wildcard (shouldn't happen in normal cases)
				result[i] = "*"
			}
		} else {
			// Keep the template part as-is (e.g., "ops", "clusters", "namespaces", "pods", "alerts")
			result[i] = templatePart
			// If this part matches the source at current index, advance source index
			// This handles cases where template has fixed parts that should match source
			if sourceIndex < len(sourceParts) {
				if templatePart == sourceParts[sourceIndex] {
					// Matches, advance both
					sourceIndex++
				}
				// If doesn't match, we still keep the template part and don't advance source
				// This allows template to have different final parts (e.g., "alerts" vs "events")
			}
		}
	}

	return strings.Join(result, ".")
}
