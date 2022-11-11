package storage

import (
	"os"

	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
)

func S3File(logger *log.Logger, option S3FileOption) (err error) {
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
	if utils.IsUploadDirection(option.Direction) {
		_, err = S3Upload(option.AK, option.SK, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
	} else if utils.IsDownloadDirection(option.Direction) {
		err = S3Download(option.AK, option.SK, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
	}
	logger.Error.Println("direction is not correct")
	return
}
