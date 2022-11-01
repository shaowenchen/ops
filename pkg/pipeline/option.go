package pipeline

import (
	"github.com/shaowenchen/opscli/pkg/host"
)

type PipelineOption struct {
	Debug     bool
	FilePath  string
	Variables map[string]string
	host.HostOption
}
