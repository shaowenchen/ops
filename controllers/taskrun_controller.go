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
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	cron "github.com/robfig/cron/v3"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
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
	actionNs := opsconstants.GetEnvActiveNamespace()
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if opsconstants.GetEnvDebug() {
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
		r.commitStatus(logger, ctx, tr, opsconstants.StatusDataInValid)
		return ctrl.Result{}, err
	}
	// add crontab
	r.addCronTab(logger, ctx, tr)
	// check run status
	if tr.Status.RunStatus != opsconstants.StatusEmpty {
		// abort running taskrun if restart or modified
		if tr.Status.RunStatus == opsconstants.StatusRunning {
			r.commitStatus(logger, ctx, tr, opsconstants.StatusAborted)
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
	logger.Info.Println(fmt.Sprintf("add ticker for taskrun %s", objRun.GetUniqueKey()))
	id, err := r.cron.AddFunc(objRun.Spec.Crontab, func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncCronRandomBias)) * time.Second)
		logger.Info.Println(fmt.Sprintf("ticker taskrun %s", objRun.GetUniqueKey()))
		if objRun.Status.RunStatus == opsconstants.StatusEmpty || objRun.Status.RunStatus == opsconstants.StatusRunning {
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
	r.clearCron.AddFunc(opsconstants.ClearCronTab, func() {
		objs := &opsv1.TaskRunList{}
		err := r.Client.List(context.Background(), objs)
		if err != nil {
			return
		}
		for _, obj := range objs.Items {
			if obj.Spec.Crontab != "" {
				continue
			}
			if obj.Status.RunStatus == opsconstants.StatusRunning || obj.Status.RunStatus == opsconstants.StatusEmpty {
				continue
			}
			if obj.GetObjectMeta().GetCreationTimestamp().Add(opsconstants.DefaultTTLSecondsAfterFinished * time.Second).After(time.Now()) {
				continue
			}
			r.Client.Delete(context.Background(), &obj)
		}
	})
	r.clearCron.Start()
}

func (r *TaskRunReconciler) run(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (err error) {
	tr.Status.ClearNodeStatus()
	r.commitStatus(logger, ctx, tr, opsconstants.StatusRunning)

	tr.MergeVariables(t)
	cluster := r.getCluster(logger, ctx, t, tr)
	hosts := r.getHosts(logger, ctx, t, tr)

	if cluster.IsCurrentCluster() && len(hosts) > 0 {
		for _, h := range hosts {
			logger.Info.Println(fmt.Sprintf("run task %s on host %s", t.GetUniqueKey(), t.Spec.Host))
			cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
			err = r.runTaskOnHost(cliLogger, ctx, t, tr, &h)
			if err != nil {
				logger.Error.Println(err)
			}
			cliLogger.Flush()
		}
	} else {
		cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()
		r.runTaskOnKube(cliLogger, ctx, t, tr, &cluster)
		cliLogger.Flush()
	}
	// get taskrun status
	finallyStatus := opsconstants.StatusSuccessed
	for _, node := range tr.Status.TaskRunNodeStatus {
		if node.RunStatus != opsconstants.StatusSuccessed {
			finallyStatus = opsconstants.StatusFailed
		}
	}
	r.commitStatus(logger, ctx, tr, finallyStatus)
	// push event
	go opsevent.FactoryTaskRun().Publish(ctx, opsevent.EventTaskRun{
		Cluster:       opsconstants.GetEnvCluster(),
		TaskRef:       tr.Spec.TaskRef,
		Desc:          tr.Spec.Desc,
		Variables:     tr.Spec.Variables,
		TaskRunStatus: tr.Status,
	})
	return
}

func (r *TaskRunReconciler) runTaskOnHost(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, h *opsv1.Host) (err error) {
	// fill variables
	variables := tr.Spec.Variables
	variables["hostname"] = h.GetHostname()

	// insert host labels
	for k, v := range h.ObjectMeta.Labels {
		variables[k] = v
	}

	// filled host
	if h.Spec.SecretRef != "" {
		err = filledHostFromSecret(h, r.Client, h.Spec.SecretRef)
		if err != nil {
			logger.Error.Println("fill host secretRef error", err)
			return
		}
	}
	// connecting
	hc, err := opshost.NewHostConnBase64(h)
	if err != nil {
		return err
	}
	err = opstask.RunTaskOnHost(ctx, logger, t, tr, hc, opsoption.TaskOption{
		Variables: variables,
	})
	return err
}

func (r *TaskRunReconciler) runTaskOnKube(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun, cluster *opsv1.Cluster) (err error) {
	// connecting
	kc, err := opskube.NewClusterConnection(cluster)
	if err != nil {
		r.commitStatus(logger, ctx, tr, opsconstants.StatusFailed)
		logger.Error.Println(err)
		return err
	}
	// build options
	hostStr := tr.GetHost(t)
	// task > env > default
	runtimeImage := t.Spec.RuntimeImage
	if runtimeImage == "" {
		runtimeImage = opsconstants.GetEnvDefaultRuntimeImage()
	}
	if runtimeImage == "" {
		runtimeImage = opsconstants.DefaultRuntimeImage
	}
	kubeOpt := opsoption.KubeOption{
		Debug:        opsconstants.GetEnvDebug(),
		NodeName:     hostStr,
		RuntimeImage: runtimeImage,
		Namespace:    opsconstants.OpsNamespace,
	}
	// run
	nodes, err := opskube.GetNodes(ctx, logger, kc.Client, kubeOpt)
	if err != nil || len(nodes) == 0 {
		r.commitStatus(logger, ctx, tr, opsconstants.StatusFailed)
		return err
	}
	r.commitStatus(logger, ctx, tr, opsconstants.StatusRunning)
	for _, node := range nodes {
		vars := tr.Spec.Variables
		vars["hostname"] = node.Name
		opstask.RunTaskOnKube(logger, t, tr, kc, &node,
			opsoption.TaskOption{
				Variables: vars,
			}, kubeOpt)
	}
	return
}

func (r *TaskRunReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, tr *opsv1.TaskRun, status string) (err error) {
	if status != "" {
		tr.Status.RunStatus = status
	}
	if tr.Status.RunStatus == opsconstants.StatusRunning {
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
		logger.Info.Println("try commit times ", retries+1, "conflict detected, retrying...", err)
		time.Sleep(3 * time.Second)
	}
	logger.Error.Println("update taskrun status failed after retries", err)
	return
}

func (r *TaskRunReconciler) getCluster(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (cluster opsv1.Cluster) {
	clusterStr := tr.GetCluster(t)
	// default current cluster
	if len(clusterStr) == 0 {
		return opsv1.NewCurrentCluster()
	}
	// get cluster
	cluster = opsv1.Cluster{}
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: tr.GetNamespace(), Name: clusterStr}, &cluster)
	if err != nil {
		logger.Error.Println(err, "failed to get cluster")
		return
	}
	return
}

func (r *TaskRunReconciler) getHosts(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (hosts []opsv1.Host) {
	hostStr := tr.GetHost(t)
	// empty host
	if len(hostStr) == 0 {
		return
	}
	// single host
	if !strings.Contains(t.Spec.Host, "=") {
		host := opsv1.Host{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: t.GetNamespace(), Name: hostStr}, &host)
		if err != nil {
			return
		}
		hosts = append(hosts, host)
		return
	}
	// selector host, eg: az=cn-hangzhou
	hostList := &opsv1.HostList{}
	selector, err := metav1.ParseToLabelSelector(t.Spec.Host)
	if err != nil {
		logger.Error.Println(err, "failed to parse label selector")
		return
	}
	labelMap, err := metav1.LabelSelectorAsMap(selector)
	if err != nil {
		logger.Error.Println(err, "failed to convert label selector to map")
		return
	}
	err = r.Client.List(ctx, hostList, client.MatchingLabels(labelMap))
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
	// push event
	go opsevent.FactoryController().Publish(context.TODO(), opsevent.EventController{
		Cluster: opsconstants.GetEnvCluster(),
		Kind:    opsconstants.KindTaskRun,
	})
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
