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

	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	appsv1 "k8s.io/api/apps/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// EventReconciler reconciles a Event object
type EventReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=events/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=events/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Event object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *EventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Setup the controller to manage corev1.Event
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventsv1.Event{}).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.Funcs{
				CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
					opsevent.FactoryDeployment(e.Object.GetNamespace(), e.Object.GetName(), opsconstants.Create).Publish(context.TODO(), e)
				},
				UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
					opsevent.FactoryDeployment(e.ObjectOld.GetNamespace(), e.ObjectOld.GetName(), opsconstants.Update).Publish(context.TODO(), e)
				},
				DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
					opsevent.FactoryDeployment(e.Object.GetNamespace(), e.Object.GetName(), opsconstants.Delete).Publish(context.TODO(), e)
				},
			},
		).
		Complete(r)
}