package storage

import (
	"github.com/shaowenchen/opscli/pkg/utils"
	"os"
)

func ActionS3File(option S3FileOption) (err error) {
	ak := os.Getenv(EnvS3AsKey)
	sk := os.Getenv(EnvS3SkKey)
	if len(ak) > 0 && len(sk) > 0 {
		if IsS3UploadFlag(option.Direction) {
			_, err = s3Upload(ak, sk, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
		} else {
			err = s3Download(ak, sk, option.Region, option.Endpoint, option.Bucket, option.LocalFile, option.RemoteFile)
		}
		if err != nil {
			err = utils.PrintError(err)
		}
		return
	}
	return nil
}
