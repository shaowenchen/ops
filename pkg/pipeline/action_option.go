package pipeline

import (
	"github.com/shaowenchen/opscli/pkg/host"
)
type PipelineOption struct{
	Debug bool
	FilePath string
	host.HostOption
}