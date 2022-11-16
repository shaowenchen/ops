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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskSpec defines the desired state of Task
type TaskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Schedule  string                `json:"schedule,omitempty"`
	Variables map[string]string     `json:"variables,omitempty"`
	Steps     []Step                `json:"steps,omitempty"`
	Name      string                `json:"name,omitempty"`
	Desc      string                `json:"desc,omitempty"`
	HostRef   string                `json:"hostRef,omitempty"`
	Selector  *metav1.LabelSelector `json:"selector,omitempty"`
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
	LastStatus       string       `json:"lastStatus,omitempty"`
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
	RecentRunOutput  []*RunOutput `json:"recentRunOutput,omitempty"`
}

func (taskStatus *TaskStatus) AddRecentRunOutput(runOutput *RunOutput) {
	if len(taskStatus.RecentRunOutput) > 5 {
		taskStatus.RecentRunOutput = taskStatus.RecentRunOutput[:5]
	}
	taskStatus.RecentRunOutput = append(taskStatus.RecentRunOutput, runOutput)
}

type RunOutput struct {
	RunOutputSteps []*RunOutputStep `json:"runOutputSteps,omitempty"`
}

func (runOutput *RunOutput) AddRunOutput(name, output string) *RunOutput {
	runOutput.RunOutputSteps = append(runOutput.RunOutputSteps, &RunOutputStep{Name: name, Output: output})
	return runOutput
}

type RunOutputStep struct {
	Name   string `json:"name,omitempty"`
	Output string `json:"Output,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Task is the Schema for the tasks API
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec,omitempty"`
	Status TaskStatus `json:"status,omitempty"`
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
