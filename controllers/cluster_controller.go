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

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	timeTickerStopChans map[string]chan bool
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
	logger := opslog.BuilderFileLogger(opslog.LevelInfo, false, true)
	c := &opsv1.Cluster{}
	err = r.Get(ctx, req.NamespacedName, c)

	//if deleted, stop ticker
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, r.deleteCluster(ctx, req.NamespacedName)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	// add timeticker
	r.addTimeTicker(logger, ctx, c)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Cluster{}).
		Complete(r)
}

func (r *ClusterReconciler) deleteCluster(ctx context.Context, namespacedName types.NamespacedName) error {
	_, ok := r.timeTickerStopChans[namespacedName.String()]
	if ok {
		r.timeTickerStopChans[namespacedName.String()] <- true
	}
	return nil
}

func (r *ClusterReconciler) addTimeTicker(logger *opslog.Logger, ctx context.Context, c *opsv1.Cluster) {
	// if ticker exist, return
	_, ok := r.timeTickerStopChans[c.GetUniqueKey()]
	if ok {
		return
	}
	if r.timeTickerStopChans == nil {
		r.timeTickerStopChans = make(map[string]chan bool)
	}
	r.timeTickerStopChans[c.GetUniqueKey()] = make(chan bool)
	// create ticker
	logger.Info.Println(fmt.Sprintf("start ticker for cluster %s", c.GetUniqueKey()))
	go func() {
		ticker := time.NewTicker(time.Second * opsconstants.SyncResourceStatusHeatSeconds)
		for {
			select {
			case <-r.timeTickerStopChans[c.GetUniqueKey()]:
				ticker.Stop()
				delete(r.timeTickerStopChans, c.GetUniqueKey())
				logger.Info.Println(fmt.Sprintf("stop ticker for cluster %s", c.GetUniqueKey()))
				return
			case <-ticker.C:
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
		return r.commitStatus(logger, ctx, c, nil, opsv1.StatusFailed)
	}
	status, err := kc.GetStatus()
	if err != nil {
		logger.Error.Println(err, "failed to get cluster status")
		return r.commitStatus(logger, ctx, c, nil, opsv1.StatusFailed)
	}
	err = r.commitStatus(logger, ctx, c, status, opsv1.StatusSuccessed)
	return
}

func (r *ClusterReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, c *opsv1.Cluster, overrideStatus *opsv1.ClusterStatus, status string) (err error) {
	lastC := &opsv1.Cluster{}
	err = r.Get(ctx, types.NamespacedName{Name: c.Name, Namespace: c.Namespace}, lastC)
	if err != nil {
		logger.Error.Println(err, "failed to get last cluster")
		return
	}
	if overrideStatus != nil {
		lastC.Status = *overrideStatus
	}
	if status != "" {
		lastC.Status.HeartStatus = status
	}
	lastC.Status.HeartTime = &metav1.Time{Time: time.Now()}
	err = r.Client.Status().Update(ctx, lastC)
	if err != nil {
		logger.Error.Println(err, "update cluster status error")
	}
	return
}
