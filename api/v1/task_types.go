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
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskSpec defines the desired state of Task
type TaskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Desc                    string            `json:"desc,omitempty" yaml:"desc,omitempty"`
	ResType                 string            `json:"resType,omitempty" yaml:"resType,omitempty"`
	ResName                 string            `json:"resName,omitempty" yaml:"resName,omitempty"`
	NodeName                string            `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
	Variables               Variables         `json:"variables,omitempty" yaml:"variables,omitempty"`
	Steps                   []Step            `json:"steps,omitempty" yaml:"steps,omitempty"`
	Selector                map[string]string `json:"selector,omitempty" yaml:"selector,omitempty"`
	All                     bool              `json:"all,omitempty" yaml:"all,omitempty"`
	RuntimeImage            string            `json:"runtimeImage,omitempty" yaml:"runtimeImage,omitempty"`
	TTlSecondsAfterFinished int               `json:"ttlSecondsAfterFinished,omitempty" yaml:"ttlSecondsAfterFinished,omitempty"`
}

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
// +kubebuilder:printcolumn:name="ResType",type=string,JSONPath=`.spec.resType`
// +kubebuilder:printcolumn:name="ResName",type=string,JSONPath=`.spec.resName`
// +kubebuilder:printcolumn:name="NodeName",type=string,JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="All",type=boolean,JSONPath=`.spec.all`
// +kubebuilder:printcolumn:name="Selector",type=string,JSONPath=`.spec.selector`
type Task struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec              TaskSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status            TaskStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// TaskStatus defines the observed state of Task
type TaskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

func (obj *Task) GetTTLSecondsAfterFinished() int {
	if obj.Spec.TTlSecondsAfterFinished > 0 {
		return obj.Spec.TTlSecondsAfterFinished
	}
	return opsconstants.DefaultTTLSecondsAfterFinished
}

func (obj *Task) GetUniqueKey() string {
	return types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Name,
	}.String()
}

func (obj *Task) GetResName(variables map[string]string) string {
	if len(obj.Spec.ResName) > 0 {
		return obj.Spec.ResName
	}
	if _, ok := variables["resName"]; ok {
		return variables["resName"]
	}
	return opsconstants.CurrentRuntime
}

func (obj *Task) GetNodeName(variables map[string]string) string {
	if len(obj.Spec.NodeName) > 0 {
		return obj.Spec.NodeName
	}
	if _, ok := variables["nodeName"]; ok {
		return variables["nodeName"]
	}
	return ""
}

func (obj *Task) IsHostResType() bool {
	return obj.Spec.ResType == opsconstants.ResTypeHost
}

func (obj *Task) IsClusterResType() bool {
	return obj.Spec.ResType == opsconstants.ResTypeCluster || obj.Spec.NodeName == opsconstants.AnyMaster
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
