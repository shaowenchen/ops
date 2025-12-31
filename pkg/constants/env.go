package constants

import (
	"os"
	"strconv"
	"strings"
)

const (
	EnvDebugKey                    = "DEBUG"
	EnvActiveNamespaceKey          = "ACTIVE_NAMESPACE"
	EnvDefaultRuntimeImage         = "DEFAULT_RUNTIME_IMAGE"
	EnvProxy                       = "PROXY"
	EnvEventClusterKey             = "EVENT_CLUSTER"
	EnvEventEndpointKey            = "EVENT_ENDPOINT"
	EnvEventQueryTimeoutKey        = "EVENT_QUERY_TIMEOUT"
	EnvEventListSubjectsTimeoutKey = "EVENT_LIST_SUBJECTS_TIMEOUT"
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

func GetEnvProxy() string {
	return os.Getenv(EnvProxy)
}

// GetEventQueryTimeout returns the event query timeout with priority:
// 1. Environment variable EVENT_QUERY_TIMEOUT (in seconds)
// 2. Config file value (passed as parameter, in seconds)
// 3. Default value (10 seconds)
func GetEventQueryTimeout(configValue uint) uint {
	// Priority 1: Environment variable
	timeout := os.Getenv(EnvEventQueryTimeoutKey)
	if timeout != "" {
		result, err := strconv.ParseUint(timeout, 10, 32)
		if err == nil {
			return uint(result)
		}
	}
	// Priority 2: Config file value
	if configValue > 0 {
		return configValue
	}
	// Priority 3: Default value
	return 10
}

// GetEventListSubjectsTimeout returns the event list subjects timeout with priority:
// 1. Environment variable EVENT_LIST_SUBJECTS_TIMEOUT (in seconds)
// 2. Config file value (passed as parameter, in seconds)
// 3. Default value (10 seconds)
func GetEventListSubjectsTimeout(configValue uint) uint {
	// Priority 1: Environment variable
	timeout := os.Getenv(EnvEventListSubjectsTimeoutKey)
	if timeout != "" {
		result, err := strconv.ParseUint(timeout, 10, 32)
		if err == nil {
			return uint(result)
		}
	}
	// Priority 2: Config file value
	if configValue > 0 {
		return configValue
	}
	// Priority 3: Default value
	return 10
}

// GetEnvEventQueryTimeout is kept for backward compatibility
// It only reads from environment variable, use GetEventQueryTimeout for full support
func GetEnvEventQueryTimeout() uint {
	return GetEventQueryTimeout(0)
}

// GetEnvEventListSubjectsTimeout is kept for backward compatibility
// It only reads from environment variable, use GetEventListSubjectsTimeout for full support
func GetEnvEventListSubjectsTimeout() uint {
	return GetEventListSubjectsTimeout(0)
}
