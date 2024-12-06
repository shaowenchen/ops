/*
Copyright 2024 shaowenchen.

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
	"os"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opseventhook "github.com/shaowenchen/ops/pkg/eventhook"
	opslog "github.com/shaowenchen/ops/pkg/log"
)

// EventHooksReconciler reconciles a EventHooks object
type EventHooksReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	mutex       sync.RWMutex
	eventbusMap map[string]context.CancelFunc
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=eventhooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=eventhooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=eventhooks/finalizers,verbs=update

func (r *EventHooksReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	actionNs := opsconstants.GetEnvActiveNamespace()
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if opsconstants.GetEnvDebug() {
		logger.SetVerbose("debug").Build()
	}
	obj := &opsv1.EventHooks{}
	err := r.Get(ctx, req.NamespacedName, obj)

	// If delete, stop watch and exit the goroutine
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, r.delete(logger, ctx, req.NamespacedName)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Debug.Println("Reconcile EventHooks: ", obj.Name)
	go r.create(logger, ctx, obj)
	return ctrl.Result{}, nil
}

func (r *EventHooksReconciler) create(logger *opslog.Logger, ctx context.Context, obj *opsv1.EventHooks) error {
	if r.eventbusMap == nil {
		r.eventbusMap = make(map[string]context.CancelFunc)
	}

	// Cancel existing context if it exists
	if cancel, ok := r.eventbusMap[obj.Namespace]; ok {
		cancel()
		r.mutex.Lock()
		delete(r.eventbusMap, obj.Namespace)
		r.mutex.Unlock()
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.mutex.Lock()
	r.eventbusMap[obj.Namespace] = cancel
	r.mutex.Unlock()

	client := &opsevent.EventBus{}
	opseventhook.NotificationMap[obj.Spec.Type].Post(obj.Spec.URL, obj.Spec.Options, "eventhook config success, subject: "+obj.Spec.Subject+", time: "+time.Now().Format(time.RFC3339), "")
	client.WithEndpoint(os.Getenv("EVENT_ENDPOINT")).WithSubject(obj.Spec.Subject).Subscribe(ctx, func(ctx context.Context, event cloudevents.Event) {
		select {
		case <-ctx.Done():
			logger.Debug.Println("Exiting goroutine for EventHooks: ", obj.Name)
			return
		default:
		}
		eventStrings := opsevent.GetCloudEventString(event)
		if len(obj.Spec.Keywords) > 0 {
			skip := true
			for _, keyword := range obj.Spec.Keywords {
				if strings.ContainsAny(eventStrings, keyword) {
					skip = false
					break
				}
			}
			if skip {
				return
			}
		}
		opseventhook.NotificationMap[obj.Spec.Type].Post(obj.Spec.URL, obj.Spec.Options, eventStrings, obj.Spec.Additional)
		return
	})
	return nil
}

func (r *EventHooksReconciler) delete(logger *opslog.Logger, ctx context.Context, namespacedName types.NamespacedName) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if cancel, ok := r.eventbusMap[namespacedName.Namespace]; ok {
		cancel() // Cancel the context to stop the goroutine
		delete(r.eventbusMap, namespacedName.Namespace)
		logger.Debug.Println("Deleted EventBus for namespace: ", namespacedName.Namespace)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventHooksReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// push event
	namespace, err := opsconstants.GetCurrentNamespace()
	if err == nil {
		go opsevent.FactoryController(namespace, opsconstants.EventHooks, opsconstants.Setup).Publish(context.TODO(), opsevent.EventController{
			Kind: opsconstants.EventHooks,
		})
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.EventHooks{}).
		Complete(r)
}
