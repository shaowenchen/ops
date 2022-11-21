package storage

import (
	"os"

	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
)

func S3File(logger *log.Logger, option option.FileOption, s3option option.S3FileOption) (err error) {
	if len(s3option.AK) == 0 {
		s3option.AK = os.Getenv("ak")
	}
	if len(s3option.SK) == 0 {
		s3option.SK = os.Getenv("sk")
	}
	if len(s3option.AK) == 0 || len(s3option.SK) == 0 {
		logger.Error.Println("Please provide ak sk in params or env")
		return
	}
	if utils.IsUploadDirection(option.Direction) {
		_, err = S3Upload(s3option.AK, s3option.SK, s3option.Region, s3option.Endpoint, s3option.Bucket, option.LocalFile, option.RemoteFile)
	} else if utils.IsDownloadDirection(option.Direction) {
		err = S3Download(s3option.AK, s3option.SK, s3option.Region, s3option.Endpoint, s3option.Bucket, option.LocalFile, option.RemoteFile)
	}
	logger.Error.Println("direction is not correct")
	return
}
