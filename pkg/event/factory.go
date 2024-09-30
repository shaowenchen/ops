package event

import (
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
)

func FactoryController() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectController)
}

func FactoryHost() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectHost)
}

func FactoryCluster() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectCluster)
}

func FactoryTask() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectTask)
}

func FactoryTaskRun() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectTaskRun)
}

func FactoryPipeline() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectPipeline)
}

func FactoryPipelineRun() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectPipelineRun)
}

func FactoryWebhook() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectWebhook)
}

func FactoryCheck() *EventBus {
	return (&EventBus{}).WithSubject(opsconstants.SubjectCheck)
}

func FactoryWithServer(server string) *EventBus {
	return (&EventBus{}).WithServer(server)
}

func FactoryWithSubject(subject string) *EventBus {
	return (&EventBus{}).WithSubject(subject)
}
