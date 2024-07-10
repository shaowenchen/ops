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
	"math/rand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cron "github.com/robfig/cron/v3"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	Client     client.Client
	Scheme     *runtime.Scheme
	crontabMap map[string]cron.EntryID
	cron       *cron.Cron
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
	if os.Getenv("DEBUG") == "true" {
		logger.SetVerbose("debug").Build()
	}
	if r.crontabMap == nil {
		r.crontabMap = make(map[string]cron.EntryID)
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
	// validate typeRef
	if t.Spec.TypeRef == "" && t.Spec.NodeName == constants.AnyMaster {
		t.Spec.TypeRef = opsv1.TypeRefCluster
	}
	// validate crontab
	if t.GetSpec().Crontab == "" {
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
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxResourceConcurrentReconciles}).
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
	if t.GetSpec().Crontab == "" {
		return nil
	}
	r.crontabMap[t.GetUniqueKey()], err = r.cron.AddFunc(t.GetSpec().Crontab, func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncTaskCronRandomBiasSeconds)) * time.Second)
		taskRunList := opsv1.TaskRunList{}
		err := r.Client.List(ctx, &taskRunList, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				opsv1.LabelCronTaskRunKey: opsv1.LabelCronTaskRunValue,
				opsv1.LabelTaskRefKey:     t.Name,
			}),
		})
		if err != nil {
			logger.Error.Println(err)
			return
		}
		for _, tr := range taskRunList.Items {
			if tr.Status.RunStatus == opsv1.StatusRunning || tr.Status.RunStatus == opsv1.StatusEmpty {
				logger.Info.Println(fmt.Sprintf("skip running taskrun %s", tr.Name))
				return
			}
		}
		tr := opsv1.NewTaskRun(t)
		tr.Labels = map[string]string{
			opsv1.LabelCronTaskRunKey: opsv1.LabelCronTaskRunValue,
			opsv1.LabelTaskRefKey:     t.Name,
		}
		r.Client.Create(ctx, &tr)
	})
	logger.Info.Println(fmt.Sprintf("start ticker for task %s", t.GetUniqueKey()))
	if err != nil {
		return err
	}
	return nil
}
