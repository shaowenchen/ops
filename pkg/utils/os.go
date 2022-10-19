package utils

import (
	"os"
	"runtime"
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

func GetRuntimeInfo() (info map[string]string) {
	info = make(map[string]string, 0)
	info["ARCH"] = runtime.GOARCH
	info["OS"] = runtime.GOOS
	return
}
