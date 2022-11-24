package create

import (
	"fmt"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "create task resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		Createtask(logger, clusterOpt, inventory)
	},
}

func Createtask(logger *log.Logger, option option.ClusterOption, inventory string) (err error) {
	inventory = utils.GetAbsoluteFilePath(inventory)
	restConfig, err := utils.GetRestConfig(inventory)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	taskText, err := utils.ReadFile(taskpath)
	if err != nil {
		logger.Error.Println(err)
	}

	t := &opsv1.Task{}
	err = yaml.Unmarshal([]byte(taskText), t)
	if err != nil {
		logger.Error.Println(err)
	}
	t.Namespace = option.Namespace
	t.Name = option.Name
	err = kube.CreateTask(logger, restConfig, t, option.Clear)
	if err != nil {
		logger.Error.Println(err)
	}

	return
}

func init() {
	taskCmd.Flags().StringVarP(&clusterOpt.Kubeconfig, "kubeconfig", "", "~/.kube/config", "")
	taskCmd.Flags().StringVarP(&clusterOpt.Namespace, "namespace", "", "default", "")
	taskCmd.Flags().StringVarP(&clusterOpt.Name, "name", "", "", "")
	taskCmd.MarkFlagRequired("name")
	taskCmd.Flags().BoolVarP(&clusterOpt.Clear, "clear", "", false, "")

	taskCmd.Flags().StringVarP(&inventory, "inventory", "", "", "")
	taskCmd.MarkFlagRequired("inventory")
	taskCmd.Flags().StringVarP(&taskpath, "filepath", "", "", "")
}
