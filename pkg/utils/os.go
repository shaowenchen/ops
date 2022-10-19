package utils

import (
	"os"
	"strings"
)


func GetAllOsEnv() (envs map[string]string) {
	envs = make(map[string]string, 0)
	for _, keyValue := range os.Environ() {
		pair := strings.Split(keyValue, "=")
		if len(pair) == 1 {
			envs[pair[0]] = ""
		} else if len(pair) == 2 {
			envs[pair[0]] = pair[1]
		}
	}
	return
}
