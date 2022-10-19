package utils

import (
	"errors"
	"fmt"
)

func PrintError(errs ...interface{}) (err error) {
	errMsg := fmt.Sprint(errs)
	fmt.Println(errMsg)
	return errors.New(errMsg)
}

func PrintInfo(infos ...interface{}) {
	errMsg := fmt.Sprint(infos)
	fmt.Println(errMsg)
}
