package storage

import (
	"fmt"
	"os"

	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
)

func S3File(logger *log.Logger, fileOpt option.FileOption, s3Opt option.S3FileOption) (err error) {
	if len(s3Opt.AK) == 0 {
		s3Opt.AK = os.Getenv("ak")
	}
	if len(s3Opt.SK) == 0 {
		s3Opt.SK = os.Getenv("sk")
	}
	if len(s3Opt.AK) == 0 || len(s3Opt.SK) == 0 {
		logger.Error.Println("Please provide ak sk in params or env")
		return
	}
	fmt.Println(fileOpt.RemoteFile)
	if utils.IsUploadDirection(fileOpt.Direction) {
		_, err = S3Upload(s3Opt.AK, s3Opt.SK, s3Opt.Region, s3Opt.Endpoint, s3Opt.Bucket, fileOpt.LocalFile, fileOpt.RemoteFile)
	} else if utils.IsDownloadDirection(fileOpt.Direction) {
		err = S3Download(s3Opt.AK, s3Opt.SK, s3Opt.Region, s3Opt.Endpoint, s3Opt.Bucket, fileOpt.LocalFile, fileOpt.RemoteFile)
	} else {
		logger.Error.Println("Please provide a valid direction")
	}
	if err != nil {
		logger.Error.Println(err)
	}
	return
}
