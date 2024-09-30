package shell

import (
	"context"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var shellOpt option.ShellOption
var kubeOpt option.KubeOption
var hostOpt option.HostOption

var inventory string
var verbose string

var ShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "run shell on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		inventoryType := utils.GetInventoryType(inventory)
		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultShellTimeoutDuration)
		defer cancel()
		if utils.IsExistsFile(shellOpt.Content) {
			shellOpt.Content, _ = utils.ReadFile(shellOpt.Content)
		}
		if inventoryType == constants.InventoryTypeKubernetes {
			KubeShell(ctx, logger, shellOpt, kubeOpt, inventory)
		} else if inventoryType == constants.InventoryTypeHosts {
			HostShell(ctx, logger, shellOpt, hostOpt, inventory)
		}
	},
}

func KubeShell(ctx context.Context, logger *log.Logger, shellOpt option.ShellOption, kubeOpt option.KubeOption, inventory string) (err error) {
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
		logger.Error.Println("no node found")
		return
	}
	for _, node := range nodeList {
		kube.Shell(logger, client, node, shellOpt, kubeOpt)
	}
	return
}

func HostShell(ctx context.Context, logger *log.Logger, shellOpt option.ShellOption, hostOpt option.HostOption, inventory string) (err error) {
	for _, h := range host.GetHosts(logger, option.ClusterOption{}, hostOpt, inventory) {
		host.Shell(ctx, logger, h, shellOpt, hostOpt)
	}
	return
}

func init() {
	ShellCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")
	ShellCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")

	ShellCmd.Flags().BoolVarP(&shellOpt.Sudo, "sudo", "", false, "")
	ShellCmd.Flags().StringVarP(&shellOpt.Content, "content", "", "", "")
	ShellCmd.MarkFlagRequired("content")

	ShellCmd.Flags().BoolVarP(&kubeOpt.All, "all", "", false, "")
	ShellCmd.Flags().StringVarP(&kubeOpt.NodeName, "nodename", "", "", "")
	ShellCmd.Flags().StringVarP(&kubeOpt.Namespace, "opsnamespace", "", constants.DefaultNamespace, "ops work namespace")
	ShellCmd.Flags().StringVarP(&kubeOpt.RuntimeImage, "runtimeimage", "", constants.DefaultRuntimeImage, "")

	ShellCmd.Flags().StringVarP(&hostOpt.Username, "username", "", constants.GetCurrentUser(), "")
	ShellCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	ShellCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	ShellCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
	ShellCmd.Flags().IntVar(&hostOpt.Port, "port", 22, "")
}
