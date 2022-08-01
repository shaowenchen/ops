package host

import (
	"os"
	"strings"
)

func PathExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}

func SplitStr(str string) []string{
	return strings.Split(str, ",")
}