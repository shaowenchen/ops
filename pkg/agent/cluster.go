package agent

import (
	"encoding/json"
	"fmt"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"strings"
	"sync"
	"time"
)

type LLMCluster struct {
	Desc      string `json:"desc"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type LLMClustersManager struct {
	endpoint   string
	token      string
	namespace  string
	clusters   []LLMCluster
	tickerOnce sync.Once
}

func NewLLMClustersManager(endpoint, token, namespace string) *LLMClustersManager {
	return &LLMClustersManager{
		endpoint:  endpoint,
		token:     token,
		namespace: namespace,
		clusters:  make([]LLMCluster, 0),
	}
}

func (m *LLMClustersManager) GetMarkdown() string {
	var b strings.Builder
	b.WriteString("| namespace | name | desc | \n")
	b.WriteString("|-|-|-|\n")
	for _, item := range m.clusters {
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", item.Namespace, item.Name, item.Desc))
	}
	return b.String()
}

func (m *LLMClustersManager) GetText() string {
	var b strings.Builder
	for _, item := range m.clusters {
		b.WriteString(fmt.Sprintf("name: %s desc: %s,", item.Name, item.Desc))
	}
	return b.String()
}

func (m *LLMClustersManager) GetList() (list []string) {
	for _, c := range m.clusters {
		list = append(list, c.Name)
	}
	return list
}

func (m *LLMClustersManager) Update() (cs []LLMCluster, err error) {
	uri := "/api/v1/namespaces/" + m.namespace + "/clusters?page_size=999"
	body, err := makeRequest(m.endpoint, m.token, uri, "GET", nil)
	if err != nil || len(string(body)) < 10 {
		return
	}
	type ServerResponseList struct {
		Data struct {
			List []opsv1.Cluster `json:"list"`
		} `json:"data"`
	}
	var resp ServerResponseList
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	for _, c := range resp.Data.List {
		cs = append(cs, m.BuilderLLMCluster(&c))
	}
	return
}

func (m *LLMClustersManager) BuilderLLMCluster(c *opsv1.Cluster) LLMCluster {
	return LLMCluster{
		Desc:      c.Spec.Desc,
		Namespace: c.ObjectMeta.Namespace,
		Name:      c.ObjectMeta.Name,
	}
}

func (m *LLMClustersManager) StartUpdateTimer(interval time.Duration, updateFunc func() ([]LLMCluster, error)) {

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m.tickerOnce.Do(func() {
				res, err := updateFunc()
				if err != nil {
					fmt.Printf("timer update clusters err: %v\n", err)
					return
				}
				m.clusters = res
			})
		}
	}()
	m.clusters, _ = updateFunc()
}
