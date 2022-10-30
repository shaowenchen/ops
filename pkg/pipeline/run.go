package pipeline

import (
	"errors"
	"github.com/shaowenchen/opscli/pkg/host"
	"reflect"
	"strings"

	"github.com/shaowenchen/opscli/pkg/utils"
)

var internalFuncMap = map[string]interface{}{
	"GetAvailableUrl":            utils.GetAvailableUrl,
	"ScriptInstallMetricsServer": utils.ScriptInstallMetricsServer,
	"ScriptInstallOpscli":        utils.ScriptInstallOpscli,
}

func CallMap(funcName string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(internalFuncMap[funcName])
	if len(params) != f.Type().NumIn() {
		err = errors.New("the num of params is error")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func CheckWhen(when string) (needRun bool) {
	when = strings.TrimSpace(when)
	if len(when) == 0 {
		return true
	}
	if strings.Contains(when, "==") {
		whenPair := strings.Split(when, "==")
		if len(whenPair) == 2 {
			return strings.ToLower(utils.RemoveStartEndMark(whenPair[0])) == strings.ToLower(utils.RemoveStartEndMark(whenPair[1]))
		}
	} else if strings.Contains(when, "!=") {
		whenPair := strings.Split(when, "!=")
		if len(whenPair) == 2 {
			return strings.ToLower(utils.RemoveStartEndMark(whenPair[0])) != strings.ToLower(utils.RemoveStartEndMark(whenPair[1]))
		}
	}

	return false
}

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
