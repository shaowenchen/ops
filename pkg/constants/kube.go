package constants

import "path/filepath"

const AllNamespaces = "all"
const OpsNamespace = "ops-system"

const AnyMaster = "anymaster"

const LabelNodeRoleMaster = "node-role.kubernetes.io/master"
const LabelNodeRoleWorker = "node-role.kubernetes.io/worker"

const KubeAdminConfigPath = "/etc/kubernetes/admin.conf"

const DefaultRuntimeImage = "docker.io/library/ubuntu:20.04"

const SyncResourceStatusHeatSeconds = 60

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
