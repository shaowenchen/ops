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

// HostSpec defines the desired state of Host
type HostSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Address        string `yaml:"address,omitempty" json:"address,omitempty"`
	Port           int    `yaml:"port,omitempty" json:"port,omitempty"`
	Username       string `yaml:"username,omitempty" json:"username,omitempty"`
	Password       string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey     string `yaml:"privatekey,omitempty" json:"privatekey,omitempty"`
	PrivateKeyPath string `yaml:"privatekeypath,omitempty" json:"privatekeypath,omitempty"`
	Timeout        int64  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// HostStatus defines the observed state of Host
type HostStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Heartbeat        bool   `yaml:"heartbeat,omitempty" json:"heartbeat,omitempty"`
	Hostname         string `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	KernelVersion    string `yaml:"kernelversion,omitempty" json:"kernelversion,omitempty"`
	Distribution     string `yaml:"distribution,omitempty" json:"distribution,omitempty"`
	DiskTotal        string `yaml:"disktotal,omitempty" json:"disktotal,omitempty"`
	DiskUsagePercent string `yaml:"diskusagepercent,omitempty" json:"diskusagepercent,omitempty"`
	CPUTotal         string `yaml:"cputotal,omitempty" json:"cputotal,omitempty"`
	CPULoad1         string `yaml:"cpuload1,omitempty" json:"cpuload1,omitempty"`
	CPUUsagePercent  string `yaml:"cpuusagepercent,omitempty" json:"cpuusagepercent,omitempty"`
	MemTotal         string `yaml:"memtotal,omitempty" json:"memtotal,omitempty"`
	MemUsagePercent  string `yaml:"memusagepercent,omitempty" json:"memusagepercent,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="Hostname",type=string,JSONPath=`.status.hostname`
// +kubebuilder:printcolumn:name="Address",type=string,JSONPath=`.spec.address`
// +kubebuilder:printcolumn:name="Distribution",type=string,JSONPath=`.status.distribution`
// +kubebuilder:printcolumn:name="CPU",type=string,JSONPath=`.status.cputotal`
// +kubebuilder:printcolumn:name="Mem",type=string,JSONPath=`.status.memtotal`
// +kubebuilder:printcolumn:name="Disk",type=string,JSONPath=`.status.disktotal`
type Host struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostSpec   `json:"spec,omitempty"`
	Status HostStatus `json:"status,omitempty"`
}

func NewHost(namespace, name, address string, port int, username, password, privatekey, privatekeypath string) (h *Host) {
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
			PrivateKey:     privatekey,
			PrivateKeyPath: privatekeypath,
			Timeout:        10,
		},
	}
}

//+kubebuilder:object:root=true

// HostList contains a list of Host
type HostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Host `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Host{}, &HostList{})
}
