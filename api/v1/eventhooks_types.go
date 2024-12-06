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

// EventHooksSpec defines the desired state of EventHooks
type EventHooksSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of EventHooks. Edit eventhooks_types.go to remove/update
	Type       string            `json:"type,omitempty"`
	Subject    string            `json:"subject,omitempty"`
	URL        string            `json:"url,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
	Keywords   []string          `json:"keywords,omitempty"`
	Additional string            `json:"additional,omitempty"`
}

// EventHooksStatus defines the observed state of EventHooks
type EventHooksStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Subject",type=string,JSONPath=`.spec.subject`
// +kubebuilder:printcolumn:name="Desc",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// EventHooks is the Schema for the eventhooks API
type EventHooks struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventHooksSpec   `json:"spec,omitempty"`
	Status EventHooksStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EventHooksList contains a list of EventHooks
type EventHooksList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventHooks `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EventHooks{}, &EventHooksList{})
}
