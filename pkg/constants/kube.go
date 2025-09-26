package constants

import (
	"path/filepath"
)

const AllNamespaces = "all"
const OpsNamespace = "ops-system"

const CluterEmptyValue = ""

const LabelNodeRoleMaster = "node-role.kubernetes.io/master"
const LabelNodeRoleControlPlane = "node-role.kubernetes.io/control-plane"

const LabelNodeRoleWorker = "node-role.kubernetes.io/worker"

const LabelOpsTaskKey = "ops/task"

const LabelOpsTaskValue = "true"

const LabelOpsServerKey = "app.kubernetes.io/name"
const LabelOpsServerValue = "ops"

const KubeAdminConfigPath = "/etc/kubernetes/admin.conf"

const DefaultRuntimeImage = "registry.cn-beijing.aliyuncs.com/opshub/ubuntu:22.04"
const OpsCliRuntimeImage = "registry.cn-beijing.aliyuncs.com/opshub/shaowenchen-ops-cli:latest"

const SyncResourceStatusHeatSeconds = 60 * 5
const SyncResourceRandomBiasSeconds = 60 * 2
const SyncCronRandomBias = 30

const MaxResourceConcurrentReconciles = 1
const MaxTaskrunConcurrentReconciles = 50

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

const ModeHost = "host"
const ModeContainer = "container"

const AllNodes = "all"
const AllMasters = "allmasters"
const AllWorkers = "allworkers"

const AnyNode = "anynode"
const AnyMaster = "anymaster"
const AnyWorker = "anyworker"

func IsAnyKubeNode(nodeName string) bool {
	return nodeName == AnyNode || nodeName == AnyMaster || nodeName == AnyWorker
}

func IsAnyNode(nodeName string) bool {
	return nodeName == AnyNode
}

func IsAnyMaster(nodeName string) bool {
	return nodeName == AnyMaster
}

func IsAnyWorker(nodeName string) bool {
	return nodeName == AnyWorker
}
