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

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	Server string `json:"server,omitempty"`
	Config string `json:"config,omitempty"`
	Token  string `json:"token,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Version         string       `json:"version,omitempty"`
	NodeNumber      int          `json:"nodeNumber,omitempty"`
	LastHeartTime   *metav1.Time `json:"lastHeartTime,omitempty"`
	LastHeartStatus string       `json:"lastHeartstatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Server",type=string,JSONPath=`.spec.server`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="NodeNumber",type=string,JSONPath=`.status.nodeNumber`
// +kubebuilder:printcolumn:name="LastHeartTime",type=string,JSONPath=`.status.lastHeartTime`
// +kubebuilder:printcolumn:name="LastHeartStatus",type=string,JSONPath=`.status.lastHeartstatus`
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

func (c *Cluster) GetSpec() *ClusterSpec {
	return &c.Spec
}

func (c *Cluster) GetStatus() *ClusterStatus {
	return &c.Status
}

func NewCluster(namespace, name, server, config, token string) *Cluster {
	return &Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: ClusterSpec{
			Server: server,
			Config: config,
			Token:  token,
		},
	}
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
