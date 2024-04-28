package agent

import (
	"encoding/json"
	"fmt"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"strings"
	"sync"
	"time"
)

type LLMHost struct {
	Name      string `json:"name"`
	Desc      string `json:"desc"`
	Hostname  string `json:"hostname"`
	Address   string `json:"address"`
	Namespace string `json:"namespace"`
}

type LLMHostsManager struct {
	endpoint  string
	token     string
	namespace string
	hosts     []LLMHost
	tickOnce  sync.Once
}

func NewLLMHostsManager(endpoint, token, namespace string) *LLMHostsManager {
	return &LLMHostsManager{
		endpoint:  endpoint,
		token:     token,
		namespace: namespace,
		hosts:     make([]LLMHost, 0),
	}
}

func (m *LLMHostsManager) GetMarkdown() string {
	var b strings.Builder
	b.WriteString("| namespace | name | desc | address | hostname |\n")
	b.WriteString("|-|-|-|-|-|\n")
	for _, item := range m.hosts {
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", item.Namespace, item.Name, item.Desc, item.Address, item.Hostname))
	}
	return b.String()

}

func (m *LLMHostsManager) Update() (hs []LLMHost, err error) {
	uri := "/api/v1/namespaces/" + m.namespace + "/hosts?page_size=999"
	body, err := makeRequest(m.endpoint, m.token, uri, "GET", nil)
	if err != nil || len(string(body)) < 10 {
		return
	}
	type ServerResponseList struct {
		Data struct {
			List []opsv1.Host `json:"list"`
		} `json:"data"`
	}
	var resp ServerResponseList
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	for _, h := range resp.Data.List {
		hs = append(hs, m.BuilderLLMHost(&h))
	}
	return
}

func (m *LLMHostsManager) BuilderLLMHost(h *opsv1.Host) LLMHost {
	return LLMHost{
		Desc:      h.Spec.Desc,
		Namespace: h.ObjectMeta.Namespace,
		Name:      h.ObjectMeta.Name,
		Address:   h.Spec.Address,
		Hostname:  h.Status.Hostname,
	}
}

func (m *LLMHostsManager) StartUpdateTimer(interval time.Duration, updateFunc func() ([]LLMHost, error)) {

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.tickOnce.Do(func() {
				res, err := updateFunc()
				if err != nil {
					fmt.Printf("timer update hosts err: %v\n", err)
					return
				}
				m.hosts = res
			})
		}
	}()
	m.hosts, _ = updateFunc()
}
