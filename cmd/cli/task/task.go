package task

import (
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
var verbose string

var TaskCmd = &cobra.Command{
	Use:                "task",
	Short:              "command about task",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		taskOpt = parseArgs(args)
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()
		if len(taskOpt.FilePath) == 0 {
			logger.Error.Println("--filepath is must provided")
			return
		}
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		inventoryType := utils.GetInventoryType(inventory)
		if inventoryType == constants.InventoryTypeHosts {
			HostTask(logger, taskOpt, hostOpt, inventory)
		} else if inventoryType == constants.InventoryTypeKubernetes {
			KubeTask(logger, taskOpt, kubeOpt, inventory)
		}
	},
}

func HostTask(logger *log.Logger, taskOpt option.TaskOption, hostOpt option.HostOption, inventory string) (err error) {
	tasks, err := task.ReadTaskYaml(utils.GetAbsoluteFilePath(taskOpt.FilePath))
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
			hc, err := host.NewHostConnBase64(h)
			if err != nil {
				logger.Error.Println(err)
				continue
			}
			err = task.RunTaskOnHost(logger, &t, hc, taskOpt)
			if err != nil {
				logger.Error.Println(err)
				continue
			}
		}
	}
	return
}

func KubeTask(logger *log.Logger, taskOpt option.TaskOption, kubeOpt option.KubeOption, inventory string) (err error) {
	tasks, err := task.ReadTaskYaml(utils.GetAbsoluteFilePath(taskOpt.FilePath))
	if err != nil {
		logger.Error.Println(err)
		return err
	}

	kc, err := kube.NewKubeConnection(inventory)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	nodes, err := kube.GetNodes(logger, kc.Client, kubeOpt)
	for _, node := range nodes {
		for _, t := range tasks {
			err = task.RunTaskOnKube(logger, &t, kc, &node, taskOpt, kubeOpt)
			if err != nil {
				logger.Error.Println(err)
			}
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
			if fieldName == "sudo" {
				taskOption.Sudo = fieldValue == "true"
			} else if fieldName == "filepath" || fieldName == "f" {
				taskOption.FilePath = fieldValue
			} else if fieldName == "verbose" || fieldName == "v" {
				verbose = fieldValue
			} else if fieldName == "all" {
				kubeOpt.All = fieldValue == "true"
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

	TaskCmd.Flags().StringVarP(&taskOpt.FilePath, "filepath", "", "", "")
	TaskCmd.MarkFlagRequired("filepath")

	TaskCmd.Flags().BoolVarP(&kubeOpt.All, "all", "", false, "")
	TaskCmd.Flags().StringVarP(&kubeOpt.NodeName, "nodename", "", "", "")
	TaskCmd.Flags().StringVarP(&kubeOpt.RuntimeImage, "runtimeimage", "", constants.DefaultRuntimeImage, "runtime image")

	TaskCmd.Flags().IntVarP(&hostOpt.Port, "port", "", 22, "")
	TaskCmd.Flags().StringVarP(&hostOpt.Username, "username", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
}
