package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
)

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
	if len(strings.TrimSpace(str)) == 0 {
		return []string{}
	}
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

func BuildBase64Cmd(sudo bool, rawCmd string) string {
	return fmt.Sprintf("base64 -d <<< %s | %s sh", EncodingBase64(rawCmd), GetSudoString(sudo))
}

func RemoveStartEndMark(raw string) string {
	for _, item := range []string{" ", "'", "\""} {
		raw = strings.Trim(raw, item)
	}
	return raw
}

func MergeMap(target map[string]string, needMerge map[string]string) map[string]string {
	for key, value := range needMerge {
		if len(strings.TrimSpace(value)) > 0 {
			target[key] = value
		}
	}
	return target
}

func PrintMiddleFilled(text string) string {
	return PrintMiddle(text, "*")
}

func PrintMiddle(text string, fill string) string {
	total := 59
	if len(text)+1 >= total {
		return text
	}
	if len(fill) != 1 {
		return text
	}
	leftLen := (total - len(text)) / 2
	return fmt.Sprintf("%s%s%s", strings.Repeat(fill, leftLen), text, strings.Repeat(fill, (total-leftLen-len(text))))
}

func GetSudoString(sudo bool) string {
	if sudo {
		return "sudo"
	} else {
		return ""
	}
}

func IsUploadDirection(direction string) bool {
	return strings.Contains(strings.ToLower(direction), "up")
}

func IsDownloadDirection(direction string) bool {
	return strings.Contains(strings.ToLower(direction), "down")
}
