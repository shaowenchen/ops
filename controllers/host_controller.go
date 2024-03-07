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
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opshost "github.com/shaowenchen/ops/pkg/host"
	opslog "github.com/shaowenchen/ops/pkg/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HostReconciler reconciles a Host object
type HostReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	timeTickerStopChans map[string]chan bool
	tickerMutex         sync.RWMutex
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=hosts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=hosts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=hosts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Host object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *HostReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	actionNs := os.Getenv("ACTIVE_NAMESPACE")
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	h := &opsv1.Host{}
	err := r.Get(ctx, req.NamespacedName, h)

	//if delete, stop ticker
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, r.deleteHost(ctx, req.NamespacedName)
	}

	if err != nil {
		return ctrl.Result{}, err
	}

	// add timeticker
	r.addTimeTicker(logger, ctx, h)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HostReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Host{}).
		Complete(r)
}

func (r *HostReconciler) deleteHost(ctx context.Context, namespacedName types.NamespacedName) error {
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

func (r *HostReconciler) addTimeTicker(logger *opslog.Logger, ctx context.Context, h *opsv1.Host) (err error) {
	// if ticker exist, return
	r.tickerMutex.RLock()
	_, ok := r.timeTickerStopChans[h.GetUniqueKey()]
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
	r.timeTickerStopChans[h.GetUniqueKey()] = make(chan bool)
	r.tickerMutex.Unlock()
	// create ticker
	logger.Info.Println(fmt.Sprintf("start ticker for host %s", h.GetUniqueKey()))
	go func() {
		ticker := time.NewTicker(time.Second * opsconstants.SyncResourceStatusHeatSeconds)
		defer ticker.Stop()
		for {
			select {
			case <-r.timeTickerStopChans[h.GetUniqueKey()]:
				ticker.Stop()
				r.tickerMutex.Lock()
				delete(r.timeTickerStopChans, h.GetUniqueKey())
				r.tickerMutex.Unlock()
				logger.Info.Println(fmt.Sprintf("stop ticker for host %s", h.GetUniqueKey()))
				return
			case <-ticker.C:
				logger.Info.Println(fmt.Sprintf("run ticker for host %s", h.GetUniqueKey()))
				r.updateStatus(logger, ctx, h)
			}
		}
	}()
	return
}

func filledHostFromSecret(h *opsv1.Host, client client.Client, secretRef string) error {
	secret := &corev1.Secret{}
	err := client.Get(context.Background(), types.NamespacedName{Name: secretRef, Namespace: h.Namespace}, secret)
	if err != nil {
		return err
	}
	if secret.Data["privatekey"] != nil {
		privateKey := secret.Data["privatekey"]
		h.Spec.PrivateKey = base64.StdEncoding.EncodeToString(privateKey)
	}
	if secret.Data["passsword"] != nil {
		password := secret.Data["passsword"]
		h.Spec.Password = base64.StdEncoding.EncodeToString(password)
	}
	return nil
}

func (r *HostReconciler) updateStatus(logger *opslog.Logger, ctx context.Context, h *opsv1.Host) (err error) {
	if h.Spec.SecretRef != "" {
		err = filledHostFromSecret(h, r.Client, h.Spec.SecretRef)
		if err != nil {
			logger.Error.Println(err, "failed to fill host secretRef")
			return r.commitStatus(logger, ctx, h, nil, opsv1.StatusFailed)
		}
	}
	hc, err := opshost.NewHostConnBase64(h)
	if err != nil {
		logger.Error.Println(err, "failed to create host connection")
		return r.commitStatus(logger, ctx, h, nil, opsv1.StatusFailed)
	}
	status, err := hc.GetStatus(ctx, false)
	if err != nil {
		logger.Error.Println(err, "failed to get host status")
		return r.commitStatus(logger, ctx, h, status, opsv1.StatusFailed)
	}
	err = r.commitStatus(logger, ctx, h, status, opsv1.StatusSuccessed)
	return
}

func (r *HostReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, h *opsv1.Host, overrideStatus *opsv1.HostStatus, status string) (err error) {
	lastH := &opsv1.Host{}
	err = r.Get(ctx, types.NamespacedName{Name: h.Name, Namespace: h.Namespace}, lastH)
	if err != nil {
		logger.Error.Println(err, "failed to get last host")
		return
	}
	if overrideStatus != nil {
		lastH.Status = *overrideStatus
	}
	if status != "" {
		lastH.Status.HeartStatus = status
	}
	lastH.Status.HeartTime = &metav1.Time{Time: time.Now()}
	err = r.Client.Status().Update(ctx, lastH)
	if err != nil {
		logger.Error.Println(err, "update host status error")
	}
	return
}
