package storage

import (
	"os"

	"github.com/shaowenchen/opscli/pkg/log"
	"github.com/shaowenchen/opscli/pkg/utils"
)

func ActionS3File(logger *log.Logger, option S3FileOption) (err error) {
	if len(option.AK) == 0 {
		option.AK = os.Getenv("ak")
	}
	if len(option.SK) == 0 {
		option.SK = os.Getenv("sk")
	}
	if len(option.AK) == 0 || len(option.SK) == 0 {
		logger.Error.Println("Please provide ak sk in params or env")
		return
	}
	if IsUpload(option.LocalFile) {
		_, err = s3Upload(option.AK, option.SK, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
	} else {
		err = s3Download(option.AK, option.SK, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
	}
	return err
}

func IsUpload(localfile string) bool {
	isExist, err := utils.IsExistsFile(localfile)
	if err != nil && isExist {
		return true
	}
	return false
}
