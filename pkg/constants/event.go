package constants

import (
	"fmt"
	"os"
)

const EventSetup = "setup"
const EventStatus = "status"

const Source = "https://github.com/shaowenchen/ops"

const SubjectController = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + Controllers
const SubjectHost = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + Hosts
const SubjectCluster = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + Clusters
const SubjectTask = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + Tasks
const SubjectTaskRun = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + TaskRuns
const SubjectPipeline = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + Pipelines
const SubjectPipelineRun = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + PipelineRuns
const SubjectWebhook = Ops + "." + Clusters + ".%s." + Namespaces + ".%s." + EventWebhook

const (
	EventTaskRunReport = "TaskRunReport"
	EventWebhook       = "Webhook"
	EventUnknown       = "Unknown"
)

func GetClusterSubject(cluster, namespace, format string) string {
	return fmt.Sprintf(format, cluster, namespace)
}

const ActionClearDisk = "clean disk"
const ActionGetDataSetStatus = "get dataset status"
const ActionGetNodeStatus = "get node status"

func GetCurrentNamespace() (string, error) {
	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	data, err := os.ReadFile(namespaceFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
