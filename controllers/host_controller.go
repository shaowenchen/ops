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
	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/host"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// HostReconciler reconciles a Host object
type HostReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	log := log.FromContext(ctx)

	h := &opsv1.Host{}
	err := r.Get(ctx, req.NamespacedName, h)

	if err != nil {
		log.Info(err.Error())
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	err = r.updateStatus(ctx, h, 10)
	if err != nil {
		log.Info(err.Error())
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HostReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Host{}).
		Complete(r)
}

func (r *HostReconciler) updateStatus(ctx context.Context, h *opsv1.Host, heatSecond time.Duration) (err error) {
	hc, err := host.NewHostConnection(h.Spec.Address, h.Spec.Port, h.Spec.Username, h.Spec.Password, h.Spec.PrivateKey, h.Spec.PrivateKeyPath)

	go func() {
		for range time.Tick(heatSecond * time.Second) {
			err := hc.UpdateStatus(false)
			if err == nil {
				h.Status = hc.Host.Status
				h.Status.HeartStatus = true
			} else {
				h.Status.HeartStatus = false
			}
			h.Status.LatestHeartTime = &v1.Time{Time: time.Now()}
			err = r.Status().Update(ctx, h)
		}
	}()
	return
}
