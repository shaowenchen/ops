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
	"fmt"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PipelineRunSpec defines the desired state of PipelineRun
type PipelineRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	Desc        string            `json:"desc,omitempty" yaml:"desc,omitempty"`
	Crontab     string            `json:"crontab,omitempty" yaml:"crontab,omitempty"`
	Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	PipelineRef string            `json:"pipelineRef,omitempty" yaml:"pipelineRef,omitempty"`
}

// PipelineRunStatus defines the observed state of PipelineRun
type PipelineRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PipelineRunStatus []PipelineRunTaskStatus `json:"pipelineRunStatus,omitempty" yaml:"pipelineRunStatus,omitempty"`
	RunStatus         string                  `json:"runStatus,omitempty" yaml:"runStatus,omitempty"`
	StartTime         *metav1.Time            `json:"startTime,omitempty" yaml:"startTime,omitempty"`
}

func (pr *PipelineRunStatus) AddPipelineRunTaskStatus(taskName string, taskRef string, taskRunStatus *TaskRunStatus) {
	if taskName == "" || taskRef == "" || taskRunStatus == nil {
		return
	}
	found := false
	for i, task := range pr.PipelineRunStatus {
		if task.TaskName == taskName && task.TaskRef == taskRef {
			found = true
			pr.PipelineRunStatus[i].TaskRunStatus = taskRunStatus
		}
	}
	if !found {
		pr.PipelineRunStatus = append(pr.PipelineRunStatus, PipelineRunTaskStatus{
			TaskName:      taskName,
			TaskRef:       taskRef,
			TaskRunStatus: taskRunStatus,
		})
	}
	return
}

type PipelineRunTaskStatus struct {
	TaskName      string         `json:"name,omitempty" yaml:"name,omitempty"`
	TaskRef       string         `json:"taskRef,omitempty" yaml:"taskRef,omitempty"`
	TaskRunStatus *TaskRunStatus `json:"taskRunStatus,omitempty" yaml:"taskRunStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PipelineRef",type=string,JSONPath=`.spec.pipelineRef`
// +kubebuilder:printcolumn:name="Crontab",type=string,JSONPath=`.spec.crontab`
// +kubebuilder:printcolumn:name="Desc",type=string,JSONPath=`.spec.desc`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.runStatus`
// +kubebuilder:printcolumn:name="StartTime",type=date,JSONPath=`.status.startTime`
type PipelineRun struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   PipelineRunSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status PipelineRunStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

func (obj *PipelineRun) GetUniqueKey() string {
	return types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Name,
	}.String()
}

func (obj *PipelineRun) GetCluster() string {
	if obj.Spec.Variables == nil {
		return ""
	}
	return obj.Spec.Variables[opsconstants.ClusterLower]
}

func (obj *PipelineRun) SetCurrentCluster() {
	if obj.Spec.Variables == nil {
		return
	}
	obj.Spec.Variables[opsconstants.ClusterLower] = ""
}

func (obj *PipelineRun) CopyWithOutVersion() *PipelineRun {
	return &PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: obj.GetObjectMeta().GetGenerateName(),
			Name:         obj.GetObjectMeta().GetName(),
			Namespace:    obj.GetObjectMeta().GetNamespace(),
			Labels:       obj.GetObjectMeta().GetLabels(),
			Annotations:  obj.GetObjectMeta().GetAnnotations(),
		},
		Spec:   obj.Spec,
		Status: obj.Status,
	}
}

func (obj *PipelineRun) SetEnv() *PipelineRun {
	if obj.Spec.Variables == nil {
		obj.Spec.Variables = make(map[string]string)
	}
	obj.Spec.Variables["TIME"] = fmt.Sprintf("%d", time.Now().UnixMicro())
	return obj
}

func NewPipelineRun(p *Pipeline) *PipelineRun {
	if p == nil {
		return &PipelineRun{}
	}
	pr := PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: p.Name + "-",
			Namespace:    p.Namespace,
		},
		Spec: PipelineRunSpec{
			PipelineRef: p.Name,
			Variables:   make(map[string]string),
		},
	}
	for k, v := range p.Spec.Variables {
		pr.Spec.Variables[k] = v.GetValue()
	}
	// fill owner ref
	if p.UID != "" {
		pr.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: opsconstants.APIVersion,
				Kind:       opsconstants.Pipeline,
				Name:       p.Name,
				UID:        p.UID,
			},
		}
	}
	// validate
	return &pr
}

//+kubebuilder:object:root=true

// PipelineRunList contains a list of PipelineRun
type PipelineRunList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Items           []PipelineRun `json:"items" yaml:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineRun{}, &PipelineRunList{})
}
