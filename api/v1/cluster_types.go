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

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	Desc   string `json:"desc,omitempty" yaml:"desc,omitempty" `
	Server string `json:"server,omitempty" yaml:"server,omitempty" `
	Config string `json:"config,omitempty" yaml:"config,omitempty"`
	Token  string `json:"token,omitempty" yaml:"token,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Version          string       `json:"version,omitempty" yaml:"version,omitempty"`
	Node             int          `json:"node,omitempty" yaml:"node,omitempty"`
	Pod              int          `json:"pod,omitempty" yaml:"pod,omitempty"`
	RunningPod       int          `json:"runningPod,omitempty" yaml:"runningPod,omitempty"`
	HeartTime        *metav1.Time `json:"heartTime,omitempty" yaml:"heartTime,omitempty"`
	HeartStatus      string       `json:"heartstatus,omitempty" yaml:"heartstatus,omitempty"`
	CertNotAfterDays int          `json:"certNotAfterDays,omitempty" yaml:"certNotAfterDays,omitempty"`
}

const StatusSuccessed = "successed"
const StatusFailed = "failed"
const StatusRunning = "running"
const StatusInit = "init"
const StatusFiring = "firing"
const StatusEmpty = ""

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Server",type=string,JSONPath=`.spec.server`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.status.node`
// +kubebuilder:printcolumn:name="Running",type=string,JSONPath=`.status.runningPod`
// +kubebuilder:printcolumn:name="TotalPod",type=string,JSONPath=`.status.pod`
// +kubebuilder:printcolumn:name="CertDays",type=string,JSONPath=`.status.certNotAfterDays`
// +kubebuilder:printcolumn:name="HeartTime",type=date,JSONPath=`.status.heartTime`
// +kubebuilder:printcolumn:name="HeartStatus",type=string,JSONPath=`.status.heartstatus`
type Cluster struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (c *Cluster) GetSpec() *ClusterSpec {
	return &c.Spec
}

func (c *Cluster) GetStatus() *ClusterStatus {
	return &c.Status
}

func (c *Cluster) GetUniqueKey() string {
	return types.NamespacedName{
		Namespace: c.Namespace,
		Name:      c.Name,
	}.String()
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
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Cluster `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
