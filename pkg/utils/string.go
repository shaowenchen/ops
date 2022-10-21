package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
)

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

func SplitStrings(str string) []string {
	return strings.Split(str, ",")
}

func RemoveDuplicates(origin []string) []string {
	var result []string
	status := make(map[string]string, len(origin))
	for _, key := range origin {
		if _, ok := status[key]; !ok {
			result = append(result, key)
			status[key] = key
		}
	}
	return result
}

func EncodingBase64(rawCmd string) string {
	return base64.StdEncoding.EncodeToString([]byte(rawCmd))
}

func BuildBase64Cmd(rawCmd string) string {
	return fmt.Sprintf("base64 -d <<< %s | sudo sh", EncodingBase64(rawCmd))
}

func RemoveStartEndMark(raw string) string{
	for _, item := range []string{" ", "'", "\""}{
		raw = strings.Trim(raw, item)
	}
	return raw
}
