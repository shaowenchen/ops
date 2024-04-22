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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskSpec defines the desired state of Task
type TaskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Crontab             string            `json:"crontab,omitempty" yaml:"crontab,omitempty"`
	Variables           map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Steps               []Step            `json:"steps,omitempty" yaml:"steps,omitempty"`
	Name                string            `json:"name,omitempty" yaml:"name,omitempty"`
	Desc                string            `json:"desc,omitempty" yaml:"desc,omitempty"`
	TypeRef             string            `json:"typeRef,omitempty" yaml:"typeRef,omitempty"`
	Selector            map[string]string `json:"selector,omitempty" yaml:"selector,omitempty"`
	NameRef             string            `json:"nameRef,omitempty" yaml:"nameRef,omitempty"`
	NodeName            string            `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
	All                 bool              `json:"all,omitempty" yaml:"all,omitempty"`
	RuntimeImage        string            `json:"runtimeImage,omitempty" yaml:"runtimeImage,omitempty"`
	TaskRunHistoryLimit int               `json:"taskRunHistoryLimit,omitempty" yaml:"taskRunHistoryLimit,omitempty"`
}

const TaskTypeRefHost = "host"
const TaskTypeRefCluster = "cluster"

type Step struct {
	When         string `json:"when,omitempty" yaml:"when,omitempty"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	Content      string `json:"content,omitempty" yaml:"content,omitempty"`
	LocalFile    string `json:"localfile,omitempty" yaml:"localfile,omitempty"`
	RemoteFile   string `json:"remotefile,omitempty" yaml:"remotefile,omitempty"`
	Direction    string `json:"direction,omitempty" yaml:"direction,omitempty"`
	AllowFailure string `json:"allowfailure,omitempty" yaml:"allowfailure,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Crontab",type=string,JSONPath=`.spec.crontab`
// +kubebuilder:printcolumn:name="TypeRef",type=string,JSONPath=`.spec.typeRef`
// +kubebuilder:printcolumn:name="NameRef",type=string,JSONPath=`.spec.nameRef`
// +kubebuilder:printcolumn:name="NodeName",type=string,JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="All",type=boolean,JSONPath=`.spec.all`
// +kubebuilder:printcolumn:name="Selector",type=string,JSONPath=`.spec.selector`
type Task struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec TaskSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
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
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Task `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
