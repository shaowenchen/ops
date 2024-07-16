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
	"math/rand"
	"os"
	"time"

	"github.com/google/go-cmp/cmp"
	cron "github.com/robfig/cron/v3"
	crdv1 "github.com/shaowenchen/ops/api/v1"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opslog "github.com/shaowenchen/ops/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const CommitStatusMaxRetries = 5

// PipelineRunReconciler reconciles a PipelineRun object
type PipelineRunReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	crontabMap map[string]cron.EntryID
	cron       *cron.Cron
	clearCron  *cron.Cron
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelineruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=pipelineruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PipelineRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// start clear cron
	r.registerClearCron()
	// only reconcile active namespace
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

	pr := &opsv1.PipelineRun{}
	err := r.Client.Get(ctx, req.NamespacedName, pr)

	if apierrors.IsNotFound(err) {
		r.deleteCronTab(logger, ctx, req.NamespacedName)
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	// add crontab
	r.addCronTab(logger, ctx, pr)
	// had run once, skip
	if !(pr.Status.RunStatus == opsv1.StatusEmpty || pr.Status.RunStatus == opsv1.StatusRunning) {
		return ctrl.Result{}, nil
	}
	// get pipeline
	p := &opsv1.Pipeline{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.Namespace, Name: pr.Spec.Ref}, p)
	if err != nil {
		r.commitStatus(logger, ctx, pr, opsv1.StatusFailed, "", "", nil)
		return ctrl.Result{}, err
	}

	// run
	err = r.run(logger, ctx, p, pr)
	return ctrl.Result{}, err
}

func (r *PipelineRunReconciler) deleteCronTab(logger *opslog.Logger, ctx context.Context, namespacedName types.NamespacedName) error {
	_, ok := r.crontabMap[namespacedName.String()]
	if ok {
		r.cron.Remove(r.crontabMap[namespacedName.String()])
		delete(r.crontabMap, namespacedName.String())
		logger.Info.Println(fmt.Sprintf("clear ticker for taskrun %s", namespacedName.String()))
	}
	return nil
}

func (r *PipelineRunReconciler) addCronTab(logger *opslog.Logger, ctx context.Context, objRun *opsv1.PipelineRun) {
	if objRun.Spec.Crontab == "" {
		return
	}
	_, ok := r.crontabMap[objRun.GetUniqueKey()]
	if ok {
		return
	}
	id, err := r.cron.AddFunc(objRun.Spec.Crontab, func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncCronRandomBiasSeconds)) * time.Second)
		logger.Info.Println(fmt.Sprintf("ticker pipelinerun %s", objRun.Name))
		if objRun.Status.RunStatus == opsv1.StatusEmpty || objRun.Status.RunStatus == opsv1.StatusRunning {
			return
		}
		// clear pipelinerun status
		objRun.Status = opsv1.PipelineRunStatus{}
		r.commitStatus(logger, ctx, objRun, "", "", "", nil)
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: objRun.Namespace, Name: objRun.Name}, objRun)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		obj := &opsv1.Pipeline{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: objRun.Namespace, Name: objRun.Spec.Ref}, obj)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		r.run(logger, ctx, obj, objRun)
	})
	if err != nil {
		logger.Error.Println(err)
		return
	}
	r.crontabMap[objRun.GetUniqueKey()] = id
}

func (r *PipelineRunReconciler) registerClearCron() {
	if r.clearCron != nil {
		return
	}
	r.clearCron = cron.New()
	r.clearCron.AddFunc(opsv1.ClearCronTab, func() {
		objs := &opsv1.PipelineRunList{}
		err := r.Client.List(context.Background(), objs)
		if err != nil {
			return
		}
		for _, obj := range objs.Items {
			if obj.Status.RunStatus == opsv1.StatusRunning || obj.Status.RunStatus == opsv1.StatusEmpty {
				continue
			}
			if obj.GetObjectMeta().GetCreationTimestamp().Add(opsv1.DefaultTTLSecondsAfterFinished * time.Second).After(time.Now()) {
				continue
			}
			r.Client.Delete(context.Background(), &obj)
		}
	})
	r.clearCron.Start()
}

func (r *PipelineRunReconciler) run(logger *opslog.Logger, ctx context.Context, p *opsv1.Pipeline, pr *opsv1.PipelineRun) (err error) {
	runAlways := false
	for _, tRef := range p.Spec.Tasks {
		if runAlways && !tRef.RunAlways {
			continue
		}
		// create taskrun
		t := &opsv1.Task{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.Namespace, Name: tRef.Ref}, t)
		if err != nil {
			logger.Error.Println(err)
			runAlways = true
			r.commitStatus(logger, ctx, pr, opsv1.StatusDataInValid, tRef.Name, tRef.Ref, &opsv1.TaskRunStatus{
				RunStatus: opsv1.StatusDataInValid,
			})
			continue
		}
		tr := opsv1.NewTaskRunWithPipelineRun(pr, t, tRef)
		err = r.Client.Create(ctx, tr)
		if err != nil {
			logger.Error.Println(err)
			runAlways = true
			r.commitStatus(logger, ctx, pr, opsv1.StatusDataInValid, tRef.Name, tRef.Ref, nil)
			continue
		}
		// run task and commit status
		for {
			time.Sleep(time.Second * 3)
			trRunning := &opsv1.TaskRun{}
			if err = r.Client.Get(ctx, types.NamespacedName{Namespace: tr.Namespace, Name: tr.Name}, trRunning); err != nil {
				logger.Error.Println(err)
				break
			}
			r.commitStatus(logger, ctx, pr, opsv1.StatusRunning, tRef.Name, trRunning.Spec.Ref, &trRunning.Status)
			if trRunning.Status.RunStatus == opsv1.StatusRunning || trRunning.Status.RunStatus == opsv1.StatusEmpty {
				continue
			} else if trRunning.Status.RunStatus == opsv1.StatusSuccessed {
				break
			} else {
				runAlways = true
				break
			}
		}
	}
	finallyStatus := opsv1.StatusSuccessed
	for _, status := range pr.Status.PipelineRunStatus {
		if status.TaskRunStatus.RunStatus == opsv1.StatusFailed {
			finallyStatus = opsv1.StatusFailed
			break
		} else if status.TaskRunStatus.RunStatus == opsv1.StatusDataInValid {
			finallyStatus = opsv1.StatusDataInValid
			break
		}
	}
	r.commitStatus(logger, ctx, pr, finallyStatus, "", "", nil)
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &opsv1.PipelineRun{}, ".spec.pipelineRef", func(rawObj client.Object) []string {
		pr := rawObj.(*opsv1.PipelineRun)
		return []string{pr.Spec.Ref}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.PipelineRun{}).
		WithEventFilter(
			predicate.Funcs{
				// drop reconcile for status updates
				UpdateFunc: func(e event.UpdateEvent) bool {
					if _, ok := e.ObjectOld.(*opsv1.PipelineRun); !ok {
						return true
					}

					oldObject := e.ObjectOld.(*opsv1.PipelineRun).DeepCopy()
					newObject := e.ObjectNew.(*opsv1.PipelineRun).DeepCopy()

					oldObjectCmp := &opsv1.PipelineRun{}
					newObjectCmp := &opsv1.PipelineRun{}

					oldObjectCmp.Spec = oldObject.Spec
					newObjectCmp.Spec = newObject.Spec

					return !cmp.Equal(oldObjectCmp, newObjectCmp)
				},
			},
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: opsconstants.MaxTaskrunConcurrentReconciles}).
		Complete(r)
}

func (r *PipelineRunReconciler) commitStatus(logger *opslog.Logger, ctx context.Context, pr *opsv1.PipelineRun, prstatus string, taskName, taskRef string, trStatus *opsv1.TaskRunStatus) (err error) {
	latestPr := &opsv1.PipelineRun{}
	for retries := 0; retries < CommitStatusMaxRetries; retries++ {
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.GetNamespace(), Name: pr.GetName()}, latestPr)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		if latestPr.Status.StartTime == nil {
			latestPr.Status.StartTime = &metav1.Time{Time: time.Now()}
		}
		latestPr.Status.RunStatus = prstatus
		if trStatus != nil {
			latestPr.Status.AddPipelineRunTaskStatus(taskName, taskRef, trStatus)
		}
		err = r.Client.Status().Update(ctx, latestPr)
		if err == nil {
			return
		}
		if !apierrors.IsConflict(err) {
			logger.Error.Println(err, "update pipelinerun taskrun status error")
			return
		}
		logger.Info.Println("conflict detected, retrying...", err)
		time.Sleep(1 * time.Second)
	}
	logger.Error.Println("update pipelinerun taskrun status failed after retries", err)
	return
}
