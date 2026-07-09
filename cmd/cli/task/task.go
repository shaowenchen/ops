package task

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/cmd/cli/config"
	"github.com/shaowenchen/ops/cmd/cli/internal/complete"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	opstask "github.com/shaowenchen/ops/pkg/task"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var taskOpt option.TaskOption
var hostOpt option.HostOption
var kubeOpt option.KubeOption
var inventory string

var TaskCmd = &cobra.Command{
	Use:                "task",
	Short:              "command about task",
	DisableFlagParsing: true,
	ValidArgsFunction:  complete.TaskRunValidArgs,
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
		inventoryType, availableInventory := utils.GetInventoryType(inventory, kubeOpt.NodeName)
		tasks, err := opstask.ReadTaskYaml(utils.GetTaskAbsoluteFilePath(taskOpt.Proxy, taskOpt.FilePath))
		if err != nil {
			logger.Error.Println(err)
			return
		}
		for _, task := range tasks {
			if inventoryType == constants.InventoryTypeHosts && !task.NeedKubeExecution() {
				HostTask(context.Background(), logger, task, taskOpt, hostOpt, availableInventory)
			} else {
				KubeTask(context.Background(), logger, task, taskOpt, kubeOpt, availableInventory)
			}
		}
	},
}

func HostTask(ctx context.Context, logger *log.Logger, t opsv1.Task, taskOpt option.TaskOption, hostOpt option.HostOption, inventory string) (err error) {
	hs := host.GetHosts(logger, option.ClusterOption{}, hostOpt, inventory)
	var wg sync.WaitGroup
	for _, h := range hs {
		wg.Add(1)
		go func(h *opsv1.Host) {
			defer wg.Done()
			tr := opsv1.NewTaskRun(&t)
			hc, err := host.NewHostConnBase64(h)
			if err != nil {
				logger.Error.Println(err)
				return
			}
			newTaskOpt := withDefaultVariables(taskOpt, map[string]string{
				"host":     h.GetHostname(),
				"nodename": kubeOpt.NodeName,
				"proxy":    taskOpt.Proxy,
			})
			err = opstask.RunTaskOnHost(ctx, logger, &t, &tr, hc, newTaskOpt)
			if err != nil {
				logger.Error.Println(err)
			}
		}(h)
	}
	wg.Wait()
	return
}

func KubeTask(ctx context.Context, logger *log.Logger, t opsv1.Task, taskOpt option.TaskOption, kubeOpt option.KubeOption, inventory string) (err error) {
	kc, err := kube.NewKubeConnection(inventory)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	nodes, err := kube.GetNodes(ctx, logger, kc.Client, kubeOpt)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	if len(nodes) == 0 {
		if kubeOpt.NodeName != "" {
			logger.Error.Printf("Node '%s' not found", kubeOpt.NodeName)
		} else {
			logger.Error.Println("No nodes found")
		}
		return fmt.Errorf("no nodes found")
	}
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(node corev1.Node) {
			defer wg.Done()
			newKubeOpt := kubeOpt
			if t.Spec.RuntimeImage != "" {
				newKubeOpt.RuntimeImage = t.Spec.RuntimeImage
			}
			newTaskOpt := withDefaultVariables(taskOpt, map[string]string{
				"host":     node.GetName(),
				"nodename": kubeOpt.NodeName,
				"proxy":    taskOpt.Proxy,
			})

			// Convert Task mounts to MountConfig with variable rendering
			mountConfigs := make([]option.MountConfig, 0)
			// Prepare variables for mount rendering
			mountVars, err := opstask.GetRealVariables(&t, newTaskOpt)
			if err != nil {
				logger.Error.Println(err)
				return
			}
			for _, taskMount := range t.Spec.Mounts {
				// Render variables in mount fields
				renderedMount := opstask.RenderTaskMount(&taskMount, mountVars, nil)

				mountConfig := option.MountConfig{}
				if renderedMount.Secret != nil {
					// Secret mount
					mountConfig.Secret = &option.SecretMountConfig{
						Name:      renderedMount.Secret.Name,
						MountPath: renderedMount.Secret.MountPath,
					}
				} else if renderedMount.ConfigMap != nil {
					// ConfigMap mount
					mountConfig.ConfigMap = &option.ConfigMapMountConfig{
						Name:      renderedMount.ConfigMap.Name,
						MountPath: renderedMount.ConfigMap.MountPath,
					}
				} else {
					// HostPath mount
					mountConfig.HostPath = renderedMount.HostPath
					mountConfig.MountPath = renderedMount.MountPath
				}
				mountConfigs = append(mountConfigs, mountConfig)
			}
			// Copy existing mounts and append Task mounts
			if len(mountConfigs) > 0 {
				newKubeOpt.Mounts = make([]option.MountConfig, len(kubeOpt.Mounts))
				copy(newKubeOpt.Mounts, kubeOpt.Mounts)
				newKubeOpt.Mounts = append(newKubeOpt.Mounts, mountConfigs...)
			}

			tr := opsv1.NewTaskRun(&t)
			err = opstask.RunTaskOnKube(logger, &t, &tr, kc, &node, newTaskOpt, newKubeOpt)
			if err != nil {
				logger.Error.Println(err)
			}
		}(node)
	}
	wg.Wait()
	return
}

func parseArgs(args []string) (taskOption option.TaskOption) {
	taskOption.Variables = make(map[string]string)
	runtimeImageSetViaCLI := false
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
				// CLI argument has highest priority, set it directly
				taskOption.Proxy = fieldValue
				taskOption.Variables["proxy"] = fieldValue
			} else if fieldName == "nodename" {
				kubeOpt.NodeName = fieldValue
				taskOption.Variables["nodename"] = fieldValue
			} else if fieldName == "opsnamespace" {
				kubeOpt.Namespace = fieldValue
			} else if fieldName == "runtimeimage" {
				kubeOpt.RuntimeImage = fieldValue
				runtimeImageSetViaCLI = true
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
	// Get proxy with priority: CLI > ENV > Config > Default
	// If taskOption.Proxy is empty, it means not provided via CLI args
	if taskOption.Proxy == "" {
		taskOption.Proxy = config.GetValueWithPriority("", constants.EnvProxy, "proxy", constants.DefaultProxy)
	}
	// Get runtimeimage with priority: CLI > ENV > Config > Default
	// If not set via CLI, use priority: ENV > Config > Default
	if !runtimeImageSetViaCLI {
		kubeOpt.RuntimeImage = config.GetValueWithPriority("", constants.EnvDefaultRuntimeImage, "runtimeimage", constants.DefaultRuntimeImage)
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

func withDefaultVariables(taskOpt option.TaskOption, defaults map[string]string) option.TaskOption {
	taskOpt.Variables = cloneStringMap(taskOpt.Variables)
	taskOpt.DefaultVariables = cloneStringMap(taskOpt.DefaultVariables)
	if taskOpt.DefaultVariables == nil {
		taskOpt.DefaultVariables = make(map[string]string)
	}
	for k, v := range defaults {
		taskOpt.DefaultVariables[k] = v
	}
	return taskOpt
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func init() {
	TaskCmd.Flags().StringVarP(&inventory, "inventory", "i", "", "inventory file (hosts list or kubeconfig)")

	TaskCmd.Flags().StringVarP(&taskOpt.FilePath, "filepath", "f", "", "task YAML (basename under ~/.ops/tasks or path)")

	TaskCmd.Flags().StringVarP(&kubeOpt.NodeName, "nodename", "", "", "target Kubernetes node name")
	TaskCmd.Flags().StringVarP(&kubeOpt.Namespace, "opsnamespace", "", constants.OpsNamespace, "ops work namespace")

	// Load runtimeimage with priority: ENV > Config > Default (CLI args handled in parseArgs)
	runtimeImage := config.GetValueWithPriority("", constants.EnvDefaultRuntimeImage, "runtimeimage", constants.DefaultRuntimeImage)
	TaskCmd.Flags().StringVarP(&kubeOpt.RuntimeImage, "runtimeimage", "", runtimeImage, "runtime image")

	TaskCmd.Flags().IntVarP(&hostOpt.Port, "port", "", 22, "SSH port for host inventory")
	TaskCmd.Flags().StringVarP(&hostOpt.Username, "username", "", constants.GetCurrentUser(), "SSH username for host inventory")
	TaskCmd.Flags().StringVarP(&hostOpt.Password, "password", "", "", "SSH password for host inventory")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKey, "privatekey", "", "", "base64 private key (prefer --privatekeypath)")
	TaskCmd.Flags().StringVarP(&hostOpt.PrivateKeyPath, "privatekeypath", "", constants.GetCurrentUserPrivateKeyPath(), "SSH private key file")
}
