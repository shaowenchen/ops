package constants

import (
	"os"
	"strings"
)

const (
	EnvDebugKey           = "DEBUG"
	EnvClusterKey         = "CLUSTER"
	EnvActiveNamespaceKey = "ACTIVE_NAMESPACE"
	EnvDefaultRuntimeImage = "DEFAULT_RUNTIME_IMAGE"
)

func GetEnvActiveNamespace() string {
	return os.Getenv(EnvActiveNamespaceKey)
}

func GetEnvDebug() bool {
	return strings.ToLower(os.Getenv(EnvDebugKey)) == "true"
}

func GetEnvCluster() string {
	return os.Getenv(EnvClusterKey)
}


func GetEnvDefaultRuntimeImage() string{
	return os.Getenv(EnvDefaultRuntimeImage)
}