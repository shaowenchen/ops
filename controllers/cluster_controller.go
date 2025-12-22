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
	"time"

	"math/rand"
	"sync"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsmetrics "github.com/shaowenchen/ops/pkg/metrics"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	timeTickerStopChans map[string]chan bool
	tickerMutex         sync.RWMutex
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	controllerName := "Cluster"

	// Record metrics
	defer func() {
		resultStr := "success"
		if err != nil {
			resultStr = "error"
			opsmetrics.RecordReconcileError(controllerName, req.Namespace, "reconcile_error")
		}
		opsmetrics.RecordReconcile(controllerName, req.Namespace, resultStr)
	}()

	actionNs := opsconstants.GetEnvActiveNamespace()
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if opsconstants.GetEnvDebug() {
		logger.SetVerbose("debug").Build()
	}
	c := &opsv1.Cluster{}
	err = r.Get(ctx, req.NamespacedName, c)

	//if deleted, stop ticker
	if apierrors.IsNotFound(err) {
		// Record Cluster info metrics as deleted (value=0)
		opsmetrics.RecordClusterInfo(req.Namespace, req.Name, "", 0)
		return ctrl.Result{}, r.deleteCluster(ctx, req.NamespacedName)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	// Record Cluster info metrics (static fields only)
	opsmetrics.RecordClusterInfo(c.Namespace, c.Name, c.Spec.Server, 1)
	// Record Cluster status metrics (dynamic fields)
	status := c.Status.HeartStatus
	if status == "" {
		status = "Unknown"
	}
	opsmetrics.RecordClusterStatus(c.Namespace, c.Name, c.Status.Version, status, c.Status.Node, c.Status.Pod, c.Status.RunningPod, c.Status.CertNotAfterDays)
	// add timeticker
	r.addTimeTicker(logger, ctx, c)
	// sync tasks and pipelines
	r.syncResource(logger, ctx, c)

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) syncResource(logger *opslog.Logger, ctx context.Context, c *opsv1.Cluster) {
	kc, err := opskube.NewClusterConnection(c)
	if err != nil {
		logger.Error.Println(err, "failed to create cluster connection")
		return
	}
	// sync tasks
	taskList := &opsv1.TaskList{}
	err = r.List(ctx, taskList, &client.ListOptions{Namespace: c.Namespace})
	if err != nil {
		logger.Error.Println(err, "failed to list tasks")
		return
	}
	err = kc.SyncTasks(false, taskList.Items)
	if err != nil {
		logger.Error.Println(err, "failed to sync tasks")
		return
	}
	// sync pipelines
	pipelineList := &opsv1.PipelineList{}
	err = r.List(ctx, pipelineList, &client.ListOptions{Namespace: c.Namespace})
	if err != nil {
		logger.Error.Println(err, "failed to list pipelines")
		return
	}
	err = kc.SyncPipelines(false, pipelineList.Items)
	if err != nil {
		logger.Error.Println(err, "failed to sync all pipelines")
		return
	}
}

func (r *ClusterReconciler) deleteCluster(ctx context.Context, namespacedName types.NamespacedName) error {
	r.tickerMutex.RLock()
	_, ok := r.timeTickerStopChans[namespacedName.String()]
	r.tickerMutex.RUnlock()
	if ok {
		r.timeTickerStopChans[namespacedName.String()] <- true
		r.tickerMutex.Lock()
		delete(r.timeTickerStopChans, namespacedName.String())
		r.tickerMutex.Unlock()
	}
	return nil
}

func (r *ClusterReconciler) addTimeTicker(logger *opslog.Logger, ctx context.Context, c *opsv1.Cluster) {
	// if ticker exist, return
	r.tickerMutex.RLock()
	_, ok := r.timeTickerStopChans[c.GetUniqueKey()]
	r.tickerMutex.RUnlock()
	if ok {
		return
	}
	if r.timeTickerStopChans == nil {
		r.tickerMutex.Lock()
		r.timeTickerStopChans = make(map[string]chan bool)
		r.tickerMutex.Unlock()
	}
	r.tickerMutex.Lock()
	r.timeTickerStopChans[c.GetUniqueKey()] = make(chan bool)
	r.tickerMutex.Unlock()
	// create ticker
	logger.Info.Println(fmt.Sprintf("start ticker for cluster %s", c.GetUniqueKey()))
	go func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncResourceRandomBiasSeconds)) * time.Second)
		ticker := time.NewTicker(time.Second * opsconstants.SyncResourceStatusHeatSeconds)
		defer ticker.Stop()
		for {
			select {
			case <-r.timeTickerStopChans[c.GetUniqueKey()]:
				ticker.Stop()
				r.tickerMutex.Lock()
				delete(r.timeTickerStopChans, c.GetUniqueKey())
				r.tickerMutex.Unlock()
				logger.Info.Println(fmt.Sprintf("stop ticker for cluster %s", c.GetUniqueKey()))
				return
			case <-ticker.C:
				logger.Info.Println(fmt.Sprintf("run ticker for cluster %s", c.GetUniqueKey()))
				r.updateStatus(logger, ctx, c)
			}
		}
	}()
	return
}

func (r *ClusterReconciler) updateStatus(logger *opslog.Logger, ctx context.Context, c *opsv1.Cluster) (err error) {
	kc, err := opskube.NewClusterConnection(c)
	if err != nil {
		logger.Error.Println(err, "failed to create cluster connection")
		return r.commitStatus(logger, ctx, c, nil, opsconstants.StatusFailed)
	}
	status, err := kc.GetStatus()
	if err != nil {
		logger.Error.Println(err, "failed to get cluster status")
	}
	err = r.commitStatus(logger, ctx, c, status, "")
	// push event
	go opsevent.FactoryCluster(c.Namespace, c.Name, opsconstants.Status).Publish(ctx, opsevent.EventCluster{
		Server: c.Spec.Server,
		Status: *status,
	})
	return
}

func (r *ClusterReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, c *opsv1.Cluster, overrideStatus *opsv1.ClusterStatus, status string) (err error) {
	lastC := &opsv1.Cluster{}
	err = r.Get(ctx, types.NamespacedName{Name: c.Name, Namespace: c.Namespace}, lastC)
	if err != nil {
		logger.Error.Println(err, "failed to get last cluster")
		return
	}
	_ = lastC.Status.HeartStatus // oldStatus reserved for future use
	if overrideStatus != nil {
		lastC.Status = *overrideStatus
	}
	if status != "" {
		lastC.Status.HeartStatus = status
	}
	lastC.Status.HeartTime = &metav1.Time{Time: time.Now()}
	err = r.Client.Status().Update(ctx, lastC)
	if err == nil {
		// Record Cluster info metrics (static fields only)
		opsmetrics.RecordClusterInfo(lastC.Namespace, lastC.Name, lastC.Spec.Server, 1)
		// Record Cluster status metrics (dynamic fields)
		status := lastC.Status.HeartStatus
		if status == "" {
			status = "Unknown"
		}
		opsmetrics.RecordClusterStatus(lastC.Namespace, lastC.Name, lastC.Status.Version, status, lastC.Status.Node, lastC.Status.Pod, lastC.Status.RunningPod, lastC.Status.CertNotAfterDays)
	} else {
		logger.Error.Println(err, "update cluster status error")
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// push event
	namespace, err := opsconstants.GetCurrentNamespace()
	if err == nil {
		go opsevent.FactoryController(namespace, opsconstants.Clusters, opsconstants.Setup).Publish(context.TODO(), opsevent.EventController{
			Kind: opsconstants.Hosts,
		})
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Cluster{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxResourceConcurrentReconciles}).
		Complete(r)
}
