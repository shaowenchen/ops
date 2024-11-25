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
	"github.com/google/go-cmp/cmp"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	// default to reconcile all namespace, if ACTIVE_NAMESPACE is set, only reconcile ACTIVE_NAMESPACE
	actionNs := opsconstants.GetEnvActiveNamespace()
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if opsconstants.GetEnvDebug() {
		logger.SetVerbose("debug").Build()
	}

	obj := &opsv1.Pipeline{}
	err = r.Client.Get(ctx, req.NamespacedName, obj)

	if apierrors.IsNotFound(err) {
		obj.Namespace = req.Namespace
		obj.Name = req.Name
		r.syncResource(logger, ctx, true, obj)
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	// filled variables
	changed := r.filledVariables(logger, ctx, obj)
	if changed {
		logger.Info.Println("filled variables for pipeline", obj.GetUniqueKey())
		r.Update(ctx, obj)
		return ctrl.Result{}, nil
	}

	// sync
	r.syncResource(logger, ctx, false, obj)

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) filledVariables(logger *opslog.Logger, ctx context.Context, obj *opsv1.Pipeline) (changed bool) {
	// get tasks
	taskList := []opsv1.Task{}
	for _, t := range obj.Spec.Tasks {
		task := opsv1.Task{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: obj.Namespace, Name: t.TaskRef}, &task)
		if err != nil {
			logger.Error.Println(err, "failed to get task")
			return false
		}
		taskList = append(taskList, task)
	}
	// merge variables
	for _, t := range taskList {
		if obj.MergeVariables(t.Spec.Variables) {
			changed = true
		}
	}
	// check cluster variables
	findClusterVariable := false
	for _, v := range obj.Spec.Variables {
		if v.Value == opsconstants.ClusterLower {
			findClusterVariable = true
			break
		}
	}
	if !findClusterVariable {
		changed = true
		obj.Spec.Variables[opsconstants.ClusterLower] = opsv1.Variable{
			Desc: opsconstants.ClusterLower,
		}
	}
	// check host variables
	findHostVariable := false
	for _, v := range obj.Spec.Variables {
		if v.Value == opsconstants.HostLower {
			findHostVariable = true
			break
		}
	}
	if !findHostVariable {
		for _, t := range taskList {
			if !opsconstants.IsAnyKubeNode(t.Spec.Host) {
				changed = true
				obj.Spec.Variables[opsconstants.HostLower] = opsv1.Variable{
					Desc:     opsconstants.HostLower,
					Required: true,
				}
				break
			}
		}
	}
	return
}

func (r *PipelineReconciler) syncResource(logger *opslog.Logger, ctx context.Context, isDeleted bool, obj *opsv1.Pipeline) {
	clusterList := &opsv1.ClusterList{}
	err := r.List(ctx, clusterList, &client.ListOptions{})
	if err != nil {
		logger.Error.Println(err, "failed to list clusters")
		return
	}

	if len(clusterList.Items) > 0 {
		logger.Info.Println("sync pipeline " + obj.GetUniqueKey())
	}

	for _, c := range clusterList.Items {
		if !c.IsHealthy() {
			continue
		}
		objs := []opsv1.Pipeline{*obj}
		kc, err := opskube.NewClusterConnection(&c)
		if err != nil {
			logger.Error.Println(err, "failed to create cluster connection")
		}
		err = kc.SyncPipelines(isDeleted, objs)
		if err != nil {
			logger.Error.Println(err, "failed to sync specified pipelines")
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	namespace, err := opsconstants.GetCurrentNamespace()
	if err == nil {
		go opsevent.FactoryController(namespace, opsconstants.Pipelines, opsconstants.EventSetup).Publish(context.TODO(), opsevent.EventController{
			Kind: opsconstants.Pipelines,
		})
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Pipeline{}).
		WithEventFilter(
			predicate.Funcs{
				// drop reconcile for status updates
				UpdateFunc: func(e event.UpdateEvent) bool {
					if _, ok := e.ObjectOld.(*opsv1.Pipeline); !ok {
						return true
					}

					oldObject := e.ObjectOld.(*opsv1.Pipeline).DeepCopy()
					newObject := e.ObjectNew.(*opsv1.Pipeline).DeepCopy()

					oldObjectCmp := &opsv1.Pipeline{}
					newObjectCmp := &opsv1.Pipeline{}

					oldObjectCmp.Spec = oldObject.Spec
					newObjectCmp.Spec = newObject.Spec

					return !cmp.Equal(oldObjectCmp, newObjectCmp)
				},
			},
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxResourceConcurrentReconciles}).
		Complete(r)
}
