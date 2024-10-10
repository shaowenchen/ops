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

const LabelOpsTaskKey = "ops"

const LabelOpsTaskValue = "task"

const KubeAdminConfigPath = "/etc/kubernetes/admin.conf"

const DefaultRuntimeImage = "ac2-registry.cn-hangzhou.cr.aliyuncs.com/ac2/base:ubuntu22.04"
const OpsCliRuntimeImage = "registry.cn-hangzhou.aliyuncs.com/shaowenchen/opscli:latest"

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

const AllNodes = "all"
const AllMasters = "allmasters"
const AllWorkers = "allworkers"

const AnyNode = "anynode"
const AnyMaster = "anymaster"
const AnyWorker = "anyworker"
