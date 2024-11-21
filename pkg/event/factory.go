package event

import (
	"github.com/nats-io/nats.go"
	"github.com/shaowenchen/ops/pkg/constants"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	"os"
)

func FactoryController() *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectController)
	return (&EventBus{}).WithSubject(subject)
}

func FactoryHost() *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectHost)
	return (&EventBus{}).WithSubject(subject)
}

func FactoryCluster() *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectCluster)
	return (&EventBus{}).WithSubject(subject)
}

func FactoryTask() *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectTask)
	return (&EventBus{}).WithSubject(subject)
}

func FactoryTaskRun() *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectTaskRun)
	return (&EventBus{}).WithSubject(subject)
}

func FactoryPipeline() *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectPipeline)
	return (&EventBus{}).WithSubject(subject)
}

func FactoryPipelineRun() *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectPipelineRun)
	return (&EventBus{}).WithSubject(subject)
}

func FactoryCheck(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectCheck)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryWebhook(sub string) *EventBus {
	subject := opsconstants.SubjectWebhook
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryJetStreamClient(server string) (*nats.JetStreamContext, error) {
	if server == "" {
		if os.Getenv("EVENTBUS_ADDRESS") != "" {
			server = os.Getenv("EVENTBUS_ADDRESS")
		} else {
			server = opsconstants.DefaultEventBusServer
		}
	}
	client, err := CurrentEventBusClient.GetClient(server, "ops.>")
	if err == nil {
		return client.JetStream, nil
	}
	return nil, err
}

func FactoryWithServer(server string) *EventBus {
	return (&EventBus{}).WithServer(server)
}
