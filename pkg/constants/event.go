package constants

import (
	"fmt"
	"os"
)

const OpsStreamName = "ops"

const Setup = "setup"
const Status = "status"

const Source = "https://github.com/shaowenchen/ops"

const SubjectClusterPrefix = Ops + "." + Clusters + ".%s"
const SubjectPrefix = Ops + "." + Clusters + ".%s." + Namespaces + ".%s"
const SubjectController = SubjectPrefix + "." + Controllers
const SubjectHost = SubjectPrefix + "." + Hosts
const SubjectCluster = SubjectPrefix + "." + Clusters
const SubjectTask = SubjectPrefix + "." + Tasks
const SubjectTaskRun = SubjectPrefix + "." + TaskRuns
const SubjectPipeline = SubjectPrefix + "." + Pipelines
const SubjectPipelineRun = SubjectPrefix + "." + PipelineRuns
const SubjectWebhook = SubjectPrefix + "." + Webhooks
const SubjectDeployments = SubjectPrefix + "." + Deployments

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
