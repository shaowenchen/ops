package constants

import (
	"path/filepath"
)

const AllNamespaces = "all"
const DefaultOpsNamespace = "ops-system"
const DefaultNamespace = "default"

const CurrentRuntime = "current"

const LabelNodeRoleMaster = "node-role.kubernetes.io/master"
const LabelNodeRoleControlPlane = "node-role.kubernetes.io/control-plane"

const LabelNodeRoleWorker = "node-role.kubernetes.io/worker"

const LabelOpsTaskKey = "ops"

const LabelOpsTaskValue = "task"

const KubeAdminConfigPath = "/etc/kubernetes/admin.conf"

const DefaultRuntimeImage = "docker.io/library/ubuntu:20.04"

const SyncResourceStatusHeatSeconds = 60 * 5

const MaxConcurrentReconciles = 1

const (
	ContainersReady string = "ContainersReady"
	PodInitialized  string = "Initialized"
	PodReady        string = "Ready"
	PodScheduled    string = "PodScheduled"
)

const (
	ConditionTrue    string = "True"
	ConditionFalse   string = "False"
	ConditionUnknown string = "Unknown"
)

func GetCurrentUserKubeConfigPath() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".kube", "config")
}

const AnyMaster = "anymaster"
