package task

import (
	"context"
	"strconv"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	opstask "github.com/shaowenchen/ops/pkg/task"
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
		taskOpt = parseArgs(args)
		logger := log.NewLogger().SetVerbose("debug").SetStd().SetFile().Build()
		if len(taskOpt.FilePath) == 0 {
			logger.Error.Println("--filepath is must provided")
			return
		}
		inventory = utils.GetAbsoluteFilePath(inventory)
		hostOpt.Password = utils.EncodingStringToBase64(hostOpt.Password)
		privateKey, _ := utils.ReadFile(hostOpt.PrivateKeyPath)
		hostOpt.PrivateKey = utils.EncodingStringToBase64(privateKey)
		inventoryType := utils.GetInventoryType(inventory)
		tasks, err := opstask.ReadTaskYaml(utils.GetTaskAbsoluteFilePath(taskOpt.Proxy, taskOpt.FilePath))
		if err != nil {
			logger.Error.Println(err)
			return
		}
		taskOpt.Variables["nodename"] = kubeOpt.NodeName
		switch inventoryType {
		case constants.InventoryTypeHosts:
			HostTask(context.Background(), logger, tasks, taskOpt, hostOpt, inventory)
		case constants.InventoryTypeKubernetes:
			KubeTask(context.Background(), logger, tasks, taskOpt, kubeOpt, inventory)
		}
	},
}

func HostTask(ctx context.Context, logger *log.Logger, tasks []opsv1.Task, taskOpt option.TaskOption, hostOpt option.HostOption, inventory string) (err error) {
	hs := host.GetHosts(logger, option.ClusterOption{}, hostOpt, inventory)
	for _, h := range hs {
		if err != nil {
			logger.Error.Println(err)
			continue
		}
		for _, t := range tasks {
			tr := opsv1.NewTaskRun(&t)
			hc, err := host.NewHostConnBase64(h)
			if err != nil {
				logger.Error.Println(err)
				continue
			}
			newTaskOpt := taskOpt
			newTaskOpt.Variables["host"] = h.GetHostname()
			newTaskOpt.Variables["proxy"] = taskOpt.Proxy
			err = opstask.RunTaskOnHost(ctx, logger, &t, &tr, hc, newTaskOpt)
			if err != nil {
				logger.Error.Println(err)
				continue
			}
		}
	}
	return
}

func KubeTask(ctx context.Context, logger *log.Logger, tasks []opsv1.Task, taskOpt option.TaskOption, kubeOpt option.KubeOption, inventory string) (err error) {
	kc, err := kube.NewKubeConnection(inventory)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	nodes, err := kube.GetNodes(ctx, logger, kc.Client, kubeOpt)
	for _, node := range nodes {
		for _, t := range tasks {
			newKubeOpt := kubeOpt
			if t.Spec.RuntimeImage != "" {
				newKubeOpt.RuntimeImage = t.Spec.RuntimeImage
			}
			for k, v := range t.Spec.Variables {
				if _, ok := taskOpt.Variables[k]; !ok {
					taskOpt.Variables[k] = v.GetValue()
				}
			}
			taskOpt.Variables["host"] = node.GetName()
			taskOpt.Variables["proxy"] = taskOpt.Proxy
			tr := opsv1.NewTaskRun(&t)
			err = opstask.RunTaskOnKube(logger, &t, &tr, kc, &node, taskOpt, newKubeOpt)
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
			} else if fieldName == "proxy" {
				taskOption.Proxy = fieldValue
			} else if fieldName == "nodename" {
				kubeOpt.NodeName = fieldValue
			} else if fieldName == "opsnamespace" {
				kubeOpt.Namespace = fieldValue
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
			if taskOption.Proxy == "" {
				taskOption.Proxy = constants.DefaultProxy
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

	TaskCmd.Flags().StringVarP(&kubeOpt.NodeName, "nodename", "", "", "")
	TaskCmd.Flags().StringVarP(&kubeOpt.Namespace, "opsnamespace", "", constants.OpsNamespace, "ops work namespace")
	TaskCmd.Flags().StringVarP(&kubeOpt.RuntimeImage, "runtimeimage", "", constants.DefaultRuntimeImage, "runtime image")

	TaskCmd.Flags().IntVarP(&hostOpt.Port, "port", "", 22, "")
	TaskCmd.Flags().StringVarP(&hostOpt.Username, "username", "", constants.GetCurrentUser(), "")
	TaskCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "")
}
