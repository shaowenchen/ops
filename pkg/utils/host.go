package utils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/shaowenchen/ops/pkg/constants"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

func Mv(sudo bool, src, dst string) (stdout string, err error) {
	runner := exec.Command("sudo", "sh", "-c", ScriptMv(src, dst))
	if sudo {
		runner = exec.Command("sh", "-c", ScriptMv(src, dst))
	}
	var out, errout bytes.Buffer
	runner.Stdout = &out
	runner.Stderr = &errout
	err = runner.Run()
	if err != nil {
		stdout = errout.String()
		return
	}
	stdout = out.String()
	return
}

func findIP(input string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock
	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(input)
}

func AnalysisHostsParameter(str string) (result []string, err error) {
	isExist := IsExistsFile(GetAbsoluteFilePath(str))
	if isExist {
		// try kubeconfig
		nodeIPs, err := GetAllNodesByKubeconfig(str)
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
			line = findIP(line)
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
	}
	return path
}
