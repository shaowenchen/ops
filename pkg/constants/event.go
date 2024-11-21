package constants

import (
	"fmt"
)

const Source = "https://github.com/shaowenchen/ops"

const DefaultEventBusServer = "http://nats-headless.ops-system.svc:4222"

const SubjectController = KindOps + ".%s." + KindController
const SubjectHost = KindOps + ".%s." + KindHost
const SubjectCluster = KindOps + ".%s." + KindCluster
const SubjectTask = KindOps + ".%s." + KindTask
const SubjectTaskRun = KindOps + ".%s." + KindTaskRun
const SubjectPipeline = KindOps + ".%s." + KindPipeline
const SubjectPipelineRun = KindOps + ".%s." + KindPipelineRun
const SubjectCheck = KindOps + ".%s." + EventCheck

const SubjectWebhook = KindOps + "." + EventWebhook

func GetClusterSubject(cluster, format string) string {
	if cluster == "" {
		cluster = "default"
	}
	return fmt.Sprintf(format, cluster)
}

const (
	EventCheck   = "Check"
	EventWebhook = "Webhook"
	EventUnknown = "Unknown"
)

const ActionClearDisk = "clean disk"
const ActionGetDataSetStatus = "get dataset status"
const ActionGetNodeStatus = "get node status"
