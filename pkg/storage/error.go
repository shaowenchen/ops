package storage

import "fmt"

func PrintError(errMsg string)(err error){
	fmt.Println(errMsg)
	return fmt.Errorf(errMsg)
}

func ErrorMsgS3File(err error) string {
	return fmt.Sprintf("could not S3file: %v", err)
}

func ErrorMsgS3AKSK(err error) string {
	return fmt.Sprintf("Please set OPSCLIAK and OPSCLIAK in Env %v", err)
}