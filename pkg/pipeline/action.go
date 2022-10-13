package pipeline

import "fmt"

func ActionPipeline(option PipelineOption) (err error) {
	pipelines, err := readPipelineYaml(option.FilePath)
	if err != nil {
		fmt.Println(err)
	}
	for pi, p := range pipelines {
		fmt.Println(fmt.Sprintf("pipeline(%d/%d) -> %s", pi + 1, len(pipelines), p.Name))
		for si, s := range p.Steps {
			fmt.Println(fmt.Sprintf("  step(%d/%d) -> %s", si + 1, len(p.Steps), s.Name))
			stepFunc, err1 := getStepFunc(s)
			if err != nil {
				fmt.Println(err)
				return err1
			}
			err1 = stepFunc(s, option)
			if err != nil {
				fmt.Println(err)
				return err1
			}
		}
	}
	return
}
