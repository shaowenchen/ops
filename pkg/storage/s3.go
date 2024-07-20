package storage

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"os"
)

func S3File(fileOpt option.FileOption) (output string, err error) {
	if len(fileOpt.AK) == 0 {
		fileOpt.AK = os.Getenv("ak")
	}
	if len(fileOpt.SK) == 0 {
		fileOpt.SK = os.Getenv("sk")
	}
	if len(fileOpt.AK) == 0 || len(fileOpt.SK) == 0 {
		err = errors.New("Please provide ak sk in params or env")
		return
	}
	if fileOpt.IsUploadDirection() {
		output, err = S3Upload(fileOpt.AK, fileOpt.SK, fileOpt.Region, fileOpt.Endpoint, fileOpt.Bucket, fileOpt.LocalFile, fileOpt.RemoteFile)
	} else if fileOpt.IsDownloadDirection() {
		output, err = S3Download(fileOpt.AK, fileOpt.SK, fileOpt.Region, fileOpt.Endpoint, fileOpt.Bucket, fileOpt.LocalFile, fileOpt.RemoteFile)
	} else {
		output = "invalid direction"
		err = errors.New(output)
		return
	}
	return
}

func S3Upload(ak, sk, region, endpoint, bucket, localFilePath, remoteFile string) (output string, err error) {
	sess, _ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(ak, sk, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	})
	uploader := s3manager.NewUploader(sess)
	file, err := os.Open(localFilePath)
	if err != nil {
		return
	}
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(remoteFile),
		Body:   file,
	},
		func(u *s3manager.Uploader) {
			u.PartSize = 10 * 1024 * 1024
			u.LeavePartsOnError = true
			u.Concurrency = 3
		},
	)
	if err != nil {
		return
	}
	output = "success upload " + localFilePath + " to s3 " + result.Location
	return
}

func S3Download(ak, sk, region, endpoint, bucket, localFilePath, remoteFile string) (output string, err error) {
	sess, _ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(ak, sk, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	})
	downloader := s3manager.NewDownloader(sess)
	file, err := utils.CreateFile(localFilePath)
	if err != nil {
		return
	}
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(remoteFile),
	}
	_, err = downloader.Download(file, params)
	if err != nil {
		return
	}
	output = "success download s3 " + remoteFile + " to " + localFilePath
	return
}
