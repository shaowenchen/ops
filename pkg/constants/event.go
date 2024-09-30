package constants

import "strings"

const Source = "https://github.com/shaowenchen/ops"

const DefaultEventBusServer = "http://nats-headless:4222"

const SubjectController = KindOps + "." + KindController
const SubjectHost = KindOps + "." + KindHost
const SubjectCluster = KindOps + "." + KindCluster
const SubjectTask = KindOps + "." + KindTask
const SubjectTaskRun = KindOps + "." + KindTaskRun
const SubjectPipeline = KindOps + "." + KindPipeline
const SubjectPipelineRun = KindOps + "." + KindPipelineRun
const SubjectWebhook = KindOps + "." + EventWebhook
const SubjectCheck = KindOps + "." + EventCheck

const (
	EventCheck   = "Check"
	EventWebhook = "Webhook"
	EventUnknown = "Unknown"
)

func IsCheckEvent(event string) bool {
	return strings.ToLower(event) == strings.ToLower(EventCheck)
}

func IsWebhookEvent(event string) bool {
	return strings.ToLower(event) == strings.ToLower(EventWebhook)
}
