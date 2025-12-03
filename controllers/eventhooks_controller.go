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
	"fmt"
	"os"
	"strings"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"time"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opseventhook "github.com/shaowenchen/ops/pkg/eventhook"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsmetrics "github.com/shaowenchen/ops/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// EventHooksReconciler reconciles a EventHooks object
type EventHooksReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	subjectEventBusMapMutex sync.RWMutex
	subjectEventBusMap      map[string]opsevent.EventBus
	subjectCancelMap        map[string]context.CancelFunc
	subjectCancelMapMutex   sync.RWMutex
	objSubjectMap           map[types.NamespacedName]string
	objSubjectMapMutex      sync.RWMutex
}

func (r *EventHooksReconciler) init() {
	if r.subjectEventBusMap == nil {
		r.subjectEventBusMap = make(map[string]opsevent.EventBus)
	}
	if r.subjectCancelMap == nil {
		r.subjectCancelMap = make(map[string]context.CancelFunc)
	}
	if r.objSubjectMap == nil {
		r.objSubjectMap = make(map[types.NamespacedName]string)
	}
}

// +kubebuilder:rbac:groups=crd.chenshaowen.com,resources=eventhooks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.chenshaowen.com,resources=eventhooks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crd.chenshaowen.com,resources=eventhooks/finalizers,verbs=update
func (r *EventHooksReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	startTime := time.Now()
	controllerName := "EventHooks"

	// Record metrics
	defer func() {
		duration := time.Since(startTime)
		resultStr := "success"
		if err != nil {
			resultStr = "error"
			opsmetrics.RecordReconcileError(controllerName, req.Namespace, "reconcile_error")
		}
		opsmetrics.RecordReconcile(controllerName, req.Namespace, resultStr, duration)
	}()

	r.init()
	actionNs := opsconstants.GetEnvActiveNamespace()
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if opsconstants.GetEnvDebug() {
		logger.SetVerbose("debug").Build()
	}
	obj := &opsv1.EventHooks{}
	err = r.Get(ctx, req.NamespacedName, obj)

	if apierrors.IsNotFound(err) {
		if subject, ok := r.objSubjectMap[req.NamespacedName]; ok {
			err = r.updateSubject(logger, ctx, req.Namespace, subject)
			if err != nil {
				r.objSubjectMapMutex.Lock()
				defer r.objSubjectMapMutex.Unlock()
				delete(r.objSubjectMap, req.NamespacedName)
			}
		}
		return ctrl.Result{}, err
	}

	// record for delete object and stop watch
	r.objSubjectMapMutex.Lock()
	defer r.objSubjectMapMutex.Unlock()

	r.objSubjectMap[req.NamespacedName] = obj.Spec.Subject

	if err != nil {
		return ctrl.Result{}, err
	}
	r.updateSubject(logger, ctx, obj.Namespace, obj.Spec.Subject)
	return ctrl.Result{}, nil
}

func (r *EventHooksReconciler) stopSubject(logger *opslog.Logger, ctx context.Context, namespace, subject string) {
	if subject == "" {
		return
	}
	r.subjectEventBusMapMutex.Lock()
	if busClient, ok := r.subjectEventBusMap[subject]; ok {
		busClient.Close(ctx)
		delete(r.subjectEventBusMap, subject)
	}
	r.subjectEventBusMapMutex.Unlock()
	r.subjectCancelMapMutex.Lock()
	if cancel, ok := r.subjectCancelMap[subject]; ok {
		cancel()
		delete(r.subjectCancelMap, subject)
	}
	r.subjectCancelMapMutex.Unlock()
}

func (r *EventHooksReconciler) updateSubject(logger *opslog.Logger, ctx context.Context, namespace, subject string) error {
	if subject == "" {
		logger.Info.Println(fmt.Sprintf("eventhook subject is empty, skip"))
		return nil
	}

	existingEventHooksList := &opsv1.EventHooksList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingFields{".spec.subject": subject},
	}
	if err := r.List(ctx, existingEventHooksList, listOpts...); err != nil {
		logger.Error.Println(fmt.Sprintf("failed to list eventhooks with subject %s: %v", subject, err))
		return err
	}
	if len(existingEventHooksList.Items) == 0 {
		r.stopSubject(logger, ctx, namespace, subject)
		return nil
	}
	busClient := opsevent.EventBus{}
	busClient.WithEndpoint(os.Getenv("EVENT_ENDPOINT")).WithSubject(subject)
	r.subjectEventBusMap[subject] = busClient

	var b strings.Builder
	for i, item := range existingEventHooksList.Items {
		b.WriteString(item.Name)
		if i < len(existingEventHooksList.Items)-1 {
			b.WriteString(",")
		}
	}
	logger.Info.Println(fmt.Sprintf("eventhook [%s] watch subject %s", b.String(), subject))
	busClient.AddConsumerFunc(func(ctx context.Context, event cloudevents.Event) {
		for _, eventhook := range existingEventHooksList.Items {
			r.checkEventAndHandle(logger, ctx, event, eventhook)
		}
	})
	// start subscription
	ctx, cancel := context.WithCancel(context.TODO())
	r.subjectCancelMapMutex.Lock()
	defer r.subjectCancelMapMutex.Unlock()
	if existingCancel, ok := r.subjectCancelMap[subject]; ok {
		existingCancel()
	}
	r.subjectCancelMap[subject] = cancel
	go busClient.Subscribe(ctx)
	return nil
}

func (r *EventHooksReconciler) checkEventAndHandle(logger *opslog.Logger, ctx context.Context, event cloudevents.Event, eventhook opsv1.EventHooks) {
	eventStrings := opsevent.GetCloudEventReadable(event)

	// If no keywords are configured, trigger all events
	// Otherwise, check if any keyword matches the event
	shouldTrigger := true
	if len(eventhook.Spec.Keywords) > 0 {
		// Keywords are configured, need to check for match
		shouldTrigger = false
		for _, keyword := range eventhook.Spec.Keywords {
			if strings.Contains(eventStrings, keyword) {
				shouldTrigger = true
				logger.Info.Println(fmt.Sprintf("event %s contains keyword %s trigger eventhook %s", event.ID(), keyword, eventhook.ObjectMeta.Name))
				break
			}
		}
	}

	if shouldTrigger {
		notif, ok := opseventhook.NotificationMap[eventhook.Spec.Type]
		if !ok || notif == nil {
			logger.Error.Println(fmt.Sprintf("eventhook %s type %s not found", eventhook.ObjectMeta.Name, eventhook.Spec.Type))
			return
		}
		go notif.Post(eventhook.Spec.URL, eventhook.Spec.Options, eventStrings, eventhook.Spec.Additional)
	}
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
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &opsv1.EventHooks{}, ".spec.subject", func(rawObj client.Object) []string {
		tr := rawObj.(*opsv1.EventHooks)
		return []string{tr.Spec.Subject}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.EventHooks{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxResourceConcurrentReconciles}).
		Complete(r)
}
