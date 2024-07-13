package v1

const APIVersion = "crd.chenshaowen.com/v1"

const (
	TaskKind        = "Task"
	TaskRunKind     = "TaskRun"
	PipelineKind    = "Pipeline"
	PipelineRunKind = "PipelineRun"
)

const (
	LabelCronKey           = "ops/cron"
	LabelCronTaskValue     = "task"
	LabelCronPipelineValue = "pipeline"
	LabelTaskRefKey     = "ops/taskref"
	LabelPipelineRefKey = "ops/pipelineref"
	DefaultMaxRunHistory   = 1
)

type Variables struct {
	Default  string   `json:"default,omitempty" yaml:"default,omitempty"`
	Display  string   `json:"display,omitempty" yaml:"display,omitempty"`
	Value    string   `json:"value,omitempty" yaml:"value,omitempty"`
	Desc     string   `json:"desc,omitempty" yaml:"desc,omitempty"`
	Regex    string   `json:"regex,omitempty" yaml:"regex,omitempty"`
	Required bool     `json:"required,omitempty" yaml:"required,omitempty"`
	Enum     []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Examples []string `json:"examples,omitempty" yaml:"examples,omitempty"`
}
