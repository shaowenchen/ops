package kube

import (
	"context"
	"strings"
	"sync"
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	runtimeClientCache runtimeClient.Client
	runtimeClientMu    sync.RWMutex
)

// isConnectionError checks if the error is a connection-related error that should trigger cache invalidation
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	// Check for Kubernetes API errors that indicate connection issues
	if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) || apierrors.IsUnexpectedServerError(err) {
		return true
	}
	// Check for network errors
	errStr := err.Error()
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "reset by peer") ||
		strings.Contains(errStr, "no such host")
}

// clearRuntimeClientCache clears the cached client (thread-safe)
func clearRuntimeClientCache() {
	runtimeClientMu.Lock()
	runtimeClientCache = nil
	runtimeClientMu.Unlock()
}

// GetRuntimeClient gets a runtime client with retry mechanism and caching
// If connection fails, it will clear cache and retry once
func GetRuntimeClient(kubeconfigPath string) (client runtimeClient.Client, err error) {
	// Try to get cached client first
	runtimeClientMu.RLock()
	cachedClient := runtimeClientCache
	runtimeClientMu.RUnlock()

	// If we have a cached client, verify it's still working with a simple health check
	if cachedClient != nil {
		// Quick health check: try to get a well-known namespace (lightweight operation)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ns := &corev1.Namespace{}
		// Try to get default namespace - this is a very lightweight operation
		err = cachedClient.Get(ctx, runtimeClient.ObjectKey{Name: "default"}, ns)
		if err == nil {
			// Client is healthy, return it
			return cachedClient, nil
		}
		// If it's a NotFound error, the client is working fine (namespace might not exist)
		if apierrors.IsNotFound(err) {
			return cachedClient, nil
		}
		// Connection failed, clear cache and create new client
		if isConnectionError(err) {
			clearRuntimeClientCache()
		}
	}

	// Create new client (with retry on failure)
	maxRetries := 2
	for attempt := 0; attempt < maxRetries; attempt++ {
		client, err = createRuntimeClient(kubeconfigPath)
		if err == nil {
			// Cache the successfully created client
			runtimeClientMu.Lock()
			runtimeClientCache = client
			runtimeClientMu.Unlock()
			return client, nil
		}
		// If it's not a connection error, don't retry
		if !isConnectionError(err) {
			return nil, err
		}
		// Wait before retry (exponential backoff)
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return nil, err
}

// createRuntimeClient creates a new runtime client (internal helper)
func createRuntimeClient(kubeconfigPath string) (client runtimeClient.Client, err error) {
	scheme, err := opsv1.SchemeBuilder.Build()
	if err != nil {
		return
	}
	restConfig, err := opsutils.GetRestConfig(kubeconfigPath)
	if err != nil {
		return
	}

	client, err = runtimeClient.New(restConfig, runtimeClient.Options{Scheme: scheme})
	return
}
