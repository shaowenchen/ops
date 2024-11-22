package event

import (
	"github.com/nats-io/nats.go"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	"strings"
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

// for endpoint
func FactoryWebhook(endpoint, cluster, namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectWebhook)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

func Factory(endpoint, cluster, namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectWebhook)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithEndpoint(endpoint).WithSubject(subject)
}

var jetCache = make(map[string]*nats.JetStreamContext)

func FactoryJetStreamClient(endpoint, cluster string) (*nats.JetStreamContext, error) {
	if _, ok := jetCache[cluster]; !ok {
		nc, err := nats.Connect(endpoint)
		if err != nil {
			return nil, err
		}
		js, _ := nc.JetStream()
		jetCache[cluster] = &js
	}
	return jetCache[cluster], nil
}
