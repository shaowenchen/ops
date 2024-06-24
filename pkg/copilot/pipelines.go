package copilot

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/shaowenchen/ops/pkg/agent"
)

func GetDefaultToolCall() (call openai.ToolCall) {
	call.Function = openai.FunctionCall{
		Name:      agent.ClearUnavailableChar(pipelineHelp.Name),
		Arguments: `{}`,
	}
	return call
}

var AllPipelines = []agent.LLMPipeline{
	pipelineListCluster,
	pipelineListTask,
	pipelineListPipeline,
	pipelineRestartPod,
	pipelineRestartPodForce,
	pipelineGetClusterIP,
	pipelineClearDisk,
	pipelineHelp,
}

var pipelineListCluster = agent.LLMPipeline{
	Desc:      "Query -  list K8s cluster",
	Namespace: "ops-system",
	Name:      "list-cluster",
	NodeName:  "anymaster",
	LLMTasks: []agent.LLMTask{
		{
			Name: "list-clusters",
		},
		{
			Name: "summary",
		},
	},
}

var pipelineListTask = agent.LLMPipeline{
	Desc:      "Query -  list task",
	Namespace: "ops-system",
	Name:      "list-task",
	NodeName:  "anymaster",
	LLMTasks: []agent.LLMTask{
		{
			Name: "list-tasks",
		},
		{
			Name: "summary",
		},
	},
}

var pipelineListPipeline = agent.LLMPipeline{
	Desc:      "Query -  list pipeline",
	Namespace: "ops-system",
	Name:      "list-pipeline",
	NodeName:  "anymaster",
	LLMTasks: []agent.LLMTask{
		{
			Name: "list-pipelines",
		},
		{
			Name: "summary",
		},
	},
}

var pipelineRestartPod = agent.LLMPipeline{
	Desc:      "Restart - Restart、delete Pod",
	Namespace: "ops-system",
	Name:      "restart-pod",
	NodeName:  "anymaster",
	Variables: []agent.VariablePair{
		{
			Key:      "podname",
			Desc:     "For example, `pod: long-v1-64cf8d5478-5zsvk or name: long-v1-64cf8d5478-5zsvk`, where long-v1-64cf8d5478-5zsvk is podname",
			Required: true,
		},
	},
	LLMTasks: []agent.LLMTask{
		{
			Name: "check-pod-existed",
		},
		{
			Name: "delete-pod",
		},
		{
			Name: "summary",
		},
	},
}

var pipelineRestartPodForce = agent.LLMPipeline{
	Desc:      "Restart - Restart、delete Pod By Force",
	Namespace: "ops-system",
	Name:      "force-restart-pod",
	NodeName:  "anymaster",
	Variables: []agent.VariablePair{
		{
			Key:      "podname",
			Desc:     "For example, `pod: long-v1-64cf8d5478-5zsvk or name: long-v1-64cf8d5478-5zsvk`, where long-v1-64cf8d5478-5zsvk is podname",
			Required: true,
		},
	},
	LLMTasks: []agent.LLMTask{
		{
			Name: "check-pod-existed",
		},
		{
			Name: "delete-pod-force",
		},
		{
			Name: "summary",
		},
	},
}

var pipelineGetClusterIP = agent.LLMPipeline{
	Desc:      "Query - query cluster IP",
	Namespace: "ops-system",
	Name:      "get-cluster-ip",
	NodeName:  "anymaster",
	Variables: []agent.VariablePair{
		{
			Key:      "clusterip",
			Desc:     "For example, `clusterip: 244.178.44.111`, where 244.178.44.111 is clusterip",
			Required: true,
		},
	},
	LLMTasks: []agent.LLMTask{
		{
			Name: "inspect-clusterip",
		},
		{
			Name: "summary",
		},
	},
}

var pipelineClearDisk = agent.LLMPipeline{
	Desc:      "Clear - clear disk",
	Namespace: "ops-system",
	Name:      "clear-disk",
	Variables: []agent.VariablePair{
		{
			Key:      "nodeName",
			Desc:     "For example, `node:ai-node-4090-73` where ai-node-4090-73 is nodeName",
			Required: true,
		},
	},
	LLMTasks: []agent.LLMTask{
		{
			Name: "clear-disk",
		},
		{
			Name: "summary",
		},
	},
}

var pipelineHelp = agent.LLMPipeline{
	Desc:      "Help - help",
	Namespace: "ops-system",
	Name:      "help",
	NodeName:  "",
	LLMTasks: []agent.LLMTask{
		{
			Name: "help",
		},
	},
}
