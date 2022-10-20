package pipeline

import (
	"errors"
	"github.com/shaowenchen/opscli/pkg/utils"
	"reflect"
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
