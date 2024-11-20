package event

import (
	"github.com/shaowenchen/ops/pkg/constants"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

func FactoryController(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectController)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryHost(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectHost)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryCluster(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectCluster)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryTask(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectTask)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryTaskRun(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectTaskRun)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryPipeline(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectPipeline)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryPipelineRun(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectPipelineRun)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryWebhook(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectWebhook)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryCheck(sub string) *EventBus {
	subject := opsconstants.GetClusterSubject(constants.GetEnvCluster(), opsconstants.SubjectCheck)
	if sub != "" {
		subject = subject + "." + sub
	}
	return (&EventBus{}).WithSubject(subject)
}

func FactoryWithServer(server string) *EventBus {
	return (&EventBus{}).WithServer(server)
}

func FactoryWithSubject(subject string) *EventBus {
	return (&EventBus{}).WithSubject(subject)
}
