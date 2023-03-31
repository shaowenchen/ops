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
		left, err := strconv.Atoi(RemoveStartEndMark(expPair[0]))
		if err != nil {
			return false, err
		}
		right, err := strconv.Atoi(RemoveStartEndMark(expPair[1]))
		if err != nil {
			return false, err
		}
		return left > right, nil
	} else if strings.Contains(exp, ">=") {
		expPair := strings.Split(exp, ">=")
		left, err := strconv.Atoi(RemoveStartEndMark(expPair[0]))
		if err != nil {
			return false, err
		}
		right, err := strconv.Atoi(RemoveStartEndMark(expPair[1]))
		if err != nil {
			return false, err
		}
		return left >= right, nil
	} else if strings.Contains(exp, "<") {
		expPair := strings.Split(exp, "<")
		left, err := strconv.Atoi(RemoveStartEndMark(expPair[0]))
		if err != nil {
			return false, err
		}
		right, err := strconv.Atoi(RemoveStartEndMark(expPair[1]))
		if err != nil {
			return false, err
		}
		return left < right, nil
	} else if strings.Contains(exp, "=<") {
		expPair := strings.Split(exp, "=<")
		left, err := strconv.Atoi(RemoveStartEndMark(expPair[0]))
		if err != nil {
			return false, err
		}
		right, err := strconv.Atoi(RemoveStartEndMark(expPair[1]))
		if err != nil {
			return false, err
		}
		return left <= right, nil
	}

	return
}
