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
	"github.com/google/go-cmp/cmp"
	cron "github.com/robfig/cron/v3"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	crontabMap map[string]cron.EntryID
	cron       *cron.Cron
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	// default to reconcile all namespace, if ACTIVE_NAMESPACE is set, only reconcile ACTIVE_NAMESPACE
	actionNs := os.Getenv("ACTIVE_NAMESPACE")
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if os.Getenv("DEBUG") == "true" {
		logger.SetVerbose("debug").Build()
	}
	if r.crontabMap == nil {
		r.crontabMap = make(map[string]cron.EntryID)
	}
	if r.cron == nil {
		r.cron = cron.New()
		r.cron.Start()
	}

	obj := &opsv1.Pipeline{}
	err = r.Client.Get(ctx, req.NamespacedName, obj)

	//if delete, stop ticker
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, r.deleteTicker(ctx, req.NamespacedName)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	// validate crontab
	if obj.GetCrontab() == "" {
		return ctrl.Result{}, nil
	}
	// create task
	r.createObjRun(logger, ctx, obj)
	if err != nil {
		logger.Error.Println(err)
	}

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) createObjRun(logger *opslog.Logger, ctx context.Context, obj *opsv1.Pipeline) (err error) {
	_, ok := r.crontabMap[obj.GetUniqueKey()]
	if ok {
		logger.Info.Println(fmt.Sprintf("clear ticker for pipeline %s", obj.GetUniqueKey()))
		r.cron.Remove(r.crontabMap[obj.GetUniqueKey()])
	}
	if obj.GetCrontab() == "" {
		return nil
	}
	r.crontabMap[obj.GetUniqueKey()], err = r.cron.AddFunc(obj.GetCrontab(), func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncCronRandomBiasSeconds)) * time.Second)
		objRunList := opsv1.PipelineRunList{}
		err := r.Client.List(ctx, &objRunList, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				opsv1.LabelCronKey: opsv1.LabelCronPipelineValue,
				opsv1.LabelPipelineRefKey:     obj.Name,
			}),
		})
		if err != nil {
			logger.Error.Println(err)
			return
		}
		for _, objRun := range objRunList.Items {
			if objRun.Status.RunStatus == opsv1.StatusRunning || objRun.Status.RunStatus == opsv1.StatusEmpty {
				logger.Info.Println(fmt.Sprintf("skip running pipelinerun %s", objRun.Name))
				return
			}
		}
		objRun := opsv1.NewPipelineRun(obj)
		objRun.Labels = map[string]string{
			opsv1.LabelCronKey: opsv1.LabelCronPipelineValue,
			opsv1.LabelPipelineRefKey:     obj.Name,
		}
		r.Client.Create(ctx, &objRun)
	})
	logger.Info.Println(fmt.Sprintf("start ticker for pipeline %s", obj.GetUniqueKey()))
	if err != nil {
		return err
	}
	return nil
}

func (r *PipelineReconciler) deleteTicker(ctx context.Context, namespacedName types.NamespacedName) error {
	_, ok := r.crontabMap[namespacedName.String()]
	if ok {
		r.cron.Remove(r.crontabMap[namespacedName.String()])
		delete(r.crontabMap, namespacedName.String())
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Pipeline{}).
		WithEventFilter(
			predicate.Funcs{
				// drop reconcile for status updates
				UpdateFunc: func(e event.UpdateEvent) bool {
					if _, ok := e.ObjectOld.(*opsv1.Pipeline); !ok {
						return true
					}

					oldObject := e.ObjectOld.(*opsv1.Pipeline).DeepCopy()
					newObject := e.ObjectNew.(*opsv1.Pipeline).DeepCopy()

					oldObjectCmp := &opsv1.Pipeline{}
					newObjectCmp := &opsv1.Pipeline{}

					oldObjectCmp.Spec = oldObject.Spec
					newObjectCmp.Spec = newObject.Spec

					return !cmp.Equal(oldObjectCmp, newObjectCmp)
				},
			},
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxResourceConcurrentReconciles}).
		Complete(r)
}
