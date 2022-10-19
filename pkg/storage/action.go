package storage

import (
	"os"

	"github.com/shaowenchen/opscli/pkg/utils"
)

func ActionS3File(option S3FileOption) (err error) {
	if len(option.AK) == 0 {
		option.AK = os.Getenv("ak")
	}
	if len(option.SK) == 0 {
		option.SK = os.Getenv("sk")
	}
	if len(option.AK) == 0 || len(option.SK) == 0 {
		return utils.LogError("Please provide ak sk in params or env")
	}
	if IsS3UploadFlag(option.Direction) {
		_, err = s3Upload(option.AK, option.SK, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
	} else if IsS3DownloadFlag(option.Direction) {
		err = s3Download(option.AK, option.SK, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
	} else {
		return utils.LogError("direction must is download or upload")
	}
	return utils.LogError(err)
}

func IsS3DownloadFlag(flag string) bool {
	return flag == "download"
}

func IsS3UploadFlag(flag string) bool {
	return flag == "upload"
}
