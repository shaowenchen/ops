package eventhook

import (
	"bytes"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

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
