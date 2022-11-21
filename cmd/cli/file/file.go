package file

import (
	"fmt"

	"errors"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/storage"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var hostOpt option.HostOption
var fileOpt option.FileOption
var kubeOpt option.KubeOption
var s3Opt option.S3FileOption
var inventory string

var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "transfer between local and remote file",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		remoteFiletype := fileOpt.GetStorageType()
		if remoteFiletype == constants.RemoteStorageTypeS3 {
			err = S3File(logger, fileOpt, s3Opt)
		} else if remoteFiletype == constants.RemoteStorageTypeImage {
			err = KubeFile(logger, fileOpt, kubeOpt, inventory)
		} else if remoteFiletype == constants.RemoteStorageTypeLocal {
			err = HostFile(logger, fileOpt, hostOpt, inventory)
		}
	},
}

func HostFile(logger *log.Logger, option option.FileOption, hostOpt option.HostOption, inventory string) (err error) {
	hs := host.GetHosts(logger, hostOpt, inventory)
	if utils.IsDownloadDirection(option.Direction) && len(hs) != 1 {
		errMsg := "need only one host while downloading"
		logger.Error.Println(errMsg)
		return errors.New(errMsg)
	}
	for _, h := range hs {
		err = host.File(logger, h, fileOpt, hostOpt)
		if err != nil {
			logger.Error.Println(err)
		} else {
			logger.Info.Println("Successed!")
		}
	}
	return
}

func KubeFile(logger *log.Logger, fileOpt option.FileOption, kubeOpt option.KubeOption, inventory string) (err error) {
	client, err := utils.NewKubernetesClient(inventory)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	nodeList, err := kube.GetNodes(logger, client, kubeOpt)
	if err != nil {
		logger.Error.Println(err)
	}
	if len(nodeList) == 0 {
		logger.Info.Println("Please provide a node at least")
	}
	for _, node := range nodeList {
		kube.File(logger, client, node, fileOpt, kubeOpt)
	}
	return
}

func S3File(logger *log.Logger, option option.FileOption, s3option option.S3FileOption) (err error) {
	return storage.S3File(logger, option, s3option)
}

func init() {
	FileCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")

	FileCmd.Flags().BoolVarP(&fileOpt.Sudo, "sudo", "", false, "")
	FileCmd.Flags().StringVarP(&fileOpt.LocalFile, "localfile", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.RemoteFile, "remotefile", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.Direction, "direction", "d", "", "")

	FileCmd.Flags().StringVarP(&hostOpt.Username, "username", "", "", "")
	FileCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	FileCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	FileCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
	FileCmd.Flags().IntVar(&hostOpt.Port, "port", 22, "")

	FileCmd.Flags().StringVarP(&s3Opt.Region, "region", "", "ap-southeast-3", "")
	FileCmd.Flags().StringVarP(&s3Opt.Endpoint, "endpoint", "", "obs.ap-southeast-3.myhuaweicloud.com", "")
	FileCmd.Flags().StringVarP(&s3Opt.Bucket, "bucket", "", "", "")
	FileCmd.Flags().StringVarP(&s3Opt.AK, "ak", "", "", "")
	FileCmd.Flags().StringVarP(&s3Opt.SK, "sk", "", "", "")
}
