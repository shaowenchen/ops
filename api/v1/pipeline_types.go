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

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Desc            string               `json:"desc,omitempty" yaml:"desc,omitempty"`
	Variables       map[string]Variables `json:"variables,omitempty" yaml:"variables,omitempty"`
	Tasks           []TaskRef            `json:"tasks" yaml:"tasks"`
	RunHistoryLimit int                  `json:"runHistoryLimit,omitempty" yaml:"runHistoryLimit,omitempty"`
}

type Variables struct {
	Default  string   `json:"default,omitempty" yaml:"default,omitempty"`
	Value    string   `json:"value,omitempty" yaml:"value,omitempty"`
	Desc     string   `json:"desc,omitempty" yaml:"desc,omitempty"`
	Regex    string   `json:"regex,omitempty" yaml:"regex,omitempty"`
	Required bool     `json:"required,omitempty" yaml:"required,omitempty"`
	Enum     []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Examples []string `json:"examples,omitempty" yaml:"examples,omitempty"`
}

type TaskRef struct {
	Name         string            `json:"name"`
	TaskRef      string            `json:"taskRef,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
	AllowFailure bool              `json:"allowFailure,omitempty"`
	RunAlways    bool              `json:"runAlways,omitempty"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Pipeline is the Schema for the pipelines API
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
