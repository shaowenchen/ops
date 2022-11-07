package task

import (
	"errors"
	"reflect"
	"strings"

	"github.com/shaowenchen/ops/pkg/utils"
)

var internalFuncMap = map[string]interface{}{
	"GetAvailableUrl":     utils.GetAvailableUrl,
	"ScriptInstallOpscli": utils.ScriptInstallOpscli,
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
	if when == "0" || strings.ToLower(when) == "false" || strings.ToLower(when) == "!true" {
		return false
	}
	if len(when) == 0 || when == "1" || strings.ToLower(when) == "true" || strings.ToLower(when) == "!false" {
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
