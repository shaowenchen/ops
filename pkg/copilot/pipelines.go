package copilot

import (
	"github.com/shaowenchen/ops/pkg/agent"
)

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
			Name:   "summary",
			Always: true,
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
			Name:   "summary",
			Always: true,
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
			Name:   "summary",
			Always: true,
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
			Name:   "summary",
			Always: true,
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
			Name:   "summary",
			Always: true,
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
			Name:   "summary",
			Always: true,
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
			Name:   "summary",
			Always: true,
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
