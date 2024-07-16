/*
Copyright 2022 shaowenchen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskRunSpec defines the desired state of TaskRun
type TaskRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Crontab   string            `json:"crontab,omitempty" yaml:"crontab,omitempty"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Ref       string            `json:"ref,omitempty" yaml:"ref,omitempty"`
}

func (obj *TaskRun) Patch(t *Task) {
	if obj.Spec.Variables == nil {
		obj.Spec.Variables = make(map[string]string)
	}
	for k, v := range t.Spec.Variables {
		if _, ok := obj.Spec.Variables[k]; !ok {
			obj.Spec.Variables[k] = v.GetValue()
			continue
		}
		if v.Value != "" {
			obj.Spec.Variables[k] = v.Value
			continue
		}
	}
}

func (obj *TaskRun) GetUniqueKey() string {
	return types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Name,
	}.String()
}

func NewTaskRun(t *Task) TaskRun {
	tr := TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: t.Name + "-",
			Namespace:    t.Namespace,
		},
		Spec: TaskRunSpec{
			Ref:       t.ObjectMeta.GetName(),
			Variables: t.Spec.Variables.GetVariables(),
		},
	}
	// fill owner ref
	if t.UID != "" {
		tr.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: APIVersion,
				Kind:       TaskKind,
				Name:       t.Name,
				UID:        t.UID,
			},
		}
	}
	return tr
}

func NewTaskRunWithPipelineRun(pr *PipelineRun, t *Task, tRef TaskRef) *TaskRun {
	tr := &TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    pr.Namespace,
			GenerateName: fmt.Sprintf("%s-%s-", pr.Name, tRef.Ref),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: APIVersion,
					Kind:       PipelineRunKind,
					Name:       pr.Name,
					UID:        pr.UID,
				},
			},
		},
		Spec: TaskRunSpec{
			Ref:       t.ObjectMeta.GetName(),
			Variables: t.Spec.Variables.GetVariables(),
		},
	}
	// merge variables
	for k, value := range pr.Spec.Variables {
		tr.Spec.Variables[k] = value
	}

	return tr
}

// TaskRunStatus defines the observed state of TaskRun
type TaskRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TaskRunNodeStatus map[string]*TaskRunNodeStatus `json:"taskrunNodeStatus,omitempty" yaml:"taskrunNodeStatus,omitempty"`
	RunStatus         string                        `json:"runStatus,omitempty" yaml:"runStatus,omitempty"`
	StartTime         *metav1.Time                  `json:"startTime,omitempty" yaml:"startTime,omitempty"`
}

type TaskRunNodeStatus struct {
	NodeName    string         `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
	TaskRunStep []*TaskRunStep `json:"taskRunStep,omitempty" yaml:"taskRunStep,omitempty"`
	RunStatus   string         `json:"runStatus,omitempty" yaml:"runStatus,omitempty"`
	StartTime   *metav1.Time   `json:"startTime,omitempty" yaml:"startTime,omitempty"`
}

type TaskRunStep struct {
	StepName   string `json:"stepName,omitempty" yaml:"stepName,omitempty"`
	StepOutput string `json:"stepOutput,omitempty" yaml:"stepOutput,omitempty"`
	StepStatus string `json:"stepStatus,omitempty" yaml:"stepStatus,omitempty"`
}

func (tr *TaskRunStatus) AddOutputStep(nodeName string, stepName, stepCmd, stepOutput, stepStatus string) {
	if tr.TaskRunNodeStatus == nil {
		tr.TaskRunNodeStatus = make(map[string]*TaskRunNodeStatus)
	}
	if _, ok := tr.TaskRunNodeStatus[nodeName]; !ok {
		tr.TaskRunNodeStatus[nodeName] = &TaskRunNodeStatus{}
	}
	tr.TaskRunNodeStatus[nodeName].TaskRunStep = append(tr.TaskRunNodeStatus[nodeName].TaskRunStep, &TaskRunStep{
		StepName:   stepName,
		StepOutput: stepOutput,
		StepStatus: stepStatus,
	})
	tr.TaskRunNodeStatus[nodeName].StartTime = &metav1.Time{Time: time.Now()}
	tr.TaskRunNodeStatus[nodeName].RunStatus = stepStatus
}

func (tr *TaskRunStatus) ClearNodeStatus() {
	tr.TaskRunNodeStatus = nil
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ref",type=string,JSONPath=`.spec.ref`
// +kubebuilder:printcolumn:name="Crontab",type=string,JSONPath=`.spec.crontab`
// +kubebuilder:printcolumn:name="StartTime",type=date,JSONPath=`.status.startTime`
// +kubebuilder:printcolumn:name="RunStatus",type=string,JSONPath=`.status.runStatus`
// TaskRun is the Schema for the taskruns API
type TaskRun struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   TaskRunSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status TaskRunStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TaskRunList contains a list of TaskRun
type TaskRunList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []TaskRun `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&TaskRun{}, &TaskRunList{})
}
