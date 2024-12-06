package eventhook

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	Xiezuo = "xiezuo"
)

type PostInterface interface {
	Post(url string, options map[string]string, data string, addtional string) error
}

var NotificationMap = map[string]PostInterface{
	Xiezuo: &XiezuoPost{},
}

type XiezuoPost struct {
	URL string
}

func (xiezuo *XiezuoPost) Post(url string, options map[string]string, data string, addtional string) error {
	data = data + addtional
	woaMd := map[string]interface{}{
		"text": data,
	}
	waoMsg := map[string]interface{}{
		"msgtype":  "markdown",
		"markdown": woaMd,
	}
	waoMsgJson, _ := json.Marshal(waoMsg)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(waoMsgJson)))
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
