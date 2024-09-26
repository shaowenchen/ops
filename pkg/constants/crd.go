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
	KindDefault     = "Default"
)

const StatusSuccessed = "Successed"
const StatusFailed = "Failed"
const StatusRunning = "Running"
const StatusAborted = "Aborted"
const StatusDataInValid = "DataInValid"
const StatusEmpty = ""

const (
	LabelCronKey                   = "ops/cron"
	LabelCronTaskValue             = "task"
	LabelCronPipelineValue         = "pipeline"
	LabelTaskRefKey                = "ops/taskref"
	LabelPipelineRefKey            = "ops/pipelineref"
	DefaultTTLSecondsAfterFinished = 60 * 60
	ClearCronTab                   = "*/30 * * * *"
)
