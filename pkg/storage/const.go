package storage

import "strings"

const EnvS3AsKey = "OPSCLIAK"
const EnvS3SkKey = "OPSCLISK"

const S3DownloadFlag = "down"
const S3UploadFlag = "up"

func IsS3DownloadFlag(flag string) bool{
	return strings.Contains(strings.ToLower(flag), S3DownloadFlag)
}

func IsS3UploadFlag(flag string) bool{
	return strings.Contains(strings.ToLower(flag), S3UploadFlag)
}