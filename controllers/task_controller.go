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
	"sync"

	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cron "github.com/robfig/cron/v3"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	Client            client.Client
	Scheme            *runtime.Scheme
	crontabMap        map[string]cron.EntryID
	crontabRunningMap map[string]*sync.Mutex
	cron              *cron.Cron
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
	// default to reconcile all namespace, if ACTIVE_NAMESPACE is set, only reconcile ACTIVE_NAMESPACE
	actionNs := os.Getenv("ACTIVE_NAMESPACE")
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()

	if r.crontabMap == nil {
		r.crontabMap = make(map[string]cron.EntryID)
	}

	if r.crontabRunningMap == nil {
		r.crontabRunningMap = make(map[string]*sync.Mutex)
	}
	if r.cron == nil {
		r.cron = cron.New()
		r.cron.Start()
	}

	t := &opsv1.Task{}
	err = r.Client.Get(ctx, req.NamespacedName, t)

	//if delete, stop ticker
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, r.deleteTask(ctx, req.NamespacedName)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	// validate task
	if t.Spec.RuntimeImage == "" {
		t.Spec.RuntimeImage = constants.DefaultRuntimeImage
	}
	// had run once, skip
	if t.Status.RunStatus != opsv1.StatusEmpty && t.GetSpec().Crontab == "" {
		return ctrl.Result{}, nil
	}

	// create task
	r.createTaskrun(logger, ctx, t)
	if err != nil {
		logger.Error.Println(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Task{}).
		WithEventFilter(
			predicate.Funcs{
				// drop reconcile for status updates
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

					return !cmp.Equal(oldObjectCmp, newObjectCmp)
				},
			},
		).
		Complete(r)
}

func (r *TaskReconciler) deleteTask(ctx context.Context, namespacedName types.NamespacedName) error {
	_, ok := r.crontabMap[namespacedName.String()]
	if ok {
		r.cron.Remove(r.crontabMap[namespacedName.String()])
		delete(r.crontabMap, namespacedName.String())
	}
	return nil
}

func (r *TaskReconciler) createTaskrun(logger *opslog.Logger, ctx context.Context, t *opsv1.Task) (err error) {
	_, ok := r.crontabMap[t.GetUniqueKey()]
	if ok {
		logger.Info.Println(fmt.Sprintf("clear ticker for task %s", t.GetUniqueKey()))
		r.cron.Remove(r.crontabMap[t.GetUniqueKey()])
	}
	r.crontabRunningMap[t.GetUniqueKey()] = &sync.Mutex{}
	r.commitStatus(logger, ctx, t, &t.Status, opsv1.StatusRunning)
	if t.GetSpec().Crontab != "" {
		r.crontabMap[t.GetUniqueKey()], err = r.cron.AddFunc(t.GetSpec().Crontab, func() {
			// create taskrun
			if t.Spec.Selector == nil {
				tr := opsv1.NewTaskRun(t)
				r.Client.Create(ctx, &tr)
			} else {
				// find slector hosts and create taskrun
				if t.Spec.TypeRef == opsv1.TaskTypeRefHost {
					hosts := r.getSelectorHosts(logger, ctx, t)
					for _, h := range hosts {
						tr := opsv1.NewTaskRun(t)
						tr.Spec.NameRef = h.Name
						r.Client.Create(ctx, &tr)
					}
				} else if t.Spec.TypeRef == opsv1.TaskTypeRefCluster {
					// todo
				}
			}

		})
		logger.Info.Println(fmt.Sprintf("start ticker for task %s", t.GetUniqueKey()))
		if err != nil {
			return err
		}
	}
	return
}

func (r *TaskReconciler) getSelectorHosts(logger *opslog.Logger, ctx context.Context, t *opsv1.Task) (hosts []opsv1.Host) {
	hostList := &opsv1.HostList{}
	err := r.Client.List(ctx, hostList, client.MatchingLabels(t.Spec.Selector))
	if err != nil {
		logger.Error.Println(err, "failed to list hosts")
		return
	}
	for _, h := range hostList.Items {
		hosts = append(hosts, h)
	}
	return
}

func (r *TaskReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, overrideStatus *opsv1.TaskStatus, status string) (err error) {
	lastT := &opsv1.Task{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: t.Name, Namespace: t.Namespace}, lastT)
	if err != nil {
		logger.Error.Println(err, "failed to get last task")
		return
	}
	if overrideStatus != nil {
		lastT.Status = *overrideStatus
	}
	if status != "" {
		lastT.Status.RunStatus = status
	}
	err = r.Client.Status().Update(ctx, lastT)
	if err != nil {
		logger.Error.Println(err, "update task status error")
	}
	return
}
