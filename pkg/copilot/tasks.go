package copilot

import (
	"fmt"
	"github.com/shaowenchen/ops/pkg/agent"
	"github.com/shaowenchen/ops/pkg/log"
)

var AllTasks = []agent.LLMTask{
	taskListClusters,
	taskListTasks,
	taskListPipelines,
	taskAppSummary,
	taskHelp,
}

var taskListClusters = agent.LLMTask{
	Desc: "list clusters",
	Name: "list-clusters",
	CallFunction: func(logger *log.Logger, prManager *agent.LLMPipelineRunsManager, pr *agent.LLMPipelineRun) (output string, err error) {
		return prManager.GetClusterManager().GetMarkdown(), nil
	},
}

var taskListTasks = agent.LLMTask{
	Desc: "list tasks",
	Name: "list-tasks",
	CallFunction: func(logger *log.Logger, prManager *agent.LLMPipelineRunsManager, pr *agent.LLMPipelineRun) (output string, err error) {
		return prManager.GetTaskRunManager().GetMarkdown(), nil
	},
}

var taskListPipelines = agent.LLMTask{
	Desc: "list pipelines",
	Name: "list-pipelines",
	CallFunction: func(logger *log.Logger, prManager *agent.LLMPipelineRunsManager, pr *agent.LLMPipelineRun) (output string, err error) {
		return prManager.GetMarkdown(), nil
	},
}

var taskAppSummary = agent.LLMTask{
	Desc: "summary",
	Name: "summary",
	CallFunction: func(logger *log.Logger, prManager *agent.LLMPipelineRunsManager, pr *agent.LLMPipelineRun) (output string, err error) {
		output = "### Summary\n"
		trLength := len(pr.TaskRuns)
		for i := 0; i < trLength-1; i++ {
			t := pr.TaskRuns[i]
			t.Output = ShortOutput(t.Output)
			output += "#### " + t.TaskRef + "\n"
			output += t.Output + "\n"
		}
		return
	},
}

func ShortOutput(output string) string {
	if len(output) > 1000 {
		return output[:500] + "..." + output[len(output)-500:]
	}
	return output
}

var taskHelp = agent.LLMTask{
	Desc: "Help",
	Name: "help",
	CallFunction: func(logger *log.Logger, prManager *agent.LLMPipelineRunsManager, pr *agent.LLMPipelineRun) (output string, err error) {
		output = `## Help: `
		for _, p := range AllPipelines {
			output += fmt.Sprintf("\n- %s \n", p.Desc)
		}
		return
	},
}
