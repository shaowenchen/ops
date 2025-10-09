package event

import (
	"fmt"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

func AppendField(result *strings.Builder, label, value string) {
	if label == "" || value == "" {
		return
	}
	result.WriteString(fmt.Sprintf("%s: %s \n", label, value))
}

func AppendCluster(result *strings.Builder, ce cloudevents.Event, cluster string) {
	if cluster == "" {
		clusterInterface, _ := ce.Context.GetExtension(opsconstants.ClusterLower)
		cluster, _ = clusterInterface.(string)
	}
	AppendField(result, "cluster", cluster)
}
