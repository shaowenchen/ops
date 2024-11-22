package constants

import (
	"os"
	"strings"
)

const (
	EnvDebugKey            = "DEBUG"
	EnvActiveNamespaceKey  = "ACTIVE_NAMESPACE"
	EnvDefaultRuntimeImage = "DEFAULT_RUNTIME_IMAGE"
	EnvEventClusterKey     = "EVENT_CLUSTER"
	EnvEventEndpointKey    = "EVENT_ENDPOINT"
)

// just for controller

func GetEnvActiveNamespace() string {
	return os.Getenv(EnvActiveNamespaceKey)
}

func GetEnvDebug() bool {
	return strings.ToLower(os.Getenv(EnvDebugKey)) == "true"
}

func GetEnvEventCluster() string {
	return os.Getenv(EnvEventClusterKey)
}

func GetEnvEventEndpoint() string {
	return os.Getenv(EnvEventEndpointKey)
}

func GetEnvDefaultRuntimeImage() string {
	return os.Getenv(EnvDefaultRuntimeImage)
}
