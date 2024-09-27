package file

import (
	"context"

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
		inventoryType := utils.GetInventoryType(inventory)
		if inventoryType == constants.InventoryTypeHosts {
			HostFile(ctx, logger, fileOpt, hostOpt, inventory)
		} else if inventoryType == constants.InventoryTypeKubernetes {
			KubeFile(ctx, logger, fileOpt, inventory)
		}
	},
}

func HostFile(ctx context.Context, logger *log.Logger, fileOpt option.FileOption, hostOpt option.HostOption, inventory string) (err error) {
	hs := host.GetHosts(logger, option.ClusterOption{}, hostOpt, inventory)
	for _, h := range hs {
		output, err := host.File(ctx, logger, h, hostOpt, fileOpt)
		if err != nil {
			logger.Error.Println(err)
		}
		if len(output) > 0 {
			logger.Info.Println(output)
		}
	}
	return
}

func KubeFile(ctx context.Context, logger *log.Logger, fileOpt option.FileOption, inventory string) (err error) {
	client, err := utils.NewKubernetesClient(inventory)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	nodeList, err := kube.GetNodes(ctx, logger, client, fileOpt.KubeOption)
	if err != nil {
		logger.Error.Println(err)
	}
	if len(nodeList) == 0 {
		logger.Info.Println("Please provide a node at least")
	}
	for _, node := range nodeList {
		kube.File(logger, client, node, fileOpt)
	}
	return
}

func init() {
	FileCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")
	FileCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
	FileCmd.Flags().BoolVarP(&fileOpt.Sudo, "sudo", "", false, "")
	FileCmd.Flags().StringVarP(&fileOpt.LocalFile, "localfile", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.RemoteFile, "remotefile", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.Direction, "direction", "d", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.AesKey, "aeskey", "", storage.UnSetFlag, "if you want to encrypt or decrypt file, please provide a aes key")

	FileCmd.Flags().StringVarP(&fileOpt.Region, "region", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.Endpoint, "endpoint", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.Bucket, "bucket", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.AK, "ak", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.SK, "sk", "", "", "")

	FileCmd.Flags().StringVarP(&fileOpt.Api, "api", "", "", "")

	FileCmd.Flags().StringVarP(&hostOpt.Username, "username", "", constants.GetCurrentUser(), "")
	FileCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	FileCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	FileCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
	FileCmd.Flags().IntVar(&hostOpt.Port, "port", 22, "")

	FileCmd.Flags().StringVarP(&fileOpt.NodeName, "nodename", "", "", "")
	FileCmd.Flags().StringVarP(&fileOpt.RuntimeImage, "runtimeimage", "", constants.OpsCliRuntimeImage, "")
	FileCmd.Flags().StringVarP(&fileOpt.ResNamespace, "opsnamespace", "", constants.DefaultResNamespace, "ops work namespace")
}
