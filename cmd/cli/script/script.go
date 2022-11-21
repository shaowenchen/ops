package script

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

var scriptOpt option.ScriptOption
var hostOpt option.HostOption

var inventory string

var ScriptCmd = &cobra.Command{
	Use:   "script",
	Short: "run script on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		inventoryType := utils.GetInventoryType(inventory)
		if inventoryType == constants.InventoryTypeKubeconfig {
			KubeScript(logger, scriptOpt, inventory)
		} else if inventoryType == constants.InventoryTypeHosts {
			HostScript(logger, scriptOpt, hostOpt, inventory)
		}
	},
}

func KubeScript(logger *log.Logger, option option.ScriptOption, inventory string) (err error) {
	client, err := utils.NewKubernetesClient(inventory)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	nodeList, err := kube.GetNodes(logger, client, option.KubeOption)
	if err != nil {
		logger.Error.Println(err)
	}
	if len(nodeList) == 0 {
		logger.Info.Println("Please provide a node at least")
	}
	for _, node := range nodeList {
		kube.Script(logger, client, node, option)
	}
	return
}

func HostScript(logger *log.Logger, scriptOpt option.ScriptOption, hostOpt option.HostOption, inventory string) (err error) {
	fmt.Println(inventory)
	for _, h := range host.GetHosts(logger, hostOpt, inventory) {
		err = host.Script(logger, h, scriptOpt, hostOpt)
		if err != nil {
			logger.Error.Println(err)
		} else {
			logger.Info.Println("Successed!")
		}
	}
	return
}

func init() {
	ScriptCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")

	ScriptCmd.Flags().BoolVarP(&scriptOpt.Sudo, "sudo", "", false, "")
	ScriptCmd.Flags().StringVarP(&scriptOpt.Script, "script", "", "", "")
	ScriptCmd.Flags().StringVarP(&scriptOpt.NodeName, "nodename", "", "", "")
	ScriptCmd.Flags().StringVarP(&scriptOpt.RuntimeImage, "runtimeimage", "", "docker.io/library/alpine:latest", "runtime image")

	ScriptCmd.Flags().StringVarP(&hostOpt.Username, "username", "", "", "")
	ScriptCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	ScriptCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	ScriptCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
	ScriptCmd.Flags().IntVar(&hostOpt.Port, "port", 22, "")
}
