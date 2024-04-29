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
	Name: "summary",
	CallFunction: func(logger *log.Logger, prManager *agent.LLMPipelineRunsManager, pr *agent.LLMPipelineRun) (output string, err error) {
		prompt := `
# Give brief summaries and suggestions based on the problem and the tasks performed for the problem.
# Don't repeat the question in the answer.
# Do not repeat the result of the task in the answer.
# Do not list the implementation details of each step in your answer
# Answer the question in the language of the question.
`
		input := "Question：" + pr.Desc + "\n"
		tasksOutput := "### Some additional information\n"
		trLength := len(pr.TaskRuns)
		for i, tr := range pr.TaskRuns {
			input += tr.Output + "\n"
			if i < trLength-1 {
				tasksOutput += "#### " + tr.TaskRef + "\n"
				tasksOutput += tr.Output + "\n"
			}
		}
		logger.Debug.Println(tasksOutput)
		client := GetClient(GlobalCopilotOption.Endpoint, GlobalCopilotOption.Key)
		summaryOutput, err := ChatCompletion(logger, client, GlobalCopilotOption.Model, nil, input, prompt, 0)
		if err == nil {
			return summaryOutput, nil
		}
		return tasksOutput, err
	},
}
