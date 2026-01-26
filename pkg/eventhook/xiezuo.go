package eventhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

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
