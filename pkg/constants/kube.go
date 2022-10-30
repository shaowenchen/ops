package constants

const AllNamespaces = "all"
const OpsCliNamespace = "opscli"

const NodeKeyRoleMaster = "node-role.kubernetes.io/master"
const NodeKeyRoleWorker = "node-role.kubernetes.io/worker"


const (
	KubeAdminConfigPath =  "/etc/kubernetes/admin.conf"
	CurrentUserConfigPath =  "~/.kube/config"
)

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

const HostMountDir = "/tmp/opscli"
