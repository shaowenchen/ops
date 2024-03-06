package utils

import (
	"strconv"
	"strings"
)

func LogicExpression(exp string, ifEmptyDefault bool) (result bool, err error) {
	exp = strings.TrimSpace(exp)
	// default
	if len(exp) == 0 {
		return ifEmptyDefault, nil
	}
	// logic bool
	logicResult, err := Logic(exp)
	if err == nil {
		return logicResult, nil
	}
	// expression
	if strings.Contains(exp, "==") {
		expPair := strings.Split(exp, "==")
		if len(expPair) == 2 {
			return strings.ToLower(RemoveStartEndMark(expPair[0])) == strings.ToLower(RemoveStartEndMark(expPair[1])), nil
		}
	} else if strings.Contains(exp, "!=") {
		expPair := strings.Split(exp, "!=")
		if len(expPair) == 2 {
			return strings.ToLower(RemoveStartEndMark(expPair[0])) != strings.ToLower(RemoveStartEndMark(expPair[1])), nil
		}
	} else if strings.Contains(exp, ">") {
		expPair := strings.Split(exp, ">")
		left, err := strconv.ParseFloat(RemoveStartEndMark(expPair[0]), 64)
		if err != nil {
			return false, err
		}
		right, err := strconv.ParseFloat(RemoveStartEndMark(expPair[1]), 64)
		if err != nil {
			return false, err
		}
		return left > right, nil
	} else if strings.Contains(exp, ">=") {
		expPair := strings.Split(exp, ">=")
		left, err := strconv.ParseFloat(RemoveStartEndMark(expPair[0]), 64)
		if err != nil {
			return false, err
		}
		right, err := strconv.ParseFloat(RemoveStartEndMark(expPair[1]), 64)
		if err != nil {
			return false, err
		}
		return left >= right, nil
	} else if strings.Contains(exp, "<") {
		expPair := strings.Split(exp, "<")
		left, err := strconv.ParseFloat(RemoveStartEndMark(expPair[0]), 64)
		if err != nil {
			return false, err
		}
		right, err := strconv.ParseFloat(RemoveStartEndMark(expPair[1]), 64)
		if err != nil {
			return false, err
		}
		return left < right, nil
	} else if strings.Contains(exp, "=<") {
		expPair := strings.Split(exp, "=<")
		left, err := strconv.ParseFloat(RemoveStartEndMark(expPair[0]), 64)
		if err != nil {
			return false, err
		}
		right, err := strconv.ParseFloat(RemoveStartEndMark(expPair[1]), 64)
		if err != nil {
			return false, err
		}
		return left <= right, nil
	} else if strings.HasPrefix(exp, "startwith") {
		parms := strings.TrimPrefix(exp, "startwith")
		parmsPair := splitFuncParms(parms)
		if len(parmsPair) == 2 {
			return strings.HasPrefix(parmsPair[0], parmsPair[1]), nil
		}
	} else if strings.HasPrefix(exp, "endwith") {
		parms := strings.TrimPrefix(exp, "endwith")
		parmsPair := splitFuncParms(parms)
		if len(parmsPair) == 2 {
			return strings.HasSuffix(parmsPair[0], parmsPair[1]), nil
		}
	} else if strings.HasPrefix(exp, "not startwith") {
		parms := strings.TrimPrefix(exp, "not startwith(")
		parms = strings.TrimSuffix(parms, ")")
		parmsPair := splitFuncParms(parms)
		if len(parmsPair) == 2 {
			return !strings.HasPrefix(parmsPair[0], parmsPair[1]), nil
		}
	} else if strings.HasPrefix(exp, "not endwith") {
		parms := strings.TrimPrefix(exp, "not endwith(")
		parms = strings.TrimSuffix(parms, ")")
		parmsPair := splitFuncParms(parms)
		if len(parmsPair) == 2 {
			return !strings.HasSuffix(parmsPair[0], parmsPair[1]), nil
		}
	}

	return
}

func splitFuncParms(parms string) []string {
	parms = strings.TrimPrefix(parms, "(")
	parms = strings.TrimSuffix(parms, ")")
	parmsPair := strings.Split(parms, ",")
	for i := range parmsPair {
		parmsPair[i] = strings.TrimSpace(parmsPair[i])
	}
	return parmsPair
}
