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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskSpec defines the desired state of Task
type TaskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Crontab      string                `json:"crontab,omitempty"`
	Variables    map[string]string     `json:"variables,omitempty"`
	Steps        []Step                `json:"steps,omitempty"`
	Name         string                `json:"name,omitempty"`
	Desc         string                `json:"desc,omitempty"`
	TypeRef      string                `json:"typeRef,omitempty"`
	NameRef      string                `json:"nameRef,omitempty"`
	NodeName     string                `json:"nodeName,omitempty"`
	RuntimeImage string                `json:"runtimeImage,omitempty"`
	NodeSelector *metav1.LabelSelector `json:"nodeselector,omitempty"`
	TypeSelector *metav1.LabelSelector `json:"typeSelector,omitempty"`
}

type Step struct {
	When         string `json:"when,omitempty"`
	Name         string `json:"name,omitempty"`
	Script       string `json:"script,omitempty"`
	LocalFile    string `json:"localfile,omitempty"`
	RemoteFile   string `json:"remotefile,omitempty"`
	Direction    string `json:"direction,omitempty"`
	AllowFailure string `json:"allowfailure,omitempty"`
}

// TaskStatus defines the observed state of Task
type TaskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TaskRunStatus map[string]*TaskRunStatus `json:"taskRunStatus,omitempty"`
	LastRunStatus string                    `json:"lastRunStatus,omitempty"`
	LastRunTime   *metav1.Time              `json:"lastRunTime,omitempty"`
}

const LastRunStatusSuccessed = "successed"
const LastRunStatusFailed = "failed"

func GetRunStatus(err error) string {
	if err == nil {
		return LastRunStatusSuccessed
	}
	return LastRunStatusFailed
}

type TaskRunStatus struct {
	TaskRunStep   []*TaskRunStep `json:"taskRunStep,omitempty"`
	LastRunStatus string         `json:"lastRunStatus,omitempty"`
	LastRunTime   *metav1.Time   `json:"lastRunTime,omitempty"`
}

type TaskRunStep struct {
	StepName   string `json:"stepName,omitempty"`
	StepCmd    string `json:"stepCmd,omitempty"`
	StepOutput string `json:"stepOutput,omitempty"`
	StepStatus string `json:"stepStatus,omitempty"`
}

func (t *TaskStatus) NewTaskRun() {
	t.TaskRunStatus = make(map[string]*TaskRunStatus)
}

func (t *TaskStatus) AddOutputStep(nodeName string, stepName, stepCmd, stepOutput, stepStatus string) {
	if t.TaskRunStatus == nil {
		return
	}
	if _, ok := t.TaskRunStatus[nodeName]; !ok {
		t.TaskRunStatus[nodeName] = &TaskRunStatus{}
	}
	t.TaskRunStatus[nodeName].TaskRunStep = append(t.TaskRunStatus[nodeName].TaskRunStep, &TaskRunStep{
		StepName:   stepName,
		StepCmd:    stepCmd,
		StepOutput: stepOutput,
		StepStatus: stepStatus,
	})
	t.TaskRunStatus[nodeName].LastRunTime = &metav1.Time{Time: time.Now()}
	t.TaskRunStatus[nodeName].LastRunStatus = stepStatus
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Crontab",type=string,JSONPath=`.spec.crontab`
// +kubebuilder:printcolumn:name="TypeRef",type=string,JSONPath=`.spec.typeRef`
// +kubebuilder:printcolumn:name="NameRef",type=string,JSONPath=`.spec.nameRef`
// +kubebuilder:printcolumn:name="LastRunTime",type=date,JSONPath=`.status.lastRunTime`
// +kubebuilder:printcolumn:name="LastRunStatus",type=string,JSONPath=`.status.lastRunStatus`
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec,omitempty"`
	Status TaskStatus `json:"status,omitempty"`
}

func (t *Task) GetUniqueKey() string {
	return types.NamespacedName{
		Namespace: t.Namespace,
		Name:      t.Name,
	}.String()
}

func (t *Task) GetSpec() *TaskSpec {
	return &t.Spec
}

//+kubebuilder:object:root=true

// TaskList contains a list of Task
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Task `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
