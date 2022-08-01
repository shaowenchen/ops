package kube

import "strings"

const AllNamespacesFlag = "all"
const OpsCliNamespace = "opscli"
const NodeKeyRoleMaster = "node-role.kubernetes.io/master"
const NodeKeyRoleWorker = "node-role.kubernetes.io/worker"

func IsAllNamespacesFlag(flag string) bool{
	if strings.ToLower(flag) == AllNamespacesFlag{
		return true
	}
	return false
}

const (
    ContainersReady string = "ContainersReady"
    PodInitialized  string = "Initialized"
    PodReady   string = "Ready"
    PodScheduled  string = "PodScheduled"
)

const (
    ConditionTrue    string = "True"
    ConditionFalse   string = "False"
    ConditionUnknown string = "Unknown"
)