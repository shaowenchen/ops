package utils

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
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
	runner := exec.Command("sudo", "sh", "-c", ShellMv(src, dst))
	if sudo {
		runner = exec.Command("sh", "-c", ShellMv(src, dst))
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

func ReadFile(path string) (buff string, err error) {
	path = GetAbsoluteFilePath(path)
	buffBytes, err := os.ReadFile(path)
	return string(buffBytes), err
}

func GetFileArray(filePath string) (fileArray []string, err error) {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return
	}
	if info.IsDir() {
		files, err := ioutil.ReadDir(filePath)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			fileArray = append(fileArray, filepath.Join(filePath, f.Name()))
		}
	} else {
		fileArray = append(fileArray, filePath)
	}
	return
}
