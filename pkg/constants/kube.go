package constants

import "path/filepath"

const AllNamespaces = "all"
const OpsNamespace = "ops"

const NodeKeyRoleMaster = "node-role.kubernetes.io/master"
const NodeKeyRoleWorker = "node-role.kubernetes.io/worker"

const KubeAdminConfigPath = "/etc/kubernetes/admin.conf"

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

func GetCurrentUserConfigPath() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".kube", "config")
}
