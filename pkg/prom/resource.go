package prom

import (
	"context"
	"fmt"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	PrometheusServerKey = "prometheus-server"
)

func PromQuery(serverUrl string, promQl string) (result prommodel.Value, err error) {
	client, err := promapi.NewClient(promapi.Config{
		Address: serverUrl,
	})
	if err != nil {
		return
	}
	v1api := promv1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, _, err = v1api.Query(ctx, promQl, time.Now())
	if err != nil {
		return
	}
	return
}

func AlertPromQuery(serverUrl string, promQl string) (status string, result string, err error) {
	value, err := PromQuery(serverUrl, promQl)
	if err != nil {
		return
	}
	switch value.Type() {
	case prommodel.ValScalar:
		println("alert result is ValScalar")
	case prommodel.ValVector:
		results := value.(prommodel.Vector)
		if len(results) > 0 {
			status = opsv1.StatusFiring
		}
		items := make([]string, 0)
		for _, item := range results {
			itemMetrics := item.Metric.String()
			itemValue := item.Value.String()
			if itemMetrics == "{}" {
				items = append(items, itemValue)
			} else {
				items = append(items, fmt.Sprintf("%s %s", itemMetrics, itemValue))
			}
		}
		result = strings.Join(items, `\n`)
		return
	case prommodel.ValMatrix:
		println("alert result is ValMatrix")
	case prommodel.ValString:
		println("alert result is ValString")
	}
	return
}

func GetPrometheusServerUrl(client *kubernetes.Clientset) (url string, err error) {
	node, err := utils.GetAnyMaster(client)
	if err != nil {
		return
	}
	ip := utils.GetNodeInternalIp(node)
	port, err := GetPrometheusServerNodePort(client, "")
	if err != nil {
		return
	}
	return fmt.Sprintf("http://%s:%d", ip, port), nil
}

func GetPrometheusServerNodePort(client *kubernetes.Clientset, name string) (nodePort int32, err error) {
	if len(name) == 0 {
		name = PrometheusServerKey
	}
	svcs, err := client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, svc := range svcs.Items {
		if !strings.Contains(svc.Name, name) {
			continue
		}
		for _, port := range svc.Spec.Ports {
			if port.Name == "http" && port.Port == 80 {
				return port.NodePort, nil
			}
		}
	}
	return
}
