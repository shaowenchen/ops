package file

import (
	"context"
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
var serverOpt option.FileServerOption
var inventory string
var verbose string

var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "transfer between local and remote file",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultShellTimeoutDuration)
		defer cancel()
		// upstream is server
		if serverOpt.Api != "" {
			ServerFile(logger, fileOpt, serverOpt)
			return
		}
		// upstream is s3、image、local
		fileOpt.Filling()
		if fileOpt.StorageType == constants.RemoteStorageTypeS3 {
			S3File(logger, fileOpt, s3Opt)
		} else if fileOpt.StorageType == constants.RemoteStorageTypeImage {
			KubeFile(ctx, logger, fileOpt, kubeOpt, inventory)
		} else if fileOpt.StorageType == constants.RemoteStorageTypeLocal {
			HostFile(ctx, logger, fileOpt, hostOpt, inventory)
		}
	},
}

func HostFile(ctx context.Context, logger *log.Logger, fileOpt option.FileOption, hostOpt option.HostOption, inventory string) (err error) {
	hs := host.GetHosts(logger, option.ClusterOption{}, hostOpt, inventory)
	if utils.IsDownloadDirection(fileOpt.Direction) && len(hs) != 1 {
		errMsg := "need only one host while downloading"
		logger.Error.Println(errMsg)
		return errors.New(errMsg)
	}
	for _, h := range hs {
		err = host.File(ctx, logger, h, fileOpt, hostOpt)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	return
}

func KubeFile(ctx context.Context, logger *log.Logger, fileOpt option.FileOption, kubeOpt option.KubeOption, inventory string) (err error) {
	client, err := utils.NewKubernetesClient(inventory)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	nodeList, err := kube.GetNodes(ctx, logger, client, kubeOpt)
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

func ServerFile(logger *log.Logger, option option.FileOption, serverOpt option.FileServerOption) (err error) {
	return storage.ServerFile(logger, option, serverOpt)
}

func init() {
	FileCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")
	FileCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
	FileCmd.Flags().BoolVarP(&fileOpt.Sudo, "sudo", "", false, "")
	FileCmd.Flags().StringVarP(&fileOpt.LocalFile, "localfile", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.RemoteFile, "remotefile", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.Direction, "direction", "d", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.AesKey, "aeskey", "", storage.UnSetFlag, "if you want to encrypt or decrypt file, please provide a aes key")

	FileCmd.Flags().StringVarP(&hostOpt.Username, "username", "", constants.GetCurrentUser(), "")
	FileCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	FileCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	FileCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
	FileCmd.Flags().IntVar(&hostOpt.Port, "port", 22, "")

	FileCmd.Flags().StringVarP(&kubeOpt.NodeName, "nodename", "", "", "")
	FileCmd.Flags().StringVarP(&kubeOpt.OpsNamespace, "opsnamespace", "", constants.DefaultOpsNamespace, "ops work namespace")

	FileCmd.Flags().StringVarP(&s3Opt.Region, "region", "", "ap-southeast-3", "")
	FileCmd.Flags().StringVarP(&s3Opt.Endpoint, "endpoint", "", "obs.ap-southeast-3.myhuaweicloud.com", "")
	FileCmd.Flags().StringVarP(&s3Opt.Bucket, "bucket", "", "", "")
	FileCmd.Flags().StringVarP(&s3Opt.AK, "ak", "", "", "")
	FileCmd.Flags().StringVarP(&s3Opt.SK, "sk", "", "", "")

	FileCmd.Flags().StringVarP(&serverOpt.Api, "api", "", "", "")
}
