package constants

const APIVersion = "crd.chenshaowen.com/v1"
const (
	KindOps         = "Ops"
	KindController  = "Controller"
	KindHost        = "Host"
	KindCluster     = "Cluster"
	KindTask        = "Task"
	KindTaskRun     = "TaskRun"
	KindPipeline    = "Pipeline"
	KindPipelineRun = "PipelineRun"
)

const Host = "host"
const Cluster = "cluster"

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
	DefaultTTLSecondsAfterFinished = 60 * 60
	ClearCronTab                   = "*/30 * * * *"
)
