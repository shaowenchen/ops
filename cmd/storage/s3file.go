package storage

import (
	"github.com/shaowenchen/opscli/pkg/storage"
	"github.com/spf13/cobra"
)

// kubeCmd represents the kube command
var s3FileOption storage.S3FileOption

var s3FileCmd = &cobra.Command{
	Use:   "s3file",
	Short: "operate file in S3",
	Run: func(cmd *cobra.Command, args []string) {
		storage.ActionS3File(s3FileOption)
	},
}

func init() {
	s3FileCmd.Flags().StringVarP(&s3FileOption.Region, "region", "", "ap-southeast-3", "")
	s3FileCmd.Flags().StringVarP(&s3FileOption.Endpoint, "endpoint", "", "obs.ap-southeast-3.myhuaweicloud.com", "")
	s3FileCmd.Flags().StringVarP(&s3FileOption.Bucket, "bucket", "", "", "")
	s3FileCmd.Flags().StringVarP(&s3FileOption.AK, "ak", "", "", "")
	s3FileCmd.Flags().StringVarP(&s3FileOption.SK, "sk", "", "", "")
	s3FileCmd.Flags().StringVarP(&s3FileOption.LocalFile, "localfile", "", "", "e.g., myfile.zip")
	s3FileCmd.MarkFlagRequired("localfile")
	s3FileCmd.Flags().StringVarP(&s3FileOption.RemoteFile, "remotefile", "", "", "e.g.,archived/myfile.zip")
	s3FileCmd.MarkFlagRequired("remotefile")
}
