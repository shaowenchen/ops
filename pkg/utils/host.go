package utils

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"github.com/shaowenchen/ops/pkg/constants"
	"io"
	"os"
	"path/filepath"
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
	info["arch"] = runtime.GOARCH
	info["os"] = runtime.GOOS
	return
}

func IsExistsFile(filepath string) bool {
	s, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	if s.IsDir() {
		return false
	}

	return true
}

func CreateDir(dirpath string) error {
	return os.MkdirAll(dirpath, os.ModePerm)
}

func FileMD5(path string) (string, error) {
	path = GetAbsoluteFilePath(path)
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	m := md5.New()
	if _, err := io.Copy(m, file); err != nil {
		return "", err
	}

	fileMd5 := fmt.Sprintf("%x", m.Sum(nil))
	return fileMd5, nil
}

func AnalysisHostsParameter(str string) (result []string, err error) {
	isExist := IsExistsFile(GetAbsoluteFilePath(str))
	if isExist {
		// try kubeconfig
		nodeIPs, err := GetAllNodesFromKubeconfig(str)
		if err == nil {
			return nodeIPs, nil
		}
		//try readfile
		readFile, err := os.Open(str)
		if err != nil {
			return result, err
		}
		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			line := strings.TrimSpace(fileScanner.Text())
			if len(line) > 0 {
				result = append(result, line)
			}
		}
		readFile.Close()
	} else {
		result = SplitStrings(str)
	}
	if len(result) == 0 {
		result = append(result, constants.LocalHostIP)
	}
	return RemoveDuplicates(result), nil
}

func GetAbsoluteFilePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
		return path
	} else if strings.HasPrefix(path, "./") {
		dirname, _ := os.Getwd()
		path = filepath.Join(dirname, path[2:])
		return path
	} else if !strings.HasPrefix(path, "/") {
		dirname, _ := os.Getwd()
		path = filepath.Join(dirname, path)
		return path
	}
	return path
}
