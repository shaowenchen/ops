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

// HostSpec defines the desired state of Host
type HostSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Desc           string `json:"desc,omitempty" yaml:"desc,omitempty"`
	Address        string `json:"address,omitempty" yaml:"address,omitempty"`
	Port           int    `json:"port,omitempty" yaml:"port,omitempty"`
	Username       string `json:"username,omitempty" yaml:"username,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	PrivateKey     string `json:"privatekey,omitempty" yaml:"privatekey,omitempty"`
	PrivateKeyPath string `json:"privatekeypath,omitempty" yaml:"privatekeypath,omitempty"`
	TimeOutSeconds int64  `json:"timeoutseconds,omitempty" yaml:"timeoutseconds,omitempty" `
	SecretRef      string `json:"secretref,omitempty" yaml:"secretref,omitempty"`
}

// HostStatus defines the observed state of Host
type HostStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Hostname         string       `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	KernelVersion    string       `json:"kernelVersion,omitempty" yaml:"kernelVersion,omitempty"`
	Distribution     string       `json:"distribution,omitempty" yaml:"distribution,omitempty" `
	Arch             string       `json:"arch,omitempty" yaml:"arch,omitempty" `
	DiskTotal        string       `json:"diskTotal,omitempty" yaml:"diskTotal,omitempty" `
	DiskUsagePercent string       `json:"diskUsagePercent,omitempty" yaml:"diskUsagePercent,omitempty" `
	CPUTotal         string       `json:"cpuTotal,omitempty" yaml:"cpuTotal,omitempty" `
	CPULoad1         string       `json:"cpuLoad1,omitempty" yaml:"cpuLoad1,omitempty"`
	CPUUsagePercent  string       `json:"cpuUsagePercent,omitempty" yaml:"cpuUsagePercent,omitempty"`
	MemTotal         string       `json:"memTotal,omitempty" yaml:"memTotal,omitempty" `
	MemUsagePercent  string       `json:"memUsagePercent,omitempty" yaml:"memUsagePercent,omitempty"`
	HeartStatus      string       `json:"heartStatus,omitempty" yaml:"heartStatus,omitempty"`
	HeartTime        *metav1.Time `json:"heartTime,omitempty" yaml:"heartTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Hostname",type=string,JSONPath=`.status.hostname`
// +kubebuilder:printcolumn:name="Address",type=string,JSONPath=`.spec.address`
// +kubebuilder:printcolumn:name="Distribution",type=string,JSONPath=`.status.distribution`
// +kubebuilder:printcolumn:name="Arch",type=string,JSONPath=`.status.arch`
// +kubebuilder:printcolumn:name="CPU",type=string,JSONPath=`.status.cpuTotal`
// +kubebuilder:printcolumn:name="Mem",type=string,JSONPath=`.status.memTotal`
// +kubebuilder:printcolumn:name="Disk",type=string,JSONPath=`.status.diskTotal`
// +kubebuilder:printcolumn:name="HeartTime",type=date,JSONPath=`.status.heartTime`
// +kubebuilder:printcolumn:name="HeartStatus",type=string,JSONPath=`.status.heartStatus`
type Host struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   HostSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status HostStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (h *Host) GetUniqueKey() string {
	return types.NamespacedName{
		Namespace: h.Namespace,
		Name:      h.Name,
	}.String()
}

func (h *Host) GetSpec() HostSpec {
	return h.Spec
}

func NewHost(namespace, name, address string, port int, username, password, privateKey, privateKeyPath string, timeoutSeconds int64) (h *Host) {
	return &Host{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: HostSpec{
			Address:        address,
			Port:           port,
			Username:       username,
			Password:       password,
			PrivateKey:     privateKey,
			PrivateKeyPath: privateKeyPath,
			TimeOutSeconds: timeoutSeconds,
		},
	}
}

//+kubebuilder:object:root=true

// HostList contains a list of Host
type HostList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []Host `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&Host{}, &HostList{})
}
