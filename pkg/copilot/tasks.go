package copilot

import (
	"github.com/shaowenchen/ops/pkg/agent"
	"github.com/shaowenchen/ops/pkg/log"
)

var AllTasks = []agent.LLMTask{
	taskListClusters,
	taskListTasks,
	taskListPipelines,
	taskAppSummary,
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
	Name: "app-summary",
	CallFunction: func(logger *log.Logger, prManager *agent.LLMPipelineRunsManager, pr *agent.LLMPipelineRun) (output string, err error) {
		prompt := `
# give brief summaries and suggestions based on the problem and the tasks performed for the problem.
# Don't repeat the question in the answer.
# do not repeat the result of the task in the answer.
# do not list the implementation details of each step in your answer
`
		input := "My Question：" + pr.Desc + ", and did the following\n"
		output = "### run following tasks\n"
		trLength := len(pr.TaskRuns)
		for i, tr := range pr.TaskRuns {
			input += tr.Output + "\n"
			if i < trLength-1 {
				output += "#### " + tr.TaskRef + "\n"
				output += tr.Output + "\n"
			}
		}
		client := GetClient(GlobalCopilotOption.Endpoint, GlobalCopilotOption.Key)
		output, err = ChatCompletion(logger, client, GlobalCopilotOption.Model, nil, input, prompt, 0.3)
		if err == nil {
			output += "### Summary\n" + output
		}
		return output, err
	},
}
