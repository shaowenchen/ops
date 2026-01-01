package event

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

// for controller
var endpoint = opsconstants.GetEnvEventEndpoint()
var cluster = opsconstants.GetEnvEventCluster()

func FactoryController(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectController)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func FactoryCluster(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectCluster)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func FactoryHost(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectHost)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func FactoryTask(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectTask)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func FactoryTaskRun(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectTaskRun)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func FactoryPipeline(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectPipeline)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func FactoryPipelineRun(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectPipelineRun)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func FactoryKube(namespace string, subs ...string) *EventBus {
	subject := ""
	if namespace == "" {
		subject = fmt.Sprintf(opsconstants.SubjectClusterPrefix, cluster)
	} else {
		subject = opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectPrefix)
	}
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

// for endpoint
func FactoryWebhook(endpoint, cluster, namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectWebhook)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func Factory(endpoint, cluster, namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectPrefix)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

var (
	jetStreamConn *nats.Conn
	jetStreamJS   nats.JetStreamContext
	jetStreamMu   sync.RWMutex
)

func FactoryJetStreamClient(endpoint, cluster string) (nats.JetStreamContext, error) {
	jetStreamMu.RLock()
	// Check if cached connection is still valid
	if jetStreamConn != nil && !jetStreamConn.IsClosed() && jetStreamJS != nil {
		js := jetStreamJS
		jetStreamMu.RUnlock()
		return js, nil
	}
	jetStreamMu.RUnlock()

	// Create new connection
	nc, err := nats.Connect(endpoint)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close() // Close connection if JetStream creation fails
		return nil, err
	}

	// Cache the connection and JetStream context
	jetStreamMu.Lock()
	// Close old connection if exists
	if jetStreamConn != nil && !jetStreamConn.IsClosed() {
		jetStreamConn.Close()
	}
	jetStreamConn = nc
	jetStreamJS = js
	jetStreamMu.Unlock()

	return js, nil
}
