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

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	crdv1 "github.com/shaowenchen/ops/api/v1"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PipelineRunReconciler reconciles a PipelineRun object
type PipelineRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelineruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PipelineRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	actionNs := os.Getenv("ACTIVE_NAMESPACE")
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if os.Getenv("DEBUG") == "true" {
		logger.SetVerbose("debug").Build()
	}
	pr := &opsv1.PipelineRun{}
	err := r.Client.Get(ctx, req.NamespacedName, pr)

	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	// had run once, skip
	if !(pr.Status.RunStatus == opsv1.StatusEmpty || pr.Status.RunStatus == opsv1.StatusRunning) {
		return ctrl.Result{}, nil
	}
	// get pipeline
	p := &opsv1.Pipeline{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.Namespace, Name: pr.Spec.PipelineRef}, p)
	if err != nil {
		r.commitStatus(logger, ctx, pr, opsv1.StatusFailed, "", "", nil)
		return ctrl.Result{}, err
	}
	// Todo: clear history

	// run
	err = r.run(logger, ctx, p, pr)

	return ctrl.Result{}, err
}

func (r *PipelineRunReconciler) run(logger *opslog.Logger, ctx context.Context, p *opsv1.Pipeline, pr *opsv1.PipelineRun) (err error) {
	runAlways := false
	for _, tRef := range p.Spec.Tasks {
		if runAlways && !tRef.RunAlways {
			continue
		}
		// create taskrun
		t := &opsv1.Task{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.Namespace, Name: tRef.TaskRef}, t)
		if err != nil {
			logger.Error.Println(err)
			runAlways = true
			r.commitStatus(logger, ctx, pr, opsv1.StatusFailed, tRef.Name, tRef.TaskRef, nil)
			continue
		}
		runtimeImage := opsconstants.DefaultRuntimeImage
		if len(os.Getenv("DEFAULT_RUNTIME_IMAGE")) > 0 {
			runtimeImage = os.Getenv("DEFAULT_RUNTIME_IMAGE")
		}
		tr := &opsv1.TaskRun{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    pr.Namespace,
				GenerateName: fmt.Sprintf("%s-%s-", pr.Name, t.Name),
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "crd.chenshaowen.com/v1",
						Kind:       "PipelineRun",
						Name:       pr.Name,
						UID:        pr.UID,
					},
				},
			},
			Spec: opsv1.TaskRunSpec{
				TaskRef:      tRef.TaskRef,
				TypeRef:      mergeValue(t.Spec.TypeRef, pr.Spec.TypeRef),
				NameRef:      mergeValue(t.Spec.NameRef, pr.Spec.NameRef),
				NodeName:     mergeValue(t.Spec.NodeName, pr.Spec.NodeName),
				Variables:    mergeMapValue(t.Spec.Variables, pr.Spec.Variables),
				RuntimeImage: runtimeImage,
			},
		}
		// merge variable to spec
		if tr.Spec.Variables != nil {
			if _, ok := tr.Spec.Variables["typeRef"]; ok {
				tr.Spec.TypeRef = tr.Spec.Variables["typeRef"]
			}
			if _, ok := tr.Spec.Variables["nameRef"]; ok {
				tr.Spec.NameRef = tr.Spec.Variables["nameRef"]
			}
			if _, ok := tr.Spec.Variables["nodeName"]; ok {
				tr.Spec.NodeName = tr.Spec.Variables["nodeName"]
			}
		}
		// pin some variable
		if t.Spec.NodeName == "anymaster" {
			tr.Spec.NodeName = "anymaster"
		}
		err = r.Client.Create(ctx, tr)
		if err != nil {
			logger.Error.Println(err)
			runAlways = true
			r.commitStatus(logger, ctx, pr, opsv1.StatusDataInValid, tRef.Name, tRef.TaskRef, nil)
			continue
		}
		// run task and commit status
		for {
			trRunning := &opsv1.TaskRun{}
			if err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.Namespace, Name: tr.Name}, trRunning); err != nil {
				logger.Error.Println(err)
				break
			}
			r.commitStatus(logger, ctx, pr, opsv1.StatusRunning, tRef.Name, trRunning.Spec.TaskRef, &trRunning.Status)
			if trRunning.Status.RunStatus == opsv1.StatusRunning || trRunning.Status.RunStatus == opsv1.StatusEmpty {
				time.Sleep(time.Second * 3)
			} else {
				runAlways = true
				break
			}
		}
	}
	finallyStatus := opsv1.StatusSuccessed
	for _, status := range pr.Status.PipelineRunStatus {
		if status.TaskRunStatus.RunStatus != opsv1.StatusSuccessed {
			finallyStatus = opsv1.StatusFailed
			break
		}
	}
	r.commitStatus(logger, ctx, pr, finallyStatus, "", "", nil)
	return
}

func mergeValue(value1 string, value2 string) string {
	if value1 == "" {
		return value2
	}
	return value1
}

func mergeMapValue(value1 map[string]string, value2 map[string]string) map[string]string {
	if value2 == nil {
		return value1
	}
	// 如果 value2 中的 key 值为空，并且存在于 value1 中
	for k, v := range value2 {
		if v == "" && value1[k] != "" {
			value2[k] = value1[k]
		}
	}
	return value2
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.PipelineRun{}).
		Complete(r)
}

func (r *PipelineRunReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, pr *opsv1.PipelineRun, prstatus string, taskName, taskRef string, trStatus *opsv1.TaskRunStatus) (err error) {
	// get taskrun latest version
	latestPr := &opsv1.PipelineRun{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.GetNamespace(), Name: pr.GetName()}, latestPr)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	if latestPr.Status.StartTime == nil {
		latestPr.Status.StartTime = &metav1.Time{Time: time.Now()}
	}
	latestPr.Status.RunStatus = prstatus
	// update taskrun task status
	if trStatus != nil {
		latestPr.Status.AddPipelineRunTaskStatus(taskName, taskRef, trStatus)
	}
	err = r.Client.Status().Update(ctx, latestPr)
	if err != nil {
		logger.Error.Println(err, "update taskrun status error")
	}
	return
}
