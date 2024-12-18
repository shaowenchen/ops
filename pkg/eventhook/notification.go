package eventhook

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	Xiezuo  = "xiezuo"
	Webhook = "webhook"
)

type PostInterface interface {
	Post(url string, options map[string]string, data string, addtional string) error
}

var NotificationMap = map[string]PostInterface{
	Xiezuo:  &XiezuoPost{},
	Webhook: &WebhookPost{},
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
	data = data + addtional
	msg := XiezuoBody{}
	msg.Msgtype = "text"
	msg.Text.Content = data
	json, _ := json.Marshal(msg)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		println(err.Error())
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
