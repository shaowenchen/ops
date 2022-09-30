package host

import (
	"bufio"
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
