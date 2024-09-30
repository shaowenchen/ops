package constants

import (
	"os"
	"strings"
)

const (
	EnvDebugKey            = "DEBUG"
	EnvClusterKey          = "CLUSTER"
	EnvWebhookUrl          = "WEBHOOK_URL"
	EnvActiveNamespaceKey  = "ACTIVE_NAMESPACE"
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
func GetEnvWebhookUrl() string {
	return os.Getenv(EnvWebhookUrl)
}

func GetEnvDefaultRuntimeImage() string {
	return os.Getenv(EnvDefaultRuntimeImage)
}
