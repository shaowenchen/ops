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

	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/google/go-cmp/cmp"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsmetrics "github.com/shaowenchen/ops/pkg/metrics"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=tasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=tasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=tasks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Task object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	controllerName := "Task"

	// Record metrics
	defer func() {
		resultStr := "success"
		if err != nil {
			resultStr = "error"
			opsmetrics.RecordReconcileError(controllerName, req.Namespace, "reconcile_error")
		}
		opsmetrics.RecordReconcile(controllerName, req.Namespace, resultStr)
	}()

	// default to reconcile all namespace, if ACTIVE_NAMESPACE is set, only reconcile ACTIVE_NAMESPACE
	actionNs := opsconstants.GetEnvActiveNamespace()
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if opsconstants.GetEnvDebug() {
		logger.SetVerbose("debug").Build()
	}

	obj := &opsv1.Task{}
	err = r.Client.Get(ctx, req.NamespacedName, obj)

	if k8serrors.IsNotFound(err) {
		obj.Namespace = req.Namespace
		obj.Name = req.Name
		// Record Task info metrics as deleted (value=0)
		opsmetrics.RecordTaskInfo(obj.Namespace, obj.Name, "", "", "", 0)
		r.syncResource(logger, ctx, true, obj)
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error.Println(err)
		return ctrl.Result{}, nil
	}

	// Record Task info metrics (static fields only)
	opsmetrics.RecordTaskInfo(obj.Namespace, obj.Name, obj.Spec.Desc, obj.Spec.Host, obj.Spec.RuntimeImage, 1)
	// Record Task status metrics (dynamic fields)
	opsmetrics.RecordTaskStatus(obj.Namespace, obj.Name)

	r.syncResource(logger, ctx, false, obj)

	return ctrl.Result{}, nil
}

func (r *TaskReconciler) syncResource(logger *opslog.Logger, ctx context.Context, isDeleted bool, obj *opsv1.Task) {
	clusterList := &opsv1.ClusterList{}
	err := r.Client.List(ctx, clusterList, &client.ListOptions{})
	if err != nil {
		logger.Error.Println(err, "failed to list clusters")
		return
	}
	if len(clusterList.Items) > 0 {
		logger.Info.Println("sync task " + obj.GetUniqueKey())
	}
	for _, c := range clusterList.Items {
		if !c.IsHealthy() {
			continue
		}
		objs := []opsv1.Task{*obj}
		kc, err := opskube.NewClusterConnection(&c)
		if err != nil {
			logger.Error.Println(err, "failed to create cluster connection")
		}
		err = kc.SyncTasks(isDeleted, objs)
		if err != nil {
			logger.Error.Println(err, "failed to sync tasks")
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// push event
	namespace, err := opsconstants.GetCurrentNamespace()
	if err == nil {
		go opsevent.FactoryController(namespace, opsconstants.Tasks, opsconstants.Status).Publish(context.TODO(), opsevent.EventController{
			Kind:   opsconstants.Tasks,
			Status: opsconstants.Setup,
		})
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Task{}).
		WithEventFilter(
			predicate.Funcs{
				// Allow reconcile for spec and status updates
				UpdateFunc: func(e event.UpdateEvent) bool {
					if _, ok := e.ObjectOld.(*opsv1.Task); !ok {
						return true
					}

					oldObject := e.ObjectOld.(*opsv1.Task).DeepCopy()
					newObject := e.ObjectNew.(*opsv1.Task).DeepCopy()

					oldObjectCmp := &opsv1.Task{}
					newObjectCmp := &opsv1.Task{}

					oldObjectCmp.Spec = oldObject.Spec
					newObjectCmp.Spec = newObject.Spec
					oldObjectCmp.Status = oldObject.Status
					newObjectCmp.Status = newObject.Status

					return !cmp.Equal(oldObjectCmp, newObjectCmp)
				},
			},
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxResourceConcurrentReconciles}).
		Complete(r)
}
