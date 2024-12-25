package event

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"strings"
)

func AppendField(result *strings.Builder, label, value string) {
	if label == "" || value == "" {
		return
	}
	result.WriteString(fmt.Sprintf("%s: %s \n", label, value))
}

func AppendCluser(result *strings.Builder, ce cloudevents.Event) {
	clusterInterface, _ := ce.Context.GetExtension("cluster")
	cluster, _ := clusterInterface.(string)
	if cluster != "" {
		AppendField(result, "cluster", cluster)
	}
}
