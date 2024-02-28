package create

import (
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "create task resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(verbose).SetStd().SetFile().Build()
		Createtask(logger)
	},
}

func Createtask(logger *log.Logger) (err error) {
	kubeconfigPath := utils.GetAbsoluteFilePath(clusterOpt.Kubeconfig)
	restConfig, err := utils.GetRestConfig(kubeconfigPath)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	taskText, err := utils.ReadFile(utils.GetTaskAbsoluteFilePath(taskOpt.Proxy, taskOpt.FilePath))
	if err != nil {
		logger.Error.Println(err)
	}
	t := &opsv1.Task{}
	err = yaml.Unmarshal([]byte(taskText), t)
	if err != nil {
		logger.Error.Println(err)
	}
	if clusterOpt.Name != "" {
		t.Name = clusterOpt.Name
	}
	if clusterOpt.Namespace != "" {
		t.Namespace = clusterOpt.Namespace
	}
	if t.Namespace == "" {
		t.Namespace = constants.DefaultOpsNamespace
	}
	err = kube.CreateTask(logger, restConfig, t, taskOpt.Clear)
	if err != nil {
		logger.Error.Println(err)
	}

	return
}

func init() {
	taskCmd.Flags().StringVarP(&verbose, "verbose", "v", "", "")
	taskCmd.Flags().StringVarP(&clusterOpt.Kubeconfig, "kuebconfig", "", "~/.kube/config", "")
	taskCmd.Flags().StringVarP(&clusterOpt.Namespace, "namespace", "", "", "")
	taskCmd.Flags().StringVarP(&clusterOpt.Name, "name", "", "", "")

	taskCmd.Flags().BoolVarP(&taskOpt.Clear, "clear", "", false, "")
	taskCmd.Flags().StringVarP(&taskOpt.FilePath, "inventory", "i", "", "")
	taskCmd.MarkFlagRequired("inventory")
	taskCmd.Flags().StringVarP(&taskOpt.Proxy, "proxy", "", constants.DefaultProxy, "")
}
