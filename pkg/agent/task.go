package agent

import (
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
)

func NewLLMTask(t *opsv1.Task) LLMTask {
	return LLMTask{
		Desc:      t.Spec.Desc,
		Namespace: t.ObjectMeta.Namespace,
		Name:      t.ObjectMeta.Name,
		NodeName:  t.Spec.NodeName,
		Variables: t.Spec.Variables,
	}
}

type LLMTask struct {
	Desc         string                                                                      `json:"desc"`
	Namespace    string                                                                      `json:"namespace"`
	Name         string                                                                      `json:"name"`
	TypeRef      string                                                                      `json:"typeRef"`
	NameRef      string                                                                      `json:"nameRef"`
	NodeName     string                                                                      `json:"nodeName"`
	Variables    map[string]string                                                           `json:"variables"`
	RuntimeImage string                                                                      `json:"runtimeImage"`
	CallFunction func(*log.Logger, *LLMPipelineRunsManager, *LLMPipelineRun) (string, error) `json:"-"`
}

func (m *LLMTask) BuildTaskRun() *LLMTaskRun {
	return &LLMTaskRun{
		TypeRef:   m.TypeRef,
		Namespace: m.Namespace,
		TaskRef:   m.Name,
		NameRef:   m.NameRef,
		NodeName:  m.NodeName,
		Variables: m.Variables,
	}
}
