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
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cron "github.com/robfig/cron/v3"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/task"
	"github.com/shaowenchen/ops/pkg/utils"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	Client            client.Client
	Scheme            *runtime.Scheme
	crontabMap        map[string]cron.EntryID
	crontabRunningMap map[string]*sync.Mutex
	cron              *cron.Cron
	updateStatusMap   map[string]*sync.Mutex
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

	if r.updateStatusMap == nil {
		r.updateStatusMap = make(map[string]*sync.Mutex)
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
	// had run
	if t.Status.RunStatus != "" && t.GetSpec().Crontab == "" {
		return ctrl.Result{}, nil
	}

	// create task
	r.createTask(logger, ctx, t)
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

func (r *TaskReconciler) createTask(logger *opslog.Logger, ctx context.Context, t *opsv1.Task) (err error) {
	_, ok := r.crontabMap[t.GetUniqueKey()]
	if ok {
		logger.Info.Println(fmt.Sprintf("clear ticker for task %s", t.GetUniqueKey()))
		r.cron.Remove(r.crontabMap[t.GetUniqueKey()])
	}
	r.crontabRunningMap[t.GetUniqueKey()] = &sync.Mutex{}
	r.updateStatusMap[t.GetUniqueKey()] = &sync.Mutex{}
	t.Status.NewTaskRun()
	r.commitStatus(logger, ctx, t, &t.Status, opsv1.StatusInit)
	if t.GetSpec().TypeRef == "host" || t.GetSpec().TypeRef == "" {
		hostCmd := func() {
			h := &opsv1.Host{}
			err := r.Client.Get(ctx, types.NamespacedName{Namespace: t.GetNamespace(), Name: t.GetSpec().NameRef}, h)
			// if hostname is empty, use localhost
			if len(t.GetSpec().NameRef) > 0 && err != nil {
				fmt.Println(err.Error())
				return
			}
			logger.Info.Println(fmt.Sprintf("run task %s on host %s", t.GetUniqueKey(), t.Spec.NameRef))
			cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
			err = r.runTaskOnHost(cliLogger, ctx, t, h)
			cliLogger.Flush()
			if err != nil {
				logger.Error.Println(err)
			}
		}
		if t.GetSpec().Crontab != "" {
			r.crontabMap[t.GetUniqueKey()], err = r.cron.AddFunc(t.GetSpec().Crontab, hostCmd)
			logger.Info.Println(fmt.Sprintf("start ticker for task %s", t.GetUniqueKey()))
			if err != nil {
				return err
			}
		} else {
			hostCmd()
		}
	} else if t.GetSpec().TypeRef == "cluster" {
		clusterCmd := func() {
			c := &opsv1.Cluster{}
			kubeOpt := option.KubeOption{
				NodeName:     t.GetSpec().NameRef,
				All:          t.GetSpec().All,
				RuntimeImage: t.GetSpec().RuntimeImage,
			}
			err := r.Client.Get(ctx, types.NamespacedName{Namespace: t.GetNamespace(), Name: t.GetSpec().NameRef}, c)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			logger.Info.Println(fmt.Sprintf("run task %s on cluster %s", t.GetUniqueKey(), t.Spec.NameRef))
			cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
			err = r.runTaskOnKube(cliLogger, ctx, t, c, kubeOpt)
			cliLogger.Flush()
			if err != nil {
				logger.Error.Println(err)
			}
		}
		if t.GetSpec().Crontab != "" {
			r.crontabMap[t.GetUniqueKey()], err = r.cron.AddFunc(t.GetSpec().Crontab, clusterCmd)
			logger.Info.Println(fmt.Sprintf("start ticker for task %s", t.GetUniqueKey()))
			if err != nil {
				return err
			}
		} else {
			clusterCmd()
		}
	}
	return
}

func (r *TaskReconciler) runTaskOnHost(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, h *opsv1.Host) (err error) {
	lock := r.crontabRunningMap[t.GetUniqueKey()]
	getLock := lock.TryLock()
	if !getLock {
		logger.Info.Println(fmt.Sprintf("skiped，task %s is running", t.GetUniqueKey()))
		lock.Unlock()
		return
	}
	hc, err := host.NewHostConnBase64(h)
	if err != nil {
		r.commitStatus(logger, ctx, t, nil, opsv1.StatusFailed)
		lock.Unlock()
		return err
	}
	t.Status.NewTaskRun()
	r.commitStatus(logger, ctx, t, &t.Status, opsv1.StatusRunning)
	err = task.RunTaskOnHost(logger, t, hc, option.TaskOption{})
	if err != nil {
		r.commitStatus(logger, ctx, t, nil, opsv1.StatusFailed)
		lock.Unlock()
	} else {
		r.commitStatus(logger, ctx, t, nil, opsv1.StatusSuccessed)
	}
	lock.Unlock()
	return
}

func (r *TaskReconciler) runTaskOnKube(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, c *opsv1.Cluster, kubeOpt option.KubeOption) (err error) {
	lock := r.crontabRunningMap[t.GetUniqueKey()]
	getLock := lock.TryLock()
	if !getLock {
		logger.Info.Println(fmt.Sprintf("skiped，task %s is running", t.GetUniqueKey()))
		return
	}
	kc, err := kube.NewClusterConnection(c)
	if err != nil {
		r.commitStatus(logger, ctx, t, nil, opsv1.StatusFailed)
		lock.Unlock()
		return err
	}
	nodes, err := kube.GetNodes(logger, kc.Client, kubeOpt)
	if err != nil || len(nodes) == 0 {
		r.commitStatus(logger, ctx, t, nil, opsv1.StatusFailed)
		lock.Unlock()
		return err
	}
	t.Status.NewTaskRun()
	r.commitStatus(logger, ctx, t, &t.Status, opsv1.StatusRunning)
	for _, node := range nodes {
		logger.Info.Println(utils.FilledInMiddle(node.Name))
		err = utils.MergeError(err, task.RunTaskOnKube(logger, t, kc, &node, option.TaskOption{}, kubeOpt))
		if err != nil {
			r.commitStatus(logger, ctx, t, &t.Status, opsv1.StatusFailed)
		} else {
			r.commitStatus(logger, ctx, t, &t.Status, opsv1.StatusSuccessed)
		}
	}
	if err != nil {
		r.commitStatus(logger, ctx, t, nil, opsv1.StatusFailed)
	} else {
		r.commitStatus(logger, ctx, t, nil, opsv1.StatusSuccessed)
	}
	lock.Unlock()
	return
}

func (r *TaskReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, overrideStatus *opsv1.TaskStatus, status string) (err error) {
	lock := r.updateStatusMap[t.GetUniqueKey()]
	lock.Lock()
	lastT := &opsv1.Task{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: t.Name, Namespace: t.Namespace}, lastT)
	if err != nil {
		logger.Error.Println(err, "failed to get last task")
		lock.Unlock()
		return
	}
	if overrideStatus != nil {
		lastT.Status = *overrideStatus
	}
	if status != "" {
		lastT.Status.RunStatus = status
	}
	if lastT.Status.RunStatus == opsv1.StatusRunning {
		lastT.Status.StartTime = &metav1.Time{Time: time.Now()}
	}
	// err = r.Client.Status().Update(ctx, lastT)
	err = r.Client.Update(ctx, lastT)
	if err != nil {
		logger.Error.Println(err, "update task status error")
	}
	lock.Unlock()
	return
}
