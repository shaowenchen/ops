package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"io"
)

const LLMTaskPrefix = "task-"
const LLMPipelinePrefix = "pipeline-"
const OpsDefaultNamespace = "ops-system"

func ClearUnavailableChar(input string) string {
	for _, c := range []string{"-", "_", " "} {
		input = strings.ReplaceAll(input, c, "")
	}
	return input
}

func GetTaskListClusters() LLMTask {
	return LLMTask{
		Desc:      "use task `list-clusters` to list clusters",
		Namespace: "ops-system",
		Name:      "list-clusters",
		NodeName:  "anymaster",
	}
}

func GetTaskListTasks() LLMTask {
	return LLMTask{
		Desc:      "use task `list-tasks` to list tasks",
		Namespace: "ops-system",
		Name:      "list-tasks",
		NodeName:  "anymaster",
	}
}

func makeRequest(endpoint, token, uri, method string, payload interface{}) ([]byte, error) {
	url := endpoint + uri

	client := &http.Client{Timeout: 600 * time.Second}

	var req *http.Request
	var err error

	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(payloadBytes)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
