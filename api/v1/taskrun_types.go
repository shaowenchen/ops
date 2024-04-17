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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultMaxTaskrunHistory = 10
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskRunSpec defines the desired state of TaskRun
type TaskRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TaskRef      string            `json:"taskRef,omitempty" yaml:"taskRef,omitempty"`
	Variables    map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	TypeRef      string            `json:"typeRef,omitempty" yaml:"typeRef,omitempty"`
	NameRef      string            `json:"nameRef,omitempty" yaml:"nameRef,omitempty"`
	NodeName     string            `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
	All          bool              `json:"all,omitempty" yaml:"all,omitempty"`
	RuntimeImage string            `json:"runtimeImage,omitempty" yaml:"runtimeImage,omitempty"`
}

func (tr *TaskRun) GetSpec() *TaskRunSpec {
	return &tr.Spec
}

func NewTaskRun(t *Task) TaskRun {
	return TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: t.Name + "-",
			Namespace:    t.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: t.APIVersion,
					Kind:       t.Kind,
					Name:       t.Name,
					UID:        t.UID,
				},
			},
		},
		Spec: TaskRunSpec{
			TaskRef:      t.GetObjectMeta().GetName(),
			Variables:    t.Spec.Variables,
			TypeRef:      t.Spec.TypeRef,
			NameRef:      t.Spec.NameRef,
			NodeName:     t.Spec.NodeName,
			All:          t.Spec.All,
			RuntimeImage: t.Spec.RuntimeImage,
		},
	}
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
	StepCmd    string `json:"stepCmd,omitempty" yaml:"stepCmd,omitempty"`
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
		StepCmd:    stepCmd,
		StepOutput: stepOutput,
		StepStatus: stepStatus,
	})
	tr.TaskRunNodeStatus[nodeName].StartTime = &metav1.Time{Time: time.Now()}
	tr.TaskRunNodeStatus[nodeName].RunStatus = stepStatus
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Task",type=string,JSONPath=`.spec.taskRef`
// +kubebuilder:printcolumn:name="TypeRef",type=string,JSONPath=`.spec.typeRef`
// +kubebuilder:printcolumn:name="NameRef",type=string,JSONPath=`.spec.nameRef`
// +kubebuilder:printcolumn:name="NodeName",type=string,JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="All",type=boolean,JSONPath=`.spec.all`
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
