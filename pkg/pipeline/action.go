package pipeline

import (
	"fmt"
	"github.com/shaowenchen/opscli/pkg/host"
)

func ActionPipeline(option PipelineOption) (err error) {
	pipelines, err := readPipelineYaml(option.FilePath)
	if err != nil {
		fmt.Println(err)
	}
	for _, p := range pipelines {
		fmt.Println("[pipeline] " + p.Name)
		if len(option.Hosts) == 0 {
			option.Hosts = host.LocalHostIP
		}
		for _, addr := range host.RemoveDuplicates(host.GetSliceFromFileOrString(option.Hosts)) {
			for _, s := range p.Steps {
				fmt.Println(fmt.Sprintf("[%s] %s", addr, s.Name))
				stepFunc, err1 := getStepFunc(s)
				if err != nil {
					fmt.Println(err)
					return err1
				}
				var tempOption = option
				tempOption.Hosts = addr
				err1 = stepFunc(s, tempOption)
				if err != nil {
					fmt.Println(err)
					return err1
				}
			}
		}
	}
	return
}
