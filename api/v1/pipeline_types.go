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

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Crontab                 string    `json:"crontab,omitempty" yaml:"crontab,omitempty"`
	Desc                    string    `json:"desc,omitempty" yaml:"desc,omitempty"`
	Variables               Variables `json:"variables,omitempty" yaml:"variables,omitempty"`
	Tasks                   []TaskRef `json:"tasks" yaml:"tasks"`
	TTlSecondsAfterFinished int       `json:"ttlSecondsAfterFinished,omitempty" yaml:"ttlSecondsAfterFinished,omitempty"`
}

type TaskRef struct {
	Name         string            `json:"name" yaml:"name"`
	TaskRef      string            `json:"taskRef,omitempty" yaml:"taskRef,omitempty"`
	Variables    map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	AllowFailure bool              `json:"allowFailure,omitempty" yaml:"allowFailure,omitempty"`
	RunAlways    bool              `json:"runAlways,omitempty" yaml:"runAlways,omitempty"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Crontab",type=string,JSONPath=`.spec.crontab`
// Pipeline is the Schema for the pipelines API
type Pipeline struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (obj *Pipeline) GetUniqueKey() string {
	return types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Name,
	}.String()
}

func (obj *Pipeline) GetCrontab() string {
	return obj.Spec.Crontab
}

func (obj *Pipeline) GetTTLSecondsAfterFinished() int {
	if obj.Spec.TTlSecondsAfterFinished > 0 {
		return obj.Spec.TTlSecondsAfterFinished
	}
	return DefaultTTLSecondsAfterFinished
}

//+kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Pipeline `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
