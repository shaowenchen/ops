package pipeline

import (
	"github.com/shaowenchen/ops/pkg/host"
)

type PipelineOption struct {
	Debug     bool
	Sudo      bool
	FilePath  string
	Variables map[string]string
	host.HostOption
}
