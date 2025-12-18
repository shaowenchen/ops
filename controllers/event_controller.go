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

	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opsmetrics "github.com/shaowenchen/ops/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
func (r *EventReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	controllerName := "Event"

	// Record metrics
	defer func() {
		resultStr := "success"
		if err != nil {
			resultStr = "error"
			opsmetrics.RecordReconcileError(controllerName, req.Namespace, "reconcile_error")
		}
		opsmetrics.RecordReconcile(controllerName, req.Namespace, resultStr)
	}()

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// push event
	namespace, err := opsconstants.GetCurrentNamespace()
	if err == nil {
		go opsevent.FactoryController(namespace, opsconstants.Events, opsconstants.Setup).Publish(context.TODO(), opsevent.EventController{
			Kind: opsconstants.PipelineRuns,
		})
	}
	if !isEventsV1Available(r.Client) {
		println("events.k8s.io/v1 is not available, use core/v1.Event instead")
		return ctrl.NewControllerManagedBy(mgr).For(&corev1.Event{}).Complete(r)
	}
	println("events.k8s.io/v1 is available")
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventsv1.Event{}).
		Watches(
			&source.Kind{Type: &eventsv1.Event{}},
			&handler.Funcs{
				CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
					v1e, ok := e.Object.(*eventsv1.Event)
					if !ok {
						return
					}
					if getEventTime(v1e).Sub(time.Now().Add(-120*time.Second)) > 0 {
						opsevent.FactoryKube(v1e.Regarding.Namespace, v1e.Regarding.Kind+"s", v1e.Regarding.Name, opsconstants.Event).Publish(context.TODO(), GetEventKube(v1e))
					}
				},
				UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
					v1e, ok := e.ObjectNew.(*eventsv1.Event)
					if !ok {
						return
					}
					if getEventTime(v1e).Sub(time.Now().Add(-120*time.Second)) > 0 {
						opsevent.FactoryKube(v1e.Regarding.Namespace, v1e.Regarding.Kind+"s", v1e.Regarding.Name, opsconstants.Event).Publish(context.TODO(), GetEventKube(v1e))
					}
				},
				DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
					v1e, ok := e.Object.(*eventsv1.Event)
					if !ok {
						return
					}
					if getEventTime(v1e).Sub(time.Now().Add(-120*time.Second)) > 0 {
						opsevent.FactoryKube(v1e.Regarding.Namespace, v1e.Regarding.Kind+"s", v1e.Regarding.Name, opsconstants.Event).Publish(context.TODO(), GetEventKube(v1e))
					}
				},
			},
		).
		Complete(r)
}

func isEventsV1Available(c client.Client) bool {
	restConfig, err := config.GetConfig()
	if err != nil {
		return false
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return false
	}

	apiGroupList, err := discoveryClient.ServerGroups()
	if err != nil {
		return false
	}

	for _, group := range apiGroupList.Groups {
		if group.Name == "events.k8s.io" {
			for _, version := range group.Versions {
				if version.Version == "v1" {
					return true
				}
			}
		}
	}
	return false
}

func GetEventKube(v1e *eventsv1.Event) (ek *opsevent.EventKube) {
	eventTime := getEventTime(v1e)
	ek = &opsevent.EventKube{
		Type:      v1e.Type,
		Reason:    v1e.Reason,
		EventTime: eventTime,
		Message:   v1e.Note,
	}
	if len(v1e.ManagedFields) > 0 {
		for _, mf := range v1e.ManagedFields {
			ek.From = mf.Manager + ek.From
		}
	}
	return
}

func getEventTime(v1e *eventsv1.Event) (eventTime time.Time) {
	if !v1e.EventTime.IsZero() {
		eventTime = v1e.EventTime.Time
	} else if v1e.DeletionTimestamp != nil {
		eventTime = v1e.DeletionTimestamp.Time
	} else if !v1e.DeprecatedLastTimestamp.IsZero() {
		eventTime = v1e.DeprecatedLastTimestamp.Time
	} else if !v1e.DeprecatedFirstTimestamp.IsZero() {
		eventTime = v1e.DeprecatedFirstTimestamp.Time
	} else if !v1e.CreationTimestamp.IsZero() {
		eventTime = v1e.CreationTimestamp.Time
	} else {
		eventTime = time.Now()
	}
	return
}
