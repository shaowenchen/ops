package shell

import (
	"fmt"

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
var shellDebug bool

var ShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "run shell on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logLevel := log.LevelInfo
		if shellDebug {
			logLevel = log.LevelDebug
		}
		logger, err := log.NewStdFileLogger(false, logLevel)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		inventoryType := utils.GetInventoryType(inventory)
		if utils.IsExistsFile(shellOpt.Content) {
			shellOpt.Content, _ = utils.ReadFile(shellOpt.Content)
		}
		if inventoryType == constants.InventoryTypeKubeconfig {
			KubeShell(logger, shellOpt, kubeOpt, inventory)
		} else if inventoryType == constants.InventoryTypeHosts {
			HostShell(logger, shellOpt, hostOpt, inventory)
		}
	},
}

func KubeShell(logger *log.Logger, shellOpt option.ShellOption, kubeOpt option.KubeOption, inventory string) (err error) {
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
		logger.Info.Println(utils.FilledInMiddle(node.Name))
		kube.Shell(logger, client, node, shellOpt, kubeOpt)
	}
	return
}

func HostShell(logger *log.Logger, shellOpt option.ShellOption, hostOpt option.HostOption, inventory string) (err error) {
	for _, h := range host.GetHosts(logger, hostOpt, inventory) {
		err = host.Shell(logger, h, shellOpt, hostOpt)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	return
}

func init() {
	ShellCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")
	ShellCmd.Flags().BoolVarP(&shellDebug, "debug", "", false, "")

	ShellCmd.Flags().BoolVarP(&shellOpt.Sudo, "sudo", "", false, "")
	ShellCmd.Flags().StringVarP(&shellOpt.Content, "content", "", "", "")
	ShellCmd.MarkFlagRequired("content")

	ShellCmd.Flags().BoolVarP(&kubeOpt.All, "all", "", false, "")
	ShellCmd.Flags().StringVarP(&kubeOpt.NodeName, "nodename", "", "", "")
	ShellCmd.Flags().StringVarP(&kubeOpt.RuntimeImage, "runtimeimage", "", constants.DefaultRuntimeImage, "")

	ShellCmd.Flags().StringVarP(&hostOpt.Username, "username", "", constants.GetCurrentUser(), "")
	ShellCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	ShellCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	ShellCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
	ShellCmd.Flags().IntVar(&hostOpt.Port, "port", 22, "")
}
