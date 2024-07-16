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
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	cron "github.com/robfig/cron/v3"
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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TaskRunReconciler reconciles a TaskRun object
type TaskRunReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	crontabMap map[string]cron.EntryID
	cron       *cron.Cron
	clearCron  *cron.Cron
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
	// start clear cron
	r.registerClearCron()
	// only reconcile active namespace
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

	// get taskrun
	tr := &opsv1.TaskRun{}
	err := r.Client.Get(ctx, req.NamespacedName, tr)
	if apierrors.IsNotFound(err) {
		r.deleteCronTab(logger, ctx, req.NamespacedName)
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	// get task
	t := &opsv1.Task{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.Namespace, Name: tr.Spec.TaskRef}, t)
	if err != nil {
		r.commitStatus(logger, ctx, tr, opsv1.StatusDataInValid)
		return ctrl.Result{}, err
	}
	// add crontab
	r.addCronTab(logger, ctx, tr)
	// check run status
	if tr.Status.RunStatus != opsv1.StatusEmpty {
		// abort running taskrun if restart or modified
		if tr.Status.RunStatus == opsv1.StatusRunning {
			r.commitStatus(logger, ctx, tr, opsv1.StatusAborted)
		}
		return ctrl.Result{}, nil
	}

	err = r.run(logger, ctx, t, tr)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *TaskRunReconciler) deleteCronTab(logger *opslog.Logger, ctx context.Context, namespacedName types.NamespacedName) error {
	_, ok := r.crontabMap[namespacedName.String()]
	if ok {
		r.cron.Remove(r.crontabMap[namespacedName.String()])
		delete(r.crontabMap, namespacedName.String())
		logger.Info.Println(fmt.Sprintf("clear ticker for taskrun %s", namespacedName.String()))
	}
	return nil
}

func (r *TaskRunReconciler) addCronTab(logger *opslog.Logger, ctx context.Context, objRun *opsv1.TaskRun) {
	if objRun.Spec.Crontab == "" {
		return
	}
	_, ok := r.crontabMap[objRun.GetUniqueKey()]
	if ok {
		return
	}
	id, err := r.cron.AddFunc(objRun.Spec.Crontab, func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncCronRandomBiasSeconds)) * time.Second)
		logger.Info.Println(fmt.Sprintf("ticker taskrun %s", objRun.Name))
		if objRun.Status.RunStatus == opsv1.StatusEmpty || objRun.Status.RunStatus == opsv1.StatusRunning {
			return
		}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: objRun.Namespace, Name: objRun.Name}, objRun)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		obj := &opsv1.Task{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: objRun.Namespace, Name: objRun.Spec.TaskRef}, obj)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		r.run(logger, ctx, obj, objRun)
	})
	if err != nil {
		logger.Error.Println(err)
		return
	}
	r.crontabMap[objRun.GetUniqueKey()] = id
}

func (r *TaskRunReconciler) registerClearCron() {
	if r.clearCron != nil {
		return
	}
	r.clearCron = cron.New()
	r.clearCron.AddFunc(opsv1.ClearCronTab, func() {
		objs := &opsv1.TaskRunList{}
		err := r.Client.List(context.Background(), objs)
		if err != nil {
			return
		}
		for _, obj := range objs.Items {
			if obj.Spec.Crontab != "" {
				continue
			}
			if obj.Status.RunStatus == opsv1.StatusRunning || obj.Status.RunStatus == opsv1.StatusEmpty {
				continue
			}
			if obj.GetObjectMeta().GetCreationTimestamp().Add(opsv1.DefaultTTLSecondsAfterFinished * time.Second).After(time.Now()) {
				continue
			}
			r.Client.Delete(context.Background(), &obj)
		}
	})
	r.clearCron.Start()
}

func (r *TaskRunReconciler) run(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (err error) {
	tr.Patch(t)
	tr.Status.ClearNodeStatus()
	r.commitStatus(logger, ctx, tr, opsv1.StatusRunning)
	if t.IsHostTypeRef() {
		hs := []opsv1.Host{}
		if t.Spec.Selector == nil {
			h := opsv1.Host{}
			err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.GetNamespace(), Name: tr.Spec.NameRef}, &h)
			if err != nil {
				logger.Error.Println(err)
				r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
				return
			}
			// if hostname is empty, use localhost
			if len(t.Spec.NameRef) > 0 && err != nil {
				logger.Error.Println(err)
				return
			}
			hs = append(hs, h)
		} else {
			hs = r.getSelectorHosts(logger, ctx, t)
		}
		if len(hs) == 0 {
			r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
			return
		}
		r.commitStatus(logger, ctx, tr, opsv1.StatusRunning)
		var orErr error
		for _, h := range hs {
			// fill variables
			extraVariables := map[string]string{
				"hostname": h.GetHostname(),
			}
			// filled host
			if h.Spec.SecretRef != "" {
				err = filledHostFromSecret(&h, r.Client, h.Spec.SecretRef)
				if err != nil {
					logger.Error.Println("fill host secretRef error", err)
					return
				}
			}
			logger.Info.Println(fmt.Sprintf("run task %s on host %s", t.GetUniqueKey(), t.Spec.NameRef))
			cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
			err = r.runTaskOnHost(cliLogger, ctx, t, tr, &h, extraVariables)
			if err != nil {
				orErr = err
			}
			cliLogger.Flush()
		}
		if orErr != nil {
			r.commitStatus(logger, ctx, tr, opsv1.StatusFailed)
			return
		} else {
			r.commitStatus(logger, ctx, tr, opsv1.StatusSuccessed)
		}
	} else if t.IsClusterTypeRef() {
		c := &opsv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: tr.Spec.NameRef,
			},
		}
		// task > env > default
		runtimeImage := t.Spec.RuntimeImage
		if runtimeImage == "" {
			runtimeImage = os.Getenv("DEFAULT_RUNTIME_IMAGE")
		}
		if runtimeImage == "" {
			runtimeImage = opsconstants.DefaultRuntimeImage
		}
		kubeOpt := opsoption.KubeOption{
			Debug:        strings.ToLower(os.Getenv("DEBUG")) == "true",
			NodeName:     t.Spec.NodeName,
			All:          t.Spec.All,
			RuntimeImage: runtimeImage,
			OpsNamespace: opsconstants.DefaultOpsNamespace,
		}
		if tr.Spec.NameRef != opsconstants.CurrentRuntime {
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
	} else {
		r.commitStatus(logger, ctx, tr, opsv1.StatusDataInValid)
	}
	return
}

func (r *TaskRunReconciler) runTaskOnHost(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, h *opsv1.Host, variables map[string]string) (err error) {
	hc, err := opshost.NewHostConnBase64(h)
	if err != nil {
		return err
	}
	err = opstask.RunTaskOnHost(ctx, logger, t, tr, hc, opsoption.TaskOption{
		Variables: variables,
	})
	return err
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
				Variables: tr.Spec.Variables,
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

	for retries := 0; retries < CommitStatusMaxRetries; retries++ {
		latestTr := &opsv1.TaskRun{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.GetNamespace(), Name: tr.GetName()}, latestTr)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		latestTr.Status = tr.Status
		err = r.Client.Status().Update(ctx, latestTr)
		if err == nil {
			return
		}
		if !apierrors.IsConflict(err) {
			logger.Error.Println(err, "update taskrun status error")
			return
		}
		logger.Info.Println("conflict detected, retrying...", err)
		time.Sleep(1 * time.Second)
	}
	logger.Error.Println("update taskrun status failed after retries", err)
	return
}

func (r *TaskRunReconciler) getSelectorHosts(logger *opslog.Logger, ctx context.Context, t *opsv1.Task) (hosts []opsv1.Host) {
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
		WithEventFilter(
			predicate.Funcs{
				// drop reconcile for status updates
				UpdateFunc: func(e event.UpdateEvent) bool {
					if _, ok := e.ObjectOld.(*opsv1.TaskRun); !ok {
						return true
					}

					oldObject := e.ObjectOld.(*opsv1.TaskRun).DeepCopy()
					newObject := e.ObjectNew.(*opsv1.TaskRun).DeepCopy()

					oldObjectCmp := &opsv1.TaskRun{}
					newObjectCmp := &opsv1.TaskRun{}

					oldObjectCmp.Spec = oldObject.Spec
					newObjectCmp.Spec = newObject.Spec

					return !cmp.Equal(oldObjectCmp, newObjectCmp)
				},
			},
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxTaskrunConcurrentReconciles}).
		Complete(r)
}
