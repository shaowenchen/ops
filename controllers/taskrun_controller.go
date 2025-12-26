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
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	cron "github.com/robfig/cron/v3"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opshost "github.com/shaowenchen/ops/pkg/host"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsmetrics "github.com/shaowenchen/ops/pkg/metrics"
	opsoption "github.com/shaowenchen/ops/pkg/option"
	opstask "github.com/shaowenchen/ops/pkg/task"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	Scheme            *runtime.Scheme
	crontabMap        map[string]cron.EntryID
	crontabMapMutex   sync.RWMutex
	cron              *cron.Cron
	clearCron         *cron.Cron
	opsserverEndpoint string
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
func (r *TaskRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	controllerName := "TaskRun"

	// Record metrics
	defer func() {
		resultStr := "success"
		if err != nil {
			resultStr = "error"
			opsmetrics.RecordReconcileError(controllerName, req.Namespace, "reconcile_error")
		}
		opsmetrics.RecordReconcile(controllerName, req.Namespace, resultStr)
	}()

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
		r.cron = cron.New(cron.WithLocation(time.Local))
		r.cron.Start()
	}

	// get taskrun
	tr := &opsv1.TaskRun{}
	err = r.Client.Get(ctx, req.NamespacedName, tr)
	if apierrors.IsNotFound(err) {
		// Record TaskRun info metrics as deleted (value=0)
		opsmetrics.RecordTaskRunInfo(req.Namespace, req.Name, "", "", 0)
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
	// if has crontab, start timer and set status to Successed (don't run)
	if tr.Spec.Crontab != "" {
		r.addCronTab(logger, ctx, tr)
		// scheduled taskrun should always be Successed after timer started
		r.commitStatus(logger, ctx, tr, opsconstants.StatusSuccessed)
		return ctrl.Result{}, nil
	}
	// check run status
	if tr.Status.RunStatus != opsconstants.StatusEmpty {
		// abort running taskrun if restart or modified
		if tr.Status.RunStatus == opsconstants.StatusRunning {
			r.commitStatus(logger, ctx, tr, opsconstants.StatusAborted)
		}
		return ctrl.Result{}, nil
	}
	// run task (no crontab)
	err = r.run(logger, ctx, t, tr)
	if err != nil {
		return ctrl.Result{}, err
	}
	// send event
	err = opsevent.FactoryTaskRun(tr.Namespace, tr.Name, opsconstants.Status).Publish(ctx, tr)
	return ctrl.Result{}, nil
}

func (r *TaskRunReconciler) deleteCronTab(logger *opslog.Logger, ctx context.Context, namespacedName types.NamespacedName) error {
	r.crontabMapMutex.Lock()
	defer r.crontabMapMutex.Unlock()

	entryID, ok := r.crontabMap[namespacedName.String()]
	if ok {
		r.cron.Remove(entryID)
		delete(r.crontabMap, namespacedName.String())
		logger.Info.Println(fmt.Sprintf("clear ticker for taskrun %s", namespacedName.String()))
	}
	return nil
}

func (r *TaskRunReconciler) addCronTab(logger *opslog.Logger, ctx context.Context, objRun *opsv1.TaskRun) {
	key := objRun.GetUniqueKey()

	// if crontab is empty, remove existing cron if any
	if objRun.Spec.Crontab == "" {
		r.crontabMapMutex.Lock()
		entryID, ok := r.crontabMap[key]
		if ok {
			r.cron.Remove(entryID)
			delete(r.crontabMap, key)
			logger.Info.Println(fmt.Sprintf("clear ticker for taskrun %s (crontab is empty)", key))
		}
		r.crontabMapMutex.Unlock()
		return
	}

	// check if cron already exists
	r.crontabMapMutex.Lock()
	oldEntryID, exists := r.crontabMap[key]
	if exists {
		// remove old cron entry if exists
		r.cron.Remove(oldEntryID)
		delete(r.crontabMap, key)
		logger.Info.Println(fmt.Sprintf("remove old ticker for taskrun %s (crontab updated)", key))
	}
	r.crontabMapMutex.Unlock()

	logger.Info.Println(fmt.Sprintf("add ticker for taskrun %s with crontab %s", key, objRun.Spec.Crontab))

	id, err := r.cron.AddFunc(objRun.Spec.Crontab, func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncCronRandomBias)) * time.Second)
		// get the scheduled TaskRun to verify it still exists
		scheduledTr := &opsv1.TaskRun{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: objRun.Namespace, Name: objRun.Name}, scheduledTr)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		// verify it still has crontab
		if scheduledTr.Spec.Crontab == "" {
			logger.Info.Println(fmt.Sprintf("skip taskrun %s: crontab removed", scheduledTr.GetUniqueKey()))
			return
		}
		// check if there's already a running instance created by this scheduled taskrun
		taskRunList := &opsv1.TaskRunList{}
		labelSelector := client.MatchingLabels{
			opsconstants.LabelScheduledByKey:   scheduledTr.Name,
			opsconstants.LabelScheduledKindKey: opsconstants.TaskRun,
		}
		err = r.Client.List(ctx, taskRunList, client.InNamespace(scheduledTr.Namespace), labelSelector)
		if err != nil {
			logger.Error.Println(fmt.Sprintf("failed to list taskruns: %v", err))
			return
		}
		for _, tr := range taskRunList.Items {
			// check if this taskrun is running
			if tr.Status.RunStatus == opsconstants.StatusRunning || tr.Status.RunStatus == opsconstants.StatusEmpty {
				logger.Info.Println(fmt.Sprintf("skip taskrun %s: already has a running instance %s", scheduledTr.GetUniqueKey(), tr.GetUniqueKey()))
				return
			}
		}
		logger.Info.Println(fmt.Sprintf("cron triggered for taskrun %s, creating new execution instance", scheduledTr.GetUniqueKey()))
		// get task
		t := &opsv1.Task{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: scheduledTr.Namespace, Name: scheduledTr.Spec.TaskRef}, t)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		// create a new TaskRun without crontab for execution
		newTr := opsv1.NewTaskRun(t)
		newTr.Spec.Crontab = "" // ensure no crontab
		// copy variables from scheduled taskrun (deep copy)
		if scheduledTr.Spec.Variables != nil {
			newTr.Spec.Variables = make(map[string]string)
			for k, v := range scheduledTr.Spec.Variables {
				newTr.Spec.Variables[k] = v
			}
		}
		newTr.Spec.Desc = scheduledTr.Spec.Desc // copy desc
		// ensure labels map exists (NewTaskRun already creates it with LabelTaskRefKey)
		if newTr.Labels == nil {
			newTr.Labels = make(map[string]string)
		}
		// ensure taskref label is set
		newTr.Labels[opsconstants.LabelTaskRefKey] = t.ObjectMeta.GetName()
		// set labels to identify this is created by the scheduled taskrun
		newTr.Labels[opsconstants.LabelScheduledByKey] = scheduledTr.Name
		newTr.Labels[opsconstants.LabelScheduledKindKey] = opsconstants.TaskRun
		// set owner reference to the scheduled taskrun
		if scheduledTr.UID != "" {
			newTr.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: opsconstants.APIVersion,
					Kind:       opsconstants.TaskRun,
					Name:       scheduledTr.Name,
					UID:        scheduledTr.UID,
				},
			}
		}
		// create the new taskrun
		err = r.Client.Create(ctx, &newTr)
		if err != nil {
			logger.Error.Println(fmt.Sprintf("failed to create new taskrun for scheduled taskrun %s: %v", scheduledTr.GetUniqueKey(), err))
			return
		}
		logger.Info.Println(fmt.Sprintf("created new taskrun %s for scheduled taskrun %s", newTr.GetUniqueKey(), scheduledTr.GetUniqueKey()))
	})

	if err != nil {
		logger.Error.Println(err)
		return
	}

	r.crontabMapMutex.Lock()
	r.crontabMap[key] = id
	r.crontabMapMutex.Unlock()
}

func (r *TaskRunReconciler) registerClearCron() {
	if r.clearCron != nil {
		return
	}
	r.clearCron = cron.New(cron.WithLocation(time.Local))
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
	hosts := r.getAvaliableHosts(logger, ctx, t, tr)

	cliLogger := opslog.NewLogger().SetStd().WaitFlush().Build()

	// only run script
	if len(hosts) > 0 && t.OnlyScript() && !t.NeedKubeExecution() {
		for _, h := range hosts {
			logger.Info.Printf("run task %s on host %s", t.GetUniqueKey(), t.Spec.Host)
			err = r.runTaskOnHost(cliLogger, ctx, r.Client, t, tr, &h)
			if err != nil {
				logger.Error.Println(err)
			}
			cliLogger.Flush()
		}
	} else {
		cluster := opsv1.NewCurrentCluster()

		logger.Info.Printf("run task %s on cluster %s", t.GetUniqueKey(), cluster.Name)
		err = r.runTaskOnKube(cliLogger, ctx, t, tr, &cluster)
		if err != nil {
			logger.Error.Println(err)
		}
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
	go opsevent.FactoryTaskRun(tr.Namespace, tr.Name, opsconstants.Status).Publish(ctx, opsevent.EventTaskRun{
		TaskRef:       tr.Spec.TaskRef,
		Desc:          tr.Spec.Desc,
		Variables:     tr.Spec.Variables,
		TaskRunStatus: tr.Status,
	})
	return
}

func (r *TaskRunReconciler) runTaskOnHost(logger *opslog.Logger, ctx context.Context, client client.Client, t *opsv1.Task, tr *opsv1.TaskRun, h *opsv1.Host) (err error) {
	// fill variables
	if tr.Spec.Variables == nil {
		tr.Spec.Variables = make(map[string]string)
	}
	vars := tr.Spec.Variables
	vars["TASK"] = t.Name
	vars["TASKRUN"] = tr.Name
	vars["HOSTNAME"] = h.GetHostname()
	vars["NAMESPACE"] = tr.Namespace
	opsserverEndpoint := r.getOpsServerEndpoint(logger, t.Namespace)
	if opsserverEndpoint != "" {
		vars["OPSSERVER_ENDPOINT"] = opsserverEndpoint
		logger.Debug.Printf("injected OPSSERVER_ENDPOINT: %s", opsserverEndpoint)
	} else {
		logger.Info.Println("failed to get OPSSERVER_ENDPOINT, variable not set")
	}
	vars["EVENT_CLUSTER"] = opsconstants.GetEnvEventCluster()

	// insert host labels
	for k, v := range h.ObjectMeta.Labels {
		vars[k] = v
	}

	// filled host
	if h.Spec.SecretRef != "" {
		err = filledHostFromSecret(h, client, h.Spec.SecretRef)
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
		Variables: vars,
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
	// if find host in cluster, and can connect
	host, _ := kc.GetHost(opsconstants.OpsNamespace, tr.GetHost(t))
	if host != nil && !t.NeedKubeExecution() {
		logger.Debug.Println("use host credentials to run cluster task " + tr.Name)
		return r.runTaskOnHost(logger, ctx, *kc.OpsClient, t, tr, host)
	}
	// else use pod to run task
	// build options
	hostStr := tr.GetHost(t)
	// Priority: step > task > pipeline > env > default
	// Note: step-level runtimeImage is handled in runStepShellOnKube/runStepFileOnKube
	runtimeImage := t.Spec.RuntimeImage
	if runtimeImage == "" {
		// Check Pipeline runtimeImage from spec
		if tr.Spec.RuntimeImage != "" {
			runtimeImage = tr.Spec.RuntimeImage
		}
	}
	if runtimeImage == "" {
		runtimeImage = opsconstants.GetEnvDefaultRuntimeImage()
	}
	if runtimeImage == "" {
		runtimeImage = opsconstants.DefaultRuntimeImage
	}
	// Convert Task mounts to MountConfig
	mountConfigs := make([]opsoption.MountConfig, 0)
	for _, taskMount := range t.Spec.Mounts {
		mountConfig := opsoption.MountConfig{}
		if taskMount.Secret != nil {
			// Secret mount
			mountConfig.Secret = &opsoption.SecretMountConfig{
				Name:      taskMount.Secret.Name,
				MountPath: taskMount.Secret.MountPath,
			}
		} else if taskMount.ConfigMap != nil {
			// ConfigMap mount
			mountConfig.ConfigMap = &opsoption.ConfigMapMountConfig{
				Name:      taskMount.ConfigMap.Name,
				MountPath: taskMount.ConfigMap.MountPath,
			}
		} else {
			// HostPath mount
			mountConfig.HostPath = taskMount.HostPath
			mountConfig.MountPath = taskMount.MountPath
		}
		mountConfigs = append(mountConfigs, mountConfig)
	}

	kubeOpt := opsoption.KubeOption{
		Debug:        opsconstants.GetEnvDebug(),
		NodeName:     hostStr,
		RuntimeImage: runtimeImage,
		Namespace:    opsconstants.OpsNamespace,
		Mounts:       mountConfigs,
	}
	// run
	if kubeOpt.NodeName == "" {
		kubeOpt.NodeName = opsconstants.AnyWorker
	}
	nodes, err := opskube.GetNodes(ctx, logger, kc.Client, kubeOpt)
	if err != nil {
		r.commitStatus(logger, ctx, tr, opsconstants.StatusFailed)
		return err
	}
	r.commitStatus(logger, ctx, tr, opsconstants.StatusRunning)
	for _, node := range nodes {
		if tr.Spec.Variables == nil {
			tr.Spec.Variables = make(map[string]string)
		}
		vars := tr.Spec.Variables
		vars["HOSTNAME"] = node.Name
		vars["NAMESPACE"] = tr.Namespace
		opsserverEndpoint := r.getOpsServerEndpoint(logger, t.Namespace)
		if opsserverEndpoint != "" {
			vars["OPSSERVER_ENDPOINT"] = opsserverEndpoint
			logger.Debug.Printf("injected OPSSERVER_ENDPOINT: %s", opsserverEndpoint)
		} else {
			logger.Info.Println("failed to get OPSSERVER_ENDPOINT, variable not set")
		}
		vars["TASK"] = t.Name
		vars["TASKRUN"] = tr.Name
		opstask.RunTaskOnKube(logger, t, tr, kc, &node,
			opsoption.TaskOption{
				Variables: vars,
			}, kubeOpt)
	}
	return
}

func (r *TaskRunReconciler) getOpsServerEndpoint(logger *opslog.Logger, namespace string) string {
	// get ops server service and return cluster-internal address
	if len(r.opsserverEndpoint) > 0 {
		return r.opsserverEndpoint
	}
	// get svc
	serviceList := &corev1.ServiceList{}
	labelSelector := labels.SelectorFromSet(labels.Set{
		opsconstants.LabelOpsServerKey: opsconstants.LabelOpsServerValue,
		opsconstants.LabelOpsPartOf:    opsconstants.LabelOpsPartOfValue,
	})

	listOptions := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	}

	err := r.List(context.Background(), serviceList, listOptions)
	if err != nil {
		if logger != nil {
			logger.Error.Printf("failed to list services for OPSSERVER_ENDPOINT: %v", err)
		}
		return ""
	}
	if len(serviceList.Items) == 0 {
		if logger != nil {
			logger.Info.Printf("no service found with label %s=%s in namespace %s", opsconstants.LabelOpsServerKey, opsconstants.LabelOpsServerValue, namespace)
		}
		return ""
	}
	svc := serviceList.Items[0]
	if len(svc.Spec.Ports) == 0 {
		if logger != nil {
			logger.Info.Printf("service %s/%s has no ports configured", svc.Namespace, svc.Name)
		}
		return ""
	}
	// find port by name "http", otherwise use port 80 as default
	var port int32 = 80
	found := false
	for _, svcPort := range svc.Spec.Ports {
		if svcPort.Name == "http" {
			port = svcPort.Port
			if port == 0 {
				// if Port is 0, try TargetPort (but prefer Port as it's the service exposed port)
				if svcPort.TargetPort.IntVal > 0 {
					port = svcPort.TargetPort.IntVal
				} else {
					// if TargetPort is also 0 or string, use default 80
					port = 80
				}
			}
			found = true
			break
		}
	}
	if !found {
		// use default port 80 if no "http" named port found
		port = 80
		if logger != nil {
			logger.Info.Printf("service %s/%s has no port named 'http', using default port 80", svc.Namespace, svc.Name)
		}
	}
	// use service name only (assuming same namespace, Kubernetes will resolve it)
	// format: http://<service-name>:<port>
	r.opsserverEndpoint = fmt.Sprintf("http://%s:%d", svc.Name, port)
	if logger != nil {
		logger.Info.Printf("OPSSERVER_ENDPOINT resolved: %s", r.opsserverEndpoint)
	}
	return r.opsserverEndpoint
}

func (r *TaskRunReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, tr *opsv1.TaskRun, status string) (err error) {
	oldStatus := tr.Status.RunStatus
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
		// merge status
		if tr.Status.TaskRunNodeStatus != nil {
			latestTr.Status.TaskRunNodeStatus = tr.Status.TaskRunNodeStatus
		}
		latestTr.Status = tr.Status
		err = r.Client.Status().Update(ctx, latestTr)
		if err == nil {
			// Record CRD resource status change metrics - record every status change (value=1 for existing resource)
			if oldStatus != latestTr.Status.RunStatus {
				// Record TaskRun info metrics (static fields only)
				opsmetrics.RecordTaskRunInfo(latestTr.Namespace, latestTr.Name, latestTr.Spec.TaskRef, latestTr.Spec.Crontab, 1)
				// Record TaskRun status metrics (dynamic fields)
				opsmetrics.RecordTaskRunStatus(latestTr.Namespace, latestTr.Name, latestTr.Status.RunStatus)
				// Record scheduled task status change if this is a scheduled task (has Crontab)
				if latestTr.Spec.Crontab != "" {
					opsmetrics.RecordTaskRunInfo(latestTr.Namespace, latestTr.Name, latestTr.Spec.TaskRef, latestTr.Spec.Crontab, 1)
					opsmetrics.RecordTaskRunStatus(latestTr.Namespace, latestTr.Name, latestTr.Status.RunStatus)
				}
				// Record TaskRef status phase change (decrement old status, increment new status)
				if latestTr.Spec.TaskRef != "" {
					opsmetrics.RecordTaskRunStatusPhase(latestTr.Namespace, latestTr.Spec.TaskRef, oldStatus, latestTr.Status.RunStatus)
				}
			}
			//need to improve
			time.Sleep(3 * time.Second)
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

func (r *TaskRunReconciler) getAvaliableHosts(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (hosts []opsv1.Host) {
	selectHosts := r.getHosts(logger, ctx, t, tr)
	for _, host := range selectHosts {
		if host.Status.HeartStatus == opsconstants.StatusSuccessed {
			hosts = append(hosts, host)
		}
	}
	return
}

func (r *TaskRunReconciler) getHosts(logger *opslog.Logger, ctx context.Context, t *opsv1.Task, tr *opsv1.TaskRun) (hosts []opsv1.Host) {
	if t.Spec.RuntimeImage != "" {
		return
	}
	hostStr := tr.GetHost(t)
	// empty host
	if len(hostStr) == 0 {
		return
	}
	// anynode
	if opsconstants.IsAnyKubeNode(hostStr) {
		nodes := &corev1.NodeList{}
		nodes, err := opsutils.GetAllReadyNodesByReconcileClient(r.Client)
		if err != nil {
			logger.Error.Println(err, "failed to list nodes")
		}
		hosts := &opsv1.HostList{}
		err = r.Client.List(ctx, hosts)
		if err != nil {
			logger.Error.Println(err, "failed to list hosts")
		}
		// find node
		var targetNode *corev1.Node
		for _, node := range nodes.Items {
			if opsutils.IsMasterNode(&node) && opsconstants.IsAnyMaster(hostStr) {
				targetNode = &node
			} else if !opsutils.IsMasterNode(&node) && opsconstants.IsAnyWorker(hostStr) {
				targetNode = &node
			} else if opsconstants.IsAnyNode(hostStr) {
				targetNode = &node
			}
			if targetNode != nil {
				break
			}
		}
		// find host
		for _, host := range hosts.Items {
			if host.Spec.Address == opsutils.GetNodeInternalIp(targetNode) {
				return []opsv1.Host{host}
			}
		}

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
	// selector host, eg: alert-card=npu
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
	// push event
	namespace, err := opsconstants.GetCurrentNamespace()
	if err == nil {
		go opsevent.FactoryController(namespace, opsconstants.TaskRuns, opsconstants.Setup).Publish(context.TODO(), opsevent.EventController{
			Kind: opsconstants.TaskRuns,
		})
	}
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
