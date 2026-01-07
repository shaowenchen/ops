package eventhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	opsevent "github.com/shaowenchen/ops/pkg/event"
)

const (
	Xiezuo  = "xiezuo"
	Webhook = "webhook"
	Event   = "event"
)

type PostInterface interface {
	Post(url string, options map[string]string, data string, addtional string) error
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

func (xiezuo *XiezuoPost) Post(url string, options map[string]string, data string, addtional string) error {
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

func (webhook *WebhookPost) Post(url string, options map[string]string, data string, addtional string) error {
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

func (eventPost *EventPost) Post(url string, options map[string]string, data string, addtional string) error {
	// url is the target event subject (NATS subject)
	if url == "" {
		return nil
	}

	// Determine target subject
	targetSubject := url
	// If url doesn't start with "ops.", treat it as a suffix replacement
	if !strings.HasPrefix(url, "ops.") {
		// Get source subject from options
		sourceSubject := ""
		if options != nil {
			sourceSubject = options["sourceSubject"]
		}
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

	// Parse data as JSON
	var eventData map[string]interface{}
	if data != "" {
		err := json.Unmarshal([]byte(data), &eventData)
		if err != nil {
			// If data is not valid JSON, wrap it in a message field
			eventData = map[string]interface{}{
				"message": data + addtional,
			}
		} else if addtional != "" {
			// Merge additional data if provided
			var additionalData map[string]interface{}
			if err := json.Unmarshal([]byte(addtional), &additionalData); err == nil {
				for k, v := range additionalData {
					eventData[k] = v
				}
			} else {
				// If additional is not JSON, add it as a field
				eventData["additional"] = addtional
			}
		}
	} else {
		// If data is empty, use additional as message
		eventData = map[string]interface{}{
			"message": addtional,
		}
	}

	// Create EventBus and publish event
	endpoint := os.Getenv("EVENT_ENDPOINT")
	if endpoint == "" {
		return nil
	}

	eventBus := (&opsevent.EventBus{}).WithEndpoint(endpoint).WithSubject(targetSubject)
	ctx := context.Background()
	return eventBus.Publish(ctx, eventData)
}
