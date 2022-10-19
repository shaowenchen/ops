package utils

import (
	"errors"
	"fmt"
)

func LogError(errs ...interface{}) error {
	if len(errs) == 1 && errs[0] == nil {
		return nil
	}
	msg := fmt.Sprint(errs...)
	fmt.Println(msg)
	return errors.New(msg)
}

func LogInfo(infos ...interface{}) {
	msg := fmt.Sprint(infos...)
	fmt.Println(msg)
}
