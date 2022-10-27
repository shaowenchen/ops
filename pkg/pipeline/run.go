package pipeline

import (
	"github.com/shaowenchen/opscli/pkg/host"
)

func getStepFunc(step Step) func(Step, PipelineOption) (string, bool) {
	if len(step.Script) > 0 {
		return runStepScript
	}
	return runStepCopy
}

func runStepScript(step Step, option PipelineOption) (result string, isSuccessed bool) {
	scriptOption := host.ScriptOption{
		Content:    step.Script,
		HostOption: option.HostOption,
	}
	stdout, exit, _ := host.ActionScript(scriptOption)
	return stdout, exit == 0
}

func runStepCopy(step Step, option PipelineOption) (result string, isSuccessed bool) {
	fileOption := host.FileOption{
		LocalFile:  step.LocalFile,
		RemoteFile: step.RemoteFile,
		HostOption: option.HostOption,
	}
	return "", host.ActionFile(fileOption) == nil
}
