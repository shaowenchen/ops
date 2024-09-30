package create

import (
	"context"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var tTaskOpt option.TaskOption
var tClusterOpt option.ClusterOption
var tVerbose string

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "create task resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger().SetVerbose(tVerbose).SetStd().SetFile().Build()
		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultShellTimeoutDuration)
		defer cancel()
		Createtask(ctx, logger)
	},
}

func Createtask(ctx context.Context, logger *log.Logger) (err error) {
	kubeconfigPath := utils.GetAbsoluteFilePath(tClusterOpt.Kubeconfig)
	restConfig, err := utils.GetRestConfig(kubeconfigPath)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	taskText, err := utils.ReadFile(utils.GetTaskAbsoluteFilePath(tTaskOpt.Proxy, tTaskOpt.FilePath))
	if err != nil {
		logger.Error.Println(err)
	}
	t := &opsv1.Task{}
	err = yaml.Unmarshal([]byte(taskText), t)
	if err != nil {
		logger.Error.Println(err)
	}
	if tClusterOpt.Name != "" {
		t.Name = tClusterOpt.Name
	}
	if tClusterOpt.Namespace != "" {
		t.Namespace = tClusterOpt.Namespace
	}
	if t.Namespace == "" {
		t.Namespace = constants.DefaultNamespace
	}
	err = kube.CreateTask(ctx, logger, restConfig, t, tTaskOpt.Clear)
	if err != nil {
		logger.Error.Println(err)
	}

	return
}

func init() {
	taskCmd.Flags().StringVarP(&tVerbose, "verbose", "v", "", "")
	taskCmd.Flags().StringVarP(&tClusterOpt.Kubeconfig, "kuebconfig", "", constants.GetCurrentUserKubeConfigPath(), "")
	taskCmd.Flags().StringVarP(&tClusterOpt.Namespace, "namespace", "", constants.DefaultNamespace, "")
	taskCmd.Flags().StringVarP(&tClusterOpt.Name, "name", "", "", "")

	taskCmd.Flags().BoolVarP(&tTaskOpt.Clear, "clear", "", false, "")
	taskCmd.Flags().StringVarP(&tTaskOpt.FilePath, "inventory", "i", "", "")
	taskCmd.MarkFlagRequired("inventory")
	taskCmd.Flags().StringVarP(&tTaskOpt.Proxy, "proxy", "", constants.DefaultProxy, "")
}
