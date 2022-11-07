package task

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/task"
	"github.com/shaowenchen/ops/pkg/action"
	"github.com/spf13/cobra"
)

var taskOption task.TaskOption

var TaskCmd = &cobra.Command{
	Use:                "task",
	Short:              "command about task",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		logger, err := log.NewDefaultLogger(true, true)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		taskOption = parseArgs(args)
		if len(taskOption.FilePath) == 0 {
			fmt.Printf("--filepath is must provided")
			return
		}
		action.Task(logger, taskOption)
	},
}

func parseArgs(args []string) (taskOption task.TaskOption) {
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
			} else if fieldName == "filepath" {
				taskOption.FilePath = fieldValue
			} else if fieldName == "hosts" {
				taskOption.Hosts = fieldValue
			} else if fieldName == "port" {
				taskOption.Port, _ = strconv.Atoi(fieldValue)
			} else if fieldName == "username" {
				taskOption.Username = fieldValue
			} else if fieldName == "password" {
				taskOption.Password = fieldValue
			} else if fieldName == "privatekeypath" {
				taskOption.PrivateKeyPath = fieldValue
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
	}
	return ""
}

func init() {
	TaskCmd.Flags().BoolVarP(&taskOption.Debug, "debug", "", false, "")
	TaskCmd.Flags().StringVarP(&taskOption.FilePath, "filepath", "", "", "")
	TaskCmd.MarkFlagRequired("filepath")
	TaskCmd.Flags().StringVarP(&taskOption.Hosts, "hosts", "", "", "")
	TaskCmd.Flags().IntVarP(&taskOption.Port, "port", "", 22, "")
	TaskCmd.Flags().StringVarP(&taskOption.Username, "username", "", "", "")
	TaskCmd.Flags().StringVarP(&taskOption.Password, "password", "", "", "")
	TaskCmd.Flags().StringVarP(&taskOption.PrivateKeyPath, "privatekeypath", "", "", "")
}
