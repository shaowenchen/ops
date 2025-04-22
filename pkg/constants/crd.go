package constants

const APIVersion = "crd.chenshaowen.com/v1"
const (
	Ops           = "Ops"
	Controller    = "Controller"
	Controllers   = "Controllers"
	HostLower     = "host"
	Host          = "Host"
	Hosts         = "Hosts"
	ClusterLower  = "cluster"
	Cluster       = "Cluster"
	Clusters      = "Clusters"
	Task          = "Task"
	Tasks         = "Tasks"
	TaskRun       = "TaskRun"
	TaskRuns      = "TaskRuns"
	Pipeline      = "Pipeline"
	Pipelines     = "Pipelines"
	PipelineRun   = "PipelineRun"
	PipelineRuns  = "PipelineRuns"
	Namespace     = "Namespace"
	Namespaces    = "Namespaces"
	Webhook       = "Webhook"
	Webhooks      = "Webhooks"
	Event         = "Event"
	Events        = "Events"
	EventHook     = "EventHook"
	EventHooks    = "EventHooks"
	TaskRunReport = "TaskRunReport"
	Default       = "Default"
	Deployments   = "Deployments"
	Deployment    = "Deployment"
	Kube          = "Kube"
)

const StatusSuccessed = "Successed"
const StatusFailed = "Failed"
const StatusRunning = "Running"
const StatusAborted = "Aborted"
const StatusDataInValid = "DataInValid"
const StatusDispatched = "Dispatched"
const StatusEmpty = ""

func IsFinishedStatus(status string) bool {
	return status == StatusSuccessed || status == StatusFailed || status == StatusAborted || status == StatusDataInValid
}

const (
	LabelCronKey                   = "ops/cron"
	LabelCronTaskValue             = "task"
	LabelCronPipelineValue         = "pipeline"
	LabelTaskRefKey                = "ops/taskref"
	LabelPipelineRefKey            = "ops/pipelineref"
	DefaultTTLSecondsAfterFinished = 60 * 30
	ClearCronTab                   = "*/30 * * * *"
)
