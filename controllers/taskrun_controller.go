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

	"os"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opshost "github.com/shaowenchen/ops/pkg/host"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsoption "github.com/shaowenchen/ops/pkg/option"
	opstask "github.com/shaowenchen/ops/pkg/task"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	if tr.Status.RunStatus != opsv1.StatusEmpty {
		// abort running taskrun if restart or modified
		if tr.Status.RunStatus == opsv1.StatusRunning {
			r.commitStatus(logger, ctx, tr, opsv1.StatusAborted)
		}
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
	finishedTaskruns := []opsv1.TaskRun{}
	for _, tr := range trs.Items {
		if tr.Status.RunStatus != opsv1.StatusEmpty && tr.Status.RunStatus != opsv1.StatusRunning {
			finishedTaskruns = append(finishedTaskruns, tr)
		}
	}

	sort.Slice(finishedTaskruns, func(i, j int) bool {
		return finishedTaskruns[i].Status.StartTime.Before(finishedTaskruns[j].Status.StartTime)
	})

	if len(finishedTaskruns) > int(maxHistory) {
		finishedTaskruns = finishedTaskruns[:len(finishedTaskruns)-int(maxHistory)]
		for _, tr := range finishedTaskruns {
			err := r.Client.Delete(ctx, &tr)
			if err != nil {
				logger.Error.Println(err)
			}
		}
	}
	return
}

func (r *TaskRunReconciler) run(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (err error) {
	r.commitStatus(logger, ctx, tr, opsv1.StatusRunning)
	if tr.GetSpec().TypeRef == opsv1.TaskTypeRefHost || tr.GetSpec().TypeRef == "" {
		h := &opsv1.Host{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.GetNamespace(), Name: tr.Spec.NameRef}, h)
		if err != nil {
			logger.Error.Println(err)
			r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
			return
		}
		// if hostname is empty, use localhost
		if len(t.GetSpec().NameRef) > 0 && err != nil {
			logger.Error.Println(err)
			return
		}
		// fill variables
		extraVariables := map[string]string{
			"hostname": h.GetHostname(),
		}
		// filled host
		if h.Spec.SecretRef != "" {
			err = filledHostFromSecret(h, r.Client, h.Spec.SecretRef)
			if err != nil {
				logger.Error.Println("fill host secretRef error", err)
				return
			}
		}
		logger.Info.Println(fmt.Sprintf("run task %s on host %s", t.GetUniqueKey(), t.Spec.NameRef))
		cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
		r.runTaskOnHost(cliLogger, ctx, t, tr, h, extraVariables)
		cliLogger.Flush()
	} else if tr.GetSpec().TypeRef == opsv1.TaskTypeRefCluster {
		c := &opsv1.Cluster{}
		kubeOpt := opsoption.KubeOption{
			NodeName:     tr.GetSpec().NodeName,
			All:          tr.GetSpec().All,
			RuntimeImage: tr.GetSpec().RuntimeImage,
			OpsNamespace: opsconstants.DefaultOpsNamespace,
		}
		if tr.Spec.NameRef != "" {
			err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.GetNamespace(), Name: tr.Spec.NameRef}, c)
			if err != nil {
				logger.Error.Println(err)
				r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
				return
			}
			logger.Info.Println(fmt.Sprintf("run task %s on cluster %s", t.GetUniqueKey(), t.Spec.NameRef))
		}
		cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
		r.runTaskOnKube(cliLogger, ctx, t, tr, c, kubeOpt)
		cliLogger.Flush()
	}
	return
}

func (r *TaskRunReconciler) runTaskOnHost(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, h *opsv1.Host, variables map[string]string) (err error) {
	hc, err := opshost.NewHostConnBase64(h)
	if err != nil {
		r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
		return err
	}
	r.commitStatus(logger, ctx, tr, opsv1.StatusRunning)
	err = opstask.RunTaskOnHost(ctx, logger, t, tr, hc, opsoption.TaskOption{
		Variables: variables,
	})
	if err != nil {
		r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
	} else {
		r.commitStatus(logger, ctx, tr, opsv1.StatusSuccessed)
	}
	return
}

func (r *TaskRunReconciler) runTaskOnKube(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, c *opsv1.Cluster, kubeOpt opsoption.KubeOption) (err error) {
	kc, err := opskube.NewClusterConnection(c)
	if err != nil {
		r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
		logger.Error.Println(err)
		return err
	}
	nodes, err := opskube.GetNodes(ctx, logger, kc.Client, kubeOpt)
	if err != nil || len(nodes) == 0 {
		r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
		return err
	}
	r.commitStatus(logger, ctx, tr, opsv1.StatusRunning)
	for _, node := range nodes {
		opstask.RunTaskOnKube(logger, t, tr, kc, &node,
			opsoption.TaskOption{
				Variables: tr.GetSpec().Variables,
			}, kubeOpt)
	}
	// get taskrun status
	for _, node := range tr.Status.TaskRunNodeStatus {
		if node.RunStatus == opsv1.StatusFailed {
			r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
			return
		}
	}
	r.commitStatus(logger, ctx, tr, opsv1.StatusSuccessed)
	return
}

func (r *TaskRunReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, tr *opsv1.TaskRun, status string) (err error) {
	if status != "" {
		tr.Status.RunStatus = status
	}
	if tr.Status.RunStatus == opsv1.StatusRunning {
		tr.Status.StartTime = &metav1.Time{Time: time.Now()}
	}
	// get taskrun latest version
	latestTr := &opsv1.TaskRun{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.GetNamespace(), Name: tr.GetName()}, latestTr)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	// update taskrun status
	latestTr.Status = tr.Status
	err = r.Client.Status().Update(ctx, latestTr)
	if err != nil {
		logger.Error.Println(err, "update taskrun status error")
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
