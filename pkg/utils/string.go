package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

func Contains(origin, target string) bool {
	ignoreChar := []string{".", "_", "-"}
	for _, item := range ignoreChar {
		origin = strings.ReplaceAll(origin, item, "-")
		target = strings.ReplaceAll(target, item, "-")
	}
	return strings.Contains(strings.ToLower(origin), strings.ToLower(target))
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

func EncodingStringToBase64(rawCmd string) string {
	return base64.StdEncoding.EncodeToString([]byte(rawCmd))
}

func DecodingBase64ToString(src string) (dst string, err error) {
	buff, err := base64.StdEncoding.DecodeString(src)
	dst = string(buff)
	return
}

func BuildBase64CmdWithExecutor(sudo bool, rawCmd string, executor string) string {
	return fmt.Sprintf("base64 -d <<< %s | %s %s", EncodingStringToBase64(rawCmd), GetSudoString(sudo), executor)
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

func GetSudoString(sudo bool) string {
	if sudo {
		return "sudo "
	} else {
		return ""
	}
}

func SplitDirPath(filepath string) string {
	pathItems := strings.Split(filepath, "/")
	if len(pathItems) > 2 {
		return strings.Join(pathItems[:len(pathItems)-1], "/")
	}
	return filepath
}

func Logic(input string) (result bool, err error) {
	input = strings.TrimSpace(input)
	if input == "0" || strings.ToLower(input) == "false" || strings.ToLower(input) == "!true" {
		return false, nil
	}
	if input == "1" || strings.ToLower(input) == "true" || strings.ToLower(input) == "!false" {
		return true, nil
	}
	return false, errors.New("can't logic " + input)
}

func CodeBlock(code string) string {
	return fmt.Sprintf("%s\n%s\n%s\n", strings.Repeat("\u2193", 30), code, strings.Repeat("\u2191", 30))
}
