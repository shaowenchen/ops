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

func LogicExpression(exp string, ifEmptyDefault bool) (result bool, err error) {
	exp = strings.TrimSpace(exp)
	// default
	if len(exp) == 0 {
		return ifEmptyDefault, nil
	}
	// logic bool
	logicResult, err := utils.Logic(exp)
	if err == nil {
		return logicResult, nil
	}
	// expression
	if strings.Contains(exp, "==") {
		expPair := strings.Split(exp, "==")
		if len(expPair) == 2 {
			return strings.ToLower(utils.RemoveStartEndMark(expPair[0])) == strings.ToLower(utils.RemoveStartEndMark(expPair[1])), nil
		}
	} else if strings.Contains(exp, "!=") {
		expPair := strings.Split(exp, "!=")
		if len(expPair) == 2 {
			return strings.ToLower(utils.RemoveStartEndMark(expPair[0])) != strings.ToLower(utils.RemoveStartEndMark(expPair[1])), nil
		}
	}

	return
}
