package task

import (
	"github.com/shaowenchen/ops/pkg/host"
)

type TaskOption struct {
	Debug     bool
	Sudo      bool
	FilePath  string
	Variables map[string]string
	host.HostOption
}
