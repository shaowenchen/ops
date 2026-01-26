package eventhook

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	Xiezuo        = "xiezuo"
	Webhook       = "webhook"
	Event         = "event"
	Elasticsearch = "elasticsearch"
)

// PostInterface defines the interface for notification post handlers
type PostInterface interface {
	Post(event *cloudevents.Event, url string, options map[string]string, data string, addtional string) error
}

// NotificationMap registers all notification types
var NotificationMap = map[string]PostInterface{
	Xiezuo:        &XiezuoPost{},
	Webhook:       &WebhookPost{},
	Event:         &EventPost{},
	Elasticsearch: &ESPost{},
}
