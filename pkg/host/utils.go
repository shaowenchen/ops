package host

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
)

func isExistsFile(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func SplitStrings(str string) []string {
	return strings.Split(str, ",")
}

func GetSliceFromFileOrString(str string) []string {
	isExist, err := isExistsFile(str)
	if err != nil {
		return nil
	}
	var result []string
	if isExist {
		readFile, err := os.Open(str)
		if err != nil {
			panic(err)
		}
		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			result = append(result, strings.TrimSpace(fileScanner.Text()))
		}
		readFile.Close()
	} else {
		result = SplitStrings(str)
	}
	return result
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
