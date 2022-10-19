package utils

import "strings"

func SplitStr(str string) (strList []string) {
	return strings.Split(str, ",")
}

func IsContainKey(targets []string, target string) bool {
	for _, item := range targets {
		if item == target {
			return true
		}
	}
	return false
}

func SplitKeyValues(str string) (pair map[string]string) {
	keyLabels := strings.Split(str, ",")
	for _, keyLabel := range keyLabels {
		keyLabelPair := strings.Split(keyLabel, "=")
		if len(keyLabelPair) == 2 {
			if pair == nil {
				pair = make(map[string]string)
			}
			pair[keyLabelPair[0]] = keyLabelPair[1]
		}
	}
	return
}
