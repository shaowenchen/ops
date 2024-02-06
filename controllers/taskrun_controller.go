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
	"sort"
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opshost "github.com/shaowenchen/ops/pkg/host"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsoption "github.com/shaowenchen/ops/pkg/option"
	opstask "github.com/shaowenchen/ops/pkg/task"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TaskRunReconciler reconciles a TaskRun object
type TaskRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=taskruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=taskruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=taskruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TaskRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TaskRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// default to reconcile all namespace, if ACTIVE_NAMESPACE is set, only reconcile ACTIVE_NAMESPACE
	actionNs := os.Getenv("ACTIVE_NAMESPACE")
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	tr := &opsv1.TaskRun{}
	err := r.Client.Get(ctx, req.NamespacedName, tr)

	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	// validate task
	if tr.Spec.RuntimeImage == "" {
		tr.Spec.RuntimeImage = opsconstants.DefaultRuntimeImage
	}
	// had run once, skip
	if tr.Status.RunStatus != "" {
		return ctrl.Result{}, nil
	}
	// get task
	t := &opsv1.Task{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.Namespace, Name: tr.Spec.TaskRef}, t)
	if err != nil {
		return ctrl.Result{}, err
	}
	// clear history
	go func() {
		r.clearHistory(logger, ctx, t, tr)
	}()
	// run taskrun
	err = r.run(logger, ctx, t, tr)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *TaskRunReconciler) clearHistory(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) {
	trs := &opsv1.TaskRunList{}
	maxHistory := t.GetSpec().TaskRunHistoryLimit
	if maxHistory == 0 {
		maxHistory = opsv1.DefaultMaxTaskrunHistory
	}
	err := r.Client.List(ctx,
		trs,
		client.InNamespace(t.GetNamespace()),
		client.MatchingFields{".spec.taskRef": t.Name})
	if err != nil {
		logger.Error.Println(err)
		return
	}

	sort.Slice(trs.Items, func(i, j int) bool {
		return trs.Items[i].Status.StartTime.Before(trs.Items[j].Status.StartTime)
	})

	if len(trs.Items) > int(maxHistory) {
		trs.Items = trs.Items[:len(trs.Items)-int(maxHistory)]
		for _, tr := range trs.Items {
			err := r.Client.Delete(ctx, &tr)
			if err != nil {
				logger.Error.Println(err)
			}
		}
	}
	return
}

func (r *TaskRunReconciler) run(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (err error) {
	r.commitStatus(logger, ctx, t, tr, opsv1.StatusInit)
	if t.GetSpec().TypeRef == opsv1.TaskTypeRefHost || t.GetSpec().TypeRef == "" {
		h := &opsv1.Host{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: t.GetNamespace(), Name: t.GetSpec().NameRef}, h)
		if err != nil {
			logger.Error.Println(err)
			r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
			return
		}
		// if hostname is empty, use localhost
		if len(t.GetSpec().NameRef) > 0 && err != nil {
			logger.Error.Println(err)
			return
		}
		logger.Info.Println(fmt.Sprintf("run task %s on host %s", t.GetUniqueKey(), t.Spec.NameRef))
		cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
		err = r.runTaskOnHost(cliLogger, ctx, t, tr, h)
		cliLogger.Flush()
		if err != nil {
			logger.Error.Println(err)
			return
		}
	} else if t.GetSpec().TypeRef == opsv1.TaskTypeRefCluster {
		c := &opsv1.Cluster{}
		kubeOpt := opsoption.KubeOption{
			NodeName:     t.GetSpec().NodeName,
			All:          t.GetSpec().All,
			RuntimeImage: t.GetSpec().RuntimeImage,
			OpsNamespace: opsconstants.DefaultOpsNamespace,
		}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: t.GetNamespace(), Name: t.GetSpec().NameRef}, c)
		if err != nil {
			logger.Error.Println(err)
			r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
			return
		}
		logger.Info.Println(fmt.Sprintf("run task %s on cluster %s", t.GetUniqueKey(), t.Spec.NameRef))
		cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
		err = r.runTaskOnKube(cliLogger, ctx, t, tr, c, kubeOpt)
		cliLogger.Flush()
		if err != nil {
			logger.Error.Println(err)
			return
		}
	}
	return
}

func (r *TaskRunReconciler) runTaskOnHost(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, h *opsv1.Host) (err error) {
	hc, err := opshost.NewHostConnBase64(h)
	if err != nil {
		r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
		return err
	}
	r.commitStatus(logger, ctx, t, tr, opsv1.StatusRunning)
	err = opstask.RunTaskOnHost(logger, t, tr, hc, opsoption.TaskOption{})
	if err != nil {
		r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
	} else {
		r.commitStatus(logger, ctx, t, tr, opsv1.StatusSuccessed)
	}
	return
}

func (r *TaskRunReconciler) runTaskOnKube(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, c *opsv1.Cluster, kubeOpt opsoption.KubeOption) (err error) {
	kc, err := opskube.NewClusterConnection(c)
	if err != nil {
		r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
		return err
	}
	nodes, err := opskube.GetNodes(logger, kc.Client, kubeOpt)
	if err != nil || len(nodes) == 0 {
		r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
		return err
	}
	r.commitStatus(logger, ctx, t, tr, opsv1.StatusRunning)
	for _, node := range nodes {
		err = opsutils.MergeError(err, opstask.RunTaskOnKube(logger, t, tr, kc, &node, opsoption.TaskOption{}, kubeOpt))
		if err != nil {
			r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
		} else {
			r.commitStatus(logger, ctx, t, tr, opsv1.StatusSuccessed)
		}
	}
	if err != nil {
		r.commitStatus(logger, ctx, t, tr, opsv1.StatusFailed)
	} else {
		r.commitStatus(logger, ctx, t, tr, opsv1.StatusSuccessed)
	}
	return
}

func (r *TaskRunReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, status string) (err error) {
	if err != nil {
		logger.Error.Println(err, "failed to get last task")
		return
	}
	if status != "" {
		tr.Status.RunStatus = status
		t.Status.RunStatus = status
	}
	if tr.Status.RunStatus == opsv1.StatusRunning {
		tr.Status.StartTime = &metav1.Time{Time: time.Now()}
	}
	if t.Status.RunStatus == opsv1.StatusRunning {
		t.Status.StartTime = &metav1.Time{Time: time.Now()}
	}
	err = r.Client.Status().Update(ctx, tr)
	if err != nil {
		logger.Error.Println(err, "update taskrun status error")
	}
	err = r.Client.Status().Update(ctx, t)
	if err != nil {
		logger.Error.Println(err, "update task status error")
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &opsv1.TaskRun{}, ".spec.taskRef", func(rawObj client.Object) []string {
		tr := rawObj.(*opsv1.TaskRun)
		return []string{tr.Spec.TaskRef}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.TaskRun{}).
		Complete(r)
}
