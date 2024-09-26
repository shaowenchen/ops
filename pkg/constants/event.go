package constants

const Source = "https://github.com/shaowenchen/ops"

const DefaultEventBusServer = "http://nats-headless:4222"

const SubjectController = KindOps + "." + KindController
const SubjectHost = KindOps + "." + KindHost
const SubjectCluster = KindOps + "." + KindCluster
const SubjectTask = KindOps + "." + KindTask
const SubjectTaskRun = KindOps + "." + KindTaskRun
const SubjectPipeline = KindOps + "." + KindPipeline
const SubjectPipelineRun = KindOps + "." + KindPipelineRun
