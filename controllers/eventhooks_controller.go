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

	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opseventhook "github.com/shaowenchen/ops/pkg/eventhook"
	opslog "github.com/shaowenchen/ops/pkg/log"
)

// EventHooksReconciler reconciles a EventHooks object
type EventHooksReconciler struct {
	client.Client
	Scheme                      *runtime.Scheme
	subjectEventBusMapMutex     sync.RWMutex
	subjectEventBusMap          map[string]opsevent.EventBus
	objSubjectMapMutex          sync.RWMutex
	objSubjectMap               map[string]string
	subjectSubscriptCancelMutex sync.RWMutex
	subjectSubscriptCancel      map[string]context.CancelFunc
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
	r.update(logger, ctx, obj)
	return ctrl.Result{}, nil
}

func (r *EventHooksReconciler) update(logger *opslog.Logger, ctx context.Context, obj *opsv1.EventHooks) error {
	r.subjectEventBusMapMutex.Lock()
	if obj.Spec.Subject == "" {
		return nil
	}
	subject := obj.Spec.Subject
	if r.subjectEventBusMap == nil {
		r.subjectEventBusMap = make(map[string]opsevent.EventBus)
	}
	if r.objSubjectMap == nil {
		r.objSubjectMap = make(map[string]string)
	}
	if r.subjectSubscriptCancel == nil {
		r.subjectSubscriptCancel = make(map[string]context.CancelFunc)
	}
	if _, ok := r.objSubjectMap[obj.Name]; !ok {
		r.objSubjectMap[obj.Name] = subject
	}
	existingEventHooksList := &opsv1.EventHooksList{}
	listOpts := []client.ListOption{
		client.InNamespace(obj.Namespace),
		client.MatchingFields{".spec.subject": subject},
	}
	if err := r.List(ctx, existingEventHooksList, listOpts...); err != nil {
		return err
	}
	client := opsevent.EventBus{}
	if _, ok := r.subjectEventBusMap[subject]; ok {
		client = r.subjectEventBusMap[subject]
	} else {
		client.WithEndpoint(os.Getenv("EVENT_ENDPOINT")).WithSubject(subject)
	}
	for _, eventhook := range existingEventHooksList.Items {
		logger.Info.Println(fmt.Sprintf("registry eventhook %s for subject %s", eventhook.ObjectMeta.Name, subject))
		client.AddConsumerFunc(func(ctx context.Context, event cloudevents.Event) {
			eventStrings := opsevent.GetCloudEventReadable(event)
			notification := true
			if len(eventhook.Spec.Keywords) > 0 {
				notification = false
				for _, keyword := range obj.Spec.Keywords {
					if strings.Contains(eventStrings, keyword) {
						notification = true
						logger.Info.Println(fmt.Sprintf("event %s contains keyword %s trigger eventhook %s", event.ID(), keyword, eventhook.ObjectMeta.Name))
						break
					}
				}
			}
			if notification && opseventhook.NotificationMap[obj.Spec.Type] != nil {
				go opseventhook.NotificationMap[obj.Spec.Type].Post(obj.Spec.URL, obj.Spec.Options, eventStrings, obj.Spec.Additional)
			}

		})
		r.subjectEventBusMap[eventhook.Name] = client
	}
	r.subjectEventBusMapMutex.Unlock()

	r.subjectSubscriptCancelMutex.Lock()
	if cancle, ok := r.subjectSubscriptCancel[subject]; ok {
		cancle()
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.subjectSubscriptCancel[subject] = cancel
	r.subjectSubscriptCancelMutex.Unlock()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				client.Subscribe(ctx)
			}
		}
	}()
	return nil
}

func (r *EventHooksReconciler) delete(logger *opslog.Logger, ctx context.Context, namespacedName types.NamespacedName) error {
	if r.objSubjectMap == nil {
		return nil
	}
	if subject, ok := r.objSubjectMap[namespacedName.Name]; ok {
		r.update(logger, ctx, &opsv1.EventHooks{
			Spec: opsv1.EventHooksSpec{
				Subject: subject,
			},
		})
		r.objSubjectMapMutex.Lock()
		delete(r.objSubjectMap, namespacedName.Name)
		r.objSubjectMapMutex.Unlock()
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
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &opsv1.EventHooks{}, ".spec.subject", func(rawObj client.Object) []string {
		tr := rawObj.(*opsv1.EventHooks)
		return []string{tr.Spec.Subject}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.EventHooks{}).
		Complete(r)
}
