package docker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func ImageClear(dockerCli *client.Client, force bool, nameRegx, tagRegx string) (logs string, err error) {
	logsList := []string{}
	images, err := dockerCli.ImageList(context.TODO(), types.ImageListOptions{
		All:     true,
		Filters: filters.NewArgs(),
	})
	for _, image := range images {
		for _, imageTag := range image.RepoTags {
			splitImageTag := strings.Split(imageTag, ":")
			lenSplitImageTag := len(splitImageTag)
			//if registry don't use 443 port
			if lenSplitImageTag > 2 {
				splitImageTag[0] = strings.Join(splitImageTag[:lenSplitImageTag-1], ":")
				splitImageTag[1] = splitImageTag[lenSplitImageTag-1]
			}
			needDelete := false
			if len(nameRegx) > 0 {
				needDelete, err = regexp.MatchString(nameRegx, splitImageTag[0])
				if err != nil {
					return
				}
				if !needDelete {
					continue
				}
			}
			if len(tagRegx) > 0 {
				needDelete = false
				needDelete, err = regexp.MatchString(tagRegx, splitImageTag[1])
				if err != nil {
					return
				}
			}
			if needDelete {
				logsList = append(logsList, fmt.Sprintf("Deleting image %s:%s", splitImageTag[0], splitImageTag[1]))
				_, err = dockerCli.ImageRemove(context.TODO(), imageTag, types.ImageRemoveOptions{
					Force:         force,
					PruneChildren: true,
				})
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
	return strings.Join(logsList, "\n"), err
}

func ImagesPrune(dockerCli *client.Client) (report types.ImagesPruneReport, err error) {
	report, err = dockerCli.ImagesPrune(context.TODO(), filters.NewArgs())
	return
}
