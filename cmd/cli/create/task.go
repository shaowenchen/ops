package create

import (
	"fmt"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/create"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var taskOption create.TaskOption

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "create task resource",
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewCliLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		Createtask(logger, taskOption)
	},
}

func Createtask(logger *log.Logger, option create.TaskOption) (err error) {
	option.Kubeconfig = utils.GetAbsoluteFilePath(option.Kubeconfig)
	restConfig, err := utils.GetRestConfig(option.Kubeconfig)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	taskText, err := utils.ReadFile(option.Filepath)
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
	t.Spec.HostRef = option.HostRef
	err = create.CreateTask(logger, restConfig, t, option.Clear)
	if err != nil {
		logger.Error.Println(err)
	}

	return
}

func init() {
	taskCmd.Flags().StringVarP(&taskOption.Kubeconfig, "kubeconfig", "", "~/.kube/config", "")
	taskCmd.Flags().StringVarP(&taskOption.Namespace, "namespace", "", "default", "")
	taskCmd.Flags().StringVarP(&taskOption.Name, "name", "", "", "")
	taskCmd.MarkFlagRequired("name")
	taskCmd.Flags().StringVarP(&taskOption.HostRef, "hostref", "", "", "")
	taskCmd.MarkFlagRequired("hostref")
	taskCmd.Flags().StringVarP(&taskOption.Filepath, "filepath", "", "", "")
	taskCmd.Flags().BoolVarP(&taskOption.Clear, "clear", "", false, "")
}
