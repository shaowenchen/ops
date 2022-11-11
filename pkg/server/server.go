package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/task"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func GetTask(c *gin.Context) {
	taskExample := v1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-task",
			Namespace: "default",
		},
		Spec: v1.TaskSpec{
			Name:     "This is a task example",
			Desc:     "This desc about thie task",
			Hostname: "127.0.0.1",
			Steps: []v1.Step{
				{
					Name:   "Show OS info",
					Script: "uname -a",
				},
			},
		},
	}
	c.JSON(http.StatusOK, taskExample)
}

func CreateTask(c *gin.Context) {
	dataBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	t := &v1.Task{}
	err = json.Unmarshal(dataBytes, t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	logger, err := log.NewDefaultLogger(true, true)
	if err != nil {
		return
	}
	task.RunTaskOnHost(logger, *t, nil, task.TaskOption{})
	c.JSON(http.StatusOK, "OK")
}
