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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opslog "github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/task"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	Client     client.Client
	HostClient client.Client
	Scheme     *runtime.Scheme
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
func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	t := &opsv1.Task{}
	err := r.Client.Get(ctx, req.NamespacedName, t)

	if err != nil {
		log.Info(err.Error())
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// delete task
	if !t.DeletionTimestamp.IsZero() {
		err := r.Client.Delete(ctx, t)
		return ctrl.Result{}, err
	}

	// create task
	// err = r.createTask(ctx, t)
	// if err != nil {
	// 	log.Info(err.Error())
	// }

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Task{}).
		Complete(r)
}

func (r *TaskReconciler) createTask(ctx context.Context, t *opsv1.Task) (err error) {
	hs, err := r.analysisHosts(ctx, t)
	if err != nil {
		return
	}
	if t.Spec.Schedule == "" && t.Status.LastStatus == "" {
		r.runTask(ctx, t, hs)
	} else {
		r.runTaskPeriodic(ctx, t, hs, t.GetSpec().Schedule)
	}
	return
}

func (r *TaskReconciler) analysisHosts(ctx context.Context, t *opsv1.Task) (hs []*opsv1.Host, err error) {
	if t.GetSpec().HostRef != "" {
		h := &opsv1.Host{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: t.GetSpec().HostRef, Namespace: t.Namespace}, h)
		if err != nil {
			return
		} else {
			hs = append(hs, h)
			return
		}
	}

	return
}

func (r *TaskReconciler) runTask(ctx context.Context, t *opsv1.Task, hs []*opsv1.Host) (err error) {
	logger, _ := opslog.NewCliLogger(true, false)
	for _, h := range hs {
		task.RunTaskOnHost(logger, t, h, task.TaskOption{})
		r.Client.Status().Update(ctx, t)
	}
	return
}

func (r *TaskReconciler) runTaskPeriodic(ctx context.Context, t *opsv1.Task, hs []*opsv1.Host, schedule string) (err error) {

	ticker := time.NewTicker(time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				logger, _ := opslog.NewCliLogger(true, false)
				for _, h := range hs {
					task.RunTaskOnHost(logger, t, h, task.TaskOption{})
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return
}
