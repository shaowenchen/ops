package constants

import (
	"fmt"
)

const EventSetup = "setup"
const EventStatus = "status"

const Source = "https://github.com/shaowenchen/ops"

const SubjectController = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + KindController
const SubjectHost = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + KindHost
const SubjectCluster = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + KindCluster
const SubjectTask = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + KindTask
const SubjectTaskRun = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + KindTaskRun
const SubjectPipeline = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + KindPipeline
const SubjectPipelineRun = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + KindPipelineRun
const SubjectCheck = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + EventCheck
const SubjectWebhook = KindOps + "." + KindCluster + ".%s." + KindNamespace + ".%s." + EventWebhook

const (
	EventCheck   = "Check"
	EventWebhook = "Webhook"
	EventUnknown = "Unknown"
)

func GetClusterSubject(cluster, namespace, format string) string {
	return fmt.Sprintf(format, cluster, namespace)
}

const ActionClearDisk = "clean disk"
const ActionGetDataSetStatus = "get dataset status"
const ActionGetNodeStatus = "get node status"
