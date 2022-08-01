package docker

import (
	"fmt"
	"github.com/docker/docker/client"
)

func ActionClear(option ClearOption) (err error) {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	logs, err := ImageClear(dockerCli, option.Force, option.NameRegx, option.TagRegx)
	if len(logs) > 0 {
		fmt.Println(logs)
	}
	report, err := ImagesPrune(dockerCli)
	fmt.Println("Prune delete images num is ", len(report.ImagesDeleted))
	for _, deleteImage := range report.ImagesDeleted {
		fmt.Println(deleteImage.Deleted)
	}
	return
}
