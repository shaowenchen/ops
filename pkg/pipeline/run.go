package pipeline

import (
	"github.com/shaowenchen/opscli/pkg/host"
)

func getStepFunc(step Step) (func(Step, PipelineOption) error, error) {
	if len(step.Script) > 0 {
		return runStepScript, nil
	}
	return runStepCopy, nil
}

func runStepScript(step Step, option PipelineOption) (err error) {
	scriptOption := host.ScriptOption{
		Content:    step.Script,
		HostOption: option.HostOption,
	}
	return host.ActionScript(scriptOption)
}

func runStepCopy(step Step, option PipelineOption) (err error) {
	fileOption := host.FileOption{
		Direction:  step.Direction,
		LocalFile:  step.LocalFile,
		RemoteFile: step.RemoteFile,
		HostOption: option.HostOption,
	}

	return host.ActionFile(fileOption)
}
