package utils

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/user"
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

func GetCurrentUserHomeDir() string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDirectory
}

func GetCurrentUser() string {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	return currentUser.Username
}

func GetCurrentUserPrivateKeyPath() string {
	homeDirectory := GetCurrentUserHomeDir()
	return filepath.Join(homeDirectory, ".ssh", "id_rsa")
}

func GetCurrentUserKubeConfigPath() string {
	homeDirectory := GetCurrentUserHomeDir()
	return filepath.Join(homeDirectory, ".kube", "config")
}

func GetAdminKubeConfigPath() string {
	return "/etc/kubernetes/admin.conf"
}

func IsExistsFile(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func FileMD5(path string) (string, error) {
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

func GetSliceFromFileOrString(str string) []string {
	isExist, err := IsExistsFile(str)
	if err != nil {
		return nil
	}
	var result []string
	if isExist {
		// try kubeconfig
		node_ips, err := GetAllNodesFromKubeconfig(str)
		if err == nil {
			return node_ips
		}
		//try readfile
		readFile, err := os.Open(str)
		if err != nil {
			panic(err)
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
	return result
}
