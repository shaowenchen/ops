package event

import (
	"github.com/nats-io/nats.go"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	"strings"
)

// for controller
var server = opsconstants.GetEnvEventAddress()
var cluster = opsconstants.GetEnvEventCluster()

func FactoryController(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectController)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func FactoryCluster(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectCluster)
	subs = append(subs, namespace)
	subject = subject + "." + strings.Join(subs, ".")
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func FactoryHost(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectHost)
	subs = append(subs, namespace)
	subject = subject + "." + strings.Join(subs, ".")
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func FactoryTask(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectTask)
	subs = append(subs, namespace)
	subject = subject + "." + strings.Join(subs, ".")
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func FactoryTaskRun(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectTaskRun)
	subs = append(subs, namespace)
	subject = subject + "." + strings.Join(subs, ".")
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func FactoryPipeline(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectPipeline)
	subs = append(subs, namespace)
	subject = subject + "." + strings.Join(subs, ".")
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func FactoryPipelineRun(namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectPipelineRun)
	subs = append(subs, namespace)
	subject = subject + "." + strings.Join(subs, ".")
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

// for server
func FactoryWebhook(server, cluster, namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectWebhook)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func Factory(server, cluster, namespace string, subs ...string) *EventBus {
	subject := opsconstants.GetClusterSubject(cluster, namespace, opsconstants.SubjectWebhook)
	if len(subs) > 0 {
		subject = subject + "." + strings.Join(subs, ".")
	}
	return (&EventBus{}).WithServer(server).WithSubject(subject)
}

func FactoryJetStreamClient(server, cluster string) (*nats.JetStreamContext, error) {
	nc, err := nats.Connect(server)
	if err != nil {
		return nil, err
	}
	js, _ := nc.JetStream()
	if err == nil {
		return &js, nil
	}
	return nil, err
}
