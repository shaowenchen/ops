package task

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/task"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
)

var taskOpt option.TaskOption
var hostOpt option.HostOption
var kubeOpt option.KubeOption
var inventory string

var TaskCmd = &cobra.Command{
	Use:                "task",
	Short:              "command about task",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		taskOpt = parseArgs(args)
		if len(taskOpt.TaskPath) == 0 {
			fmt.Printf("--taskpath is must provided")
			return
		}
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		inventoryType := utils.GetInventoryType(inventory)
		if inventoryType == constants.InventoryTypeHosts {
			HostTask(logger, taskOpt, hostOpt, inventory)
		} else if inventoryType == constants.InventoryTypeKubeconfig {
			KubeTask(logger, taskOpt, kubeOpt, inventory)
		}
	},
}

func HostTask(logger *log.Logger, taskOpt option.TaskOption, hostOpt option.HostOption, inventory string) (err error) {
	tasks, err := task.ReadTaskYaml(utils.GetAbsoluteFilePath(taskOpt.TaskPath))
	if err != nil {
		logger.Error.Println(err)
		return err

	}

	hs := host.GetHosts(logger, hostOpt, inventory)
	for _, h := range hs {
		if err != nil {
			logger.Error.Println(err)
			continue
		}
		for _, t := range tasks {
			hc, err := host.NewHostConnectionBase64(
				h.Spec.Address,
				h.Spec.Port,
				h.Spec.Username,
				h.Spec.Password,
				h.Spec.PrivateKey,
				h.Spec.PrivateKeyPath,
			)
			if err != nil {
				logger.Error.Println(err)
				continue
			}
			task.RunTaskOnHost(&t, hc, taskOpt)
		}
	}
	return
}

func KubeTask(logger *log.Logger, taskOpt option.TaskOption, kubeOpt option.KubeOption, inventory string) (err error) {
	tasks, err := task.ReadTaskYaml(utils.GetAbsoluteFilePath(taskOpt.TaskPath))
	if err != nil {
		logger.Error.Println(err)
		return err
	}

	kc, err := kube.NewKubeConnection(inventory)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	nodeList, err := kc.GetNodes()
	if err != nil {
		logger.Error.Println(err)
	}
	if len(kubeOpt.NodeName) != 0 {
		nodeList, err = kc.GetNodeByName(kubeOpt.NodeName)
		if err != nil {
			logger.Error.Println(err)
		}
	}
	for _, node := range nodeList.Items {
		for _, t := range tasks {
			task.RunTaskOnKube(&t, kc, &node, taskOpt, kubeOpt)
		}
	}
	return
}

func parseArgs(args []string) (taskOption option.TaskOption) {
	taskOption.Variables = make(map[string]string)
	for i := 0; i < len(args); i++ {
		fieldName := getArgName(args[i])
		if len(fieldName) > 0 {
			fieldValue := "true"
			if (i + 1) == len(args) {
				// --clear
			} else if (i+1) < len(args) && len(getArgName(args[i+1])) > 0 {
				// --clear --username root
			} else {
				// --username root
				fieldValue = args[i+1]
			}
			if fieldName == "debug" {
				taskOption.Debug = fieldValue == "true"
			} else if fieldName == "sudo" {
				taskOption.Sudo = fieldValue == "true"
			} else if fieldName == "taskpath" {
				taskOption.TaskPath = fieldValue
			} else if fieldName == "nodename" {
				kubeOpt.NodeName = fieldValue
			} else if fieldName == "runtimeimage" {
				kubeOpt.RuntimeImage = fieldValue
			} else if fieldName == "inventory" || fieldName == "i" {
				inventory = fieldValue
			} else if fieldName == "port" {
				hostOpt.Port, _ = strconv.Atoi(fieldValue)
			} else if fieldName == "username" {
				hostOpt.Username = fieldValue
			} else if fieldName == "password" {
				hostOpt.Password = fieldValue
			} else if fieldName == "privatekeypath" {
				hostOpt.PrivateKeyPath = fieldValue
			} else {
				taskOption.Variables[fieldName] = fieldValue
			}
		}
	}
	return
}

func getArgName(arg string) string {
	if strings.HasPrefix(arg, "--") {
		return arg[2:]
	} else if strings.HasPrefix(arg, "-") {
		return arg[1:]
	}
	return ""
}

func init() {
	TaskCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "")

	TaskCmd.Flags().BoolVarP(&taskOpt.Debug, "debug", "", false, "")
	TaskCmd.Flags().StringVarP(&taskOpt.TaskPath, "taskpath", "", "", "")
	TaskCmd.MarkFlagRequired("taskpath")

	TaskCmd.Flags().StringVarP(&kubeOpt.NodeName, "nodename", "", "", "")
	TaskCmd.Flags().StringVarP(&kubeOpt.RuntimeImage, "runtimeimage", "", constants.DefaultRuntimeImage, "runtime image")
	TaskCmd.Flags().IntVarP(&hostOpt.Port, "port", "", 22, "")
	TaskCmd.Flags().StringVarP(&hostOpt.Username, "username", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", "", "")
}
