package create

import (
	"context"
	"strings"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var hClusterOpt option.ClusterOption
var hHostOpt option.HostOption
var hInventory string
var hVerbose string

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "create host resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(hVerbose).SetStd().SetFile().Build()
		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultTaskStepTimeoutDuration)
		defer cancel()
		err := CreateHost(ctx, logger, hClusterOpt, hHostOpt, hInventory)
		if err != nil {
			logger.Error.Println(err)
			return
		}
	},
}

func CreateHost(ctx context.Context, logger *log.Logger, clusterOpt option.ClusterOption, hostOpt option.HostOption, inventory string) (err error) {
	kubeconfigPath := utils.GetAbsoluteFilePath(clusterOpt.Kubeconfig)
	restConfig, err := utils.GetRestConfig(kubeconfigPath)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if hostOpt.SecretRef == "" && hostOpt.PrivateKey == "" {
		hostOpt.PrivateKey, err = utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(hostOpt.PrivateKey)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
	hs := host.GetHosts(logger, clusterOpt, hostOpt, inventory)

	for _, h := range hs {
		h.Namespace = clusterOpt.Namespace
		// one name, one host
		if len(hs) == 1 {
			if clusterOpt.Name == "" {
				clusterOpt.Name = strings.ReplaceAll(h.Spec.Address, ".", "-")
			}
			hs[0].Name = clusterOpt.Name
		}
		// no name, multi host
		err = kube.CreateHost(ctx, logger, restConfig, h, clusterOpt.Clear)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	return
}

func init() {
	hostCmd.Flags().StringVarP(&hVerbose, "verbose", "v", "", "")
	hostCmd.Flags().StringVarP(&hClusterOpt.Kubeconfig, "kubeconfig", "", constants.GetCurrentUserKubeConfigPath(), "")
	hostCmd.Flags().StringVarP(&hClusterOpt.Namespace, "namespace", "", constants.DefaultOpsNamespace, "")
	hostCmd.Flags().StringVarP(&hClusterOpt.Name, "name", "", "", "")
	hostCmd.Flags().BoolVarP(&hClusterOpt.Clear, "clear", "", false, "")

	hostCmd.Flags().StringVarP(&hHostOpt.Username, "username", "", "root", "")
	hostCmd.Flags().StringVarP(&hHostOpt.Password, "password", "", "", "")
	hostCmd.Flags().StringVarP(&hHostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
	hostCmd.Flags().StringVarP(&hHostOpt.SecretRef, "secretref", "", "", "")
	hostCmd.Flags().StringVarP(&hInventory, "inventory", "i", "", "")
	hostCmd.MarkFlagRequired("inventory")
	hostCmd.Flags().IntVar(&hHostOpt.Port, "port", 22, "")
}
