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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	cron "github.com/robfig/cron/v3"
	crdv1 "github.com/shaowenchen/ops/api/v1"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	opsconstants "github.com/shaowenchen/ops/pkg/constants"
	opsevent "github.com/shaowenchen/ops/pkg/event"
	opskube "github.com/shaowenchen/ops/pkg/kube"
	opslog "github.com/shaowenchen/ops/pkg/log"
	opsmetrics "github.com/shaowenchen/ops/pkg/metrics"
	opstask "github.com/shaowenchen/ops/pkg/task"
	opsutils "github.com/shaowenchen/ops/pkg/utils"
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
	Scheme          *runtime.Scheme
	crontabMap      map[string]cron.EntryID
	crontabMapMutex sync.RWMutex
	cron            *cron.Cron
	clearCron       *cron.Cron
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
func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	controllerName := "PipelineRun"

	// Record metrics
	defer func() {
		resultStr := "success"
		if err != nil {
			resultStr = "error"
			opsmetrics.RecordReconcileError(controllerName, req.Namespace, "reconcile_error")
		}
		opsmetrics.RecordReconcile(controllerName, req.Namespace, resultStr)
	}()

	// start clear cron
	r.registerClearCron()
	// only reconcile active namespace
	actionNs := opsconstants.GetEnvActiveNamespace()
	if actionNs != "" && actionNs != req.Namespace {
		return ctrl.Result{}, nil
	}
	logger := opslog.NewLogger().SetStd().SetFlag().Build()
	if opsconstants.GetEnvDebug() {
		logger.SetVerbose("debug").Build()
	}
	if r.crontabMap == nil {
		r.crontabMap = make(map[string]cron.EntryID)
	}
	if r.cron == nil {
		r.cron = cron.New(cron.WithLocation(time.Local))
		r.cron.Start()
	}

	pr := &opsv1.PipelineRun{}
	err = r.Client.Get(ctx, req.NamespacedName, pr)

	if apierrors.IsNotFound(err) {
		// Record PipelineRun info metrics as deleted (value=0)
		opsmetrics.RecordPipelineRunInfo(req.Namespace, req.Name, "", "", 0)
		r.deleteCronTab(logger, ctx, req.NamespacedName)
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	if opsconstants.IsFinishedStatus(pr.Status.RunStatus) {
		return ctrl.Result{}, nil
	}
	// insert env
	pr.SetEnv()
	// if is others cluster, send and just sync status
	cluster := r.isOtherCluster(pr)
	if cluster != nil {
		// send pr
		logger.Info.Printf("send pipelinerun %s to cluster %s", pr.Name, cluster.Name)
		kc, err := opskube.NewClusterConnection(cluster)
		if err != nil {
			logger.Error.Println(err, "failed to create cluster connection")
			return ctrl.Result{}, err
		}
		pr.SetCurrentCluster()
		err = kc.CreatePipelineRun(pr)
		if err != nil {
			logger.Error.Println(err, "failed to create pr")
			return ctrl.Result{}, err
		}
		r.commitStatus(logger, ctx, pr, opsconstants.StatusDispatched, "", "", nil)
		MaxTimes := 60 / 3 * 30
		go func() {
			for retries := 0; retries < MaxTimes; retries++ {
				time.Sleep(3 * time.Second)
				logger.Info.Printf("check pipelinerun %s status, times %d", pr.Name, retries)
				othersPr := &opsv1.PipelineRun{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: pr.Namespace,
						Name:      pr.Name,
					},
				}
				err = kc.GetPipelineRun(othersPr)
				if err != nil {
					logger.Error.Println(err, "failed to get others pr")
					return
				} else if !opsconstants.IsFinishedStatus(othersPr.Status.RunStatus) {
					continue
				}
				currentPR := &opsv1.PipelineRun{}
				err = r.Client.Get(ctx, req.NamespacedName, currentPR)
				if err != nil {
					logger.Error.Println(err, "failed to get current pr")
					return
				}
				currentPR.Status = othersPr.Status
				err = r.Client.Status().Update(ctx, currentPR)
				if err != nil {
					logger.Error.Println(err, "failed to update current pr")
				}
				return
			}
			// send event
			go opsevent.FactoryPipelineRun(pr.Namespace, pr.Name, opsconstants.Status).Publish(ctx, &opsevent.EventPipelineRun{
				PipelineRef:       pr.Spec.PipelineRef,
				Desc:              pr.Spec.Desc,
				Variables:         pr.Spec.Variables,
				PipelineRunStatus: pr.Status,
			})
		}()
		return ctrl.Result{}, nil
	}
	// else is this cluster
	// add crontab
	// if has crontab, start timer and set status to Successed (don't run)
	if pr.Spec.Crontab != "" {
		r.addCronTab(logger, ctx, pr)
		// scheduled pipelinerun should always be Successed after timer started
		r.commitStatus(logger, ctx, pr, opsconstants.StatusSuccessed, "", "", nil)
		return ctrl.Result{}, nil
	}
	// had run once, skip
	if !(pr.Status.RunStatus == opsconstants.StatusEmpty || pr.Status.RunStatus == opsconstants.StatusRunning) {
		return ctrl.Result{}, nil
	}
	// get pipeline
	p := &opsv1.Pipeline{}
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.Namespace, Name: pr.Spec.PipelineRef}, p)
	if err != nil {
		r.commitStatus(logger, ctx, pr, opsconstants.StatusFailed, "", "", nil)
		return ctrl.Result{}, err
	}

	// run pipeline (no crontab)
	err = r.run(logger, ctx, p, pr)
	if err != nil {
		return ctrl.Result{}, err
	}
	// send event
	err = opsevent.FactoryPipelineRun(pr.Namespace, pr.Name).Publish(ctx, pr)
	return ctrl.Result{}, err
}

func (r *PipelineRunReconciler) isOtherCluster(pr *opsv1.PipelineRun) *opsv1.Cluster {
	cluster := pr.GetCluster()
	if cluster == "" {
		return nil
	}
	// judge cluster uid
	currentUID, err := opsutils.GetClusterUID(r.Client)
	if err != nil {
		return nil
	}
	clusterList := &opsv1.ClusterList{}
	err = r.Client.List(context.TODO(), clusterList)
	if err != nil {
		return nil
	}
	for _, item := range clusterList.Items {
		if currentUID != item.Status.UID && item.Name == cluster {
			return &item
		}
	}
	return nil
}

func (r *PipelineRunReconciler) deleteCronTab(logger *opslog.Logger, ctx context.Context, namespacedName types.NamespacedName) error {
	r.crontabMapMutex.Lock()
	defer r.crontabMapMutex.Unlock()

	entryID, ok := r.crontabMap[namespacedName.String()]
	if ok {
		r.cron.Remove(entryID)
		delete(r.crontabMap, namespacedName.String())
		logger.Info.Println(fmt.Sprintf("clear ticker for pipelinerun %s", namespacedName.String()))
	}
	return nil
}

func (r *PipelineRunReconciler) addCronTab(logger *opslog.Logger, ctx context.Context, objRun *opsv1.PipelineRun) {
	key := objRun.GetUniqueKey()

	// if crontab is empty, remove existing cron if any
	if objRun.Spec.Crontab == "" {
		r.crontabMapMutex.Lock()
		entryID, ok := r.crontabMap[key]
		if ok {
			r.cron.Remove(entryID)
			delete(r.crontabMap, key)
			logger.Info.Println(fmt.Sprintf("clear ticker for pipelinerun %s (crontab is empty)", key))
		}
		r.crontabMapMutex.Unlock()
		return
	}

	// check if cron already exists
	r.crontabMapMutex.Lock()
	oldEntryID, exists := r.crontabMap[key]
	if exists {
		// remove old cron entry if exists
		r.cron.Remove(oldEntryID)
		delete(r.crontabMap, key)
		logger.Info.Println(fmt.Sprintf("remove old ticker for pipelinerun %s (crontab updated)", key))
	}
	r.crontabMapMutex.Unlock()

	logger.Info.Println(fmt.Sprintf("add ticker for pipelinerun %s with crontab %s", key, objRun.Spec.Crontab))

	id, err := r.cron.AddFunc(objRun.Spec.Crontab, func() {
		time.Sleep(time.Duration(rand.Intn(opsconstants.SyncCronRandomBias)) * time.Second)
		// get the scheduled PipelineRun to verify it still exists
		scheduledPr := &opsv1.PipelineRun{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: objRun.Namespace, Name: objRun.Name}, scheduledPr)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		// verify it still has crontab
		if scheduledPr.Spec.Crontab == "" {
			logger.Info.Println(fmt.Sprintf("skip pipelinerun %s: crontab removed", scheduledPr.GetUniqueKey()))
			return
		}
		// check if there's already a running instance created by this scheduled pipelinerun
		pipelineRunList := &opsv1.PipelineRunList{}
		labelSelector := client.MatchingLabels{
			opsconstants.LabelScheduledByKey:   scheduledPr.Name,
			opsconstants.LabelScheduledKindKey: opsconstants.PipelineRun,
		}
		err = r.Client.List(ctx, pipelineRunList, client.InNamespace(scheduledPr.Namespace), labelSelector)
		if err != nil {
			logger.Error.Println(fmt.Sprintf("failed to list pipelineruns: %v", err))
			return
		}
		for _, pr := range pipelineRunList.Items {
			// check if this pipelinerun is running
			if pr.Status.RunStatus == opsconstants.StatusRunning || pr.Status.RunStatus == opsconstants.StatusEmpty {
				logger.Info.Println(fmt.Sprintf("skip pipelinerun %s: already has a running instance %s", scheduledPr.GetUniqueKey(), pr.GetUniqueKey()))
				return
			}
		}
		logger.Info.Println(fmt.Sprintf("cron triggered for pipelinerun %s, creating new execution instance", scheduledPr.GetUniqueKey()))
		// get pipeline
		p := &opsv1.Pipeline{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: scheduledPr.Namespace, Name: scheduledPr.Spec.PipelineRef}, p)
		if err != nil {
			logger.Error.Println(err)
			return
		}
		// create a new PipelineRun without crontab for execution
		newPr := opsv1.NewPipelineRun(p)
		newPr.Spec.Crontab = "" // ensure no crontab
		// copy variables from scheduled pipelinerun (deep copy)
		if scheduledPr.Spec.Variables != nil {
			newPr.Spec.Variables = make(map[string]string)
			for k, v := range scheduledPr.Spec.Variables {
				newPr.Spec.Variables[k] = v
			}
		}
		newPr.Spec.Desc = scheduledPr.Spec.Desc // copy desc
		// ensure labels map exists (NewPipelineRun already creates it with LabelPipelineRefKey)
		if newPr.Labels == nil {
			newPr.Labels = make(map[string]string)
		}
		// ensure pipelineref label is set
		newPr.Labels[opsconstants.LabelPipelineRefKey] = p.Name
		// set labels to identify this is created by the scheduled pipelinerun
		newPr.Labels[opsconstants.LabelScheduledByKey] = scheduledPr.Name
		newPr.Labels[opsconstants.LabelScheduledKindKey] = opsconstants.PipelineRun
		// set owner reference to the scheduled pipelinerun
		if scheduledPr.UID != "" {
			newPr.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: opsconstants.APIVersion,
					Kind:       opsconstants.PipelineRun,
					Name:       scheduledPr.Name,
					UID:        scheduledPr.UID,
				},
			}
		}
		// create the new pipelinerun
		err = r.Client.Create(ctx, newPr)
		if err != nil {
			logger.Error.Println(fmt.Sprintf("failed to create new pipelinerun for scheduled pipelinerun %s: %v", scheduledPr.GetUniqueKey(), err))
			return
		}
		logger.Info.Println(fmt.Sprintf("created new pipelinerun %s for scheduled pipelinerun %s", newPr.GetUniqueKey(), scheduledPr.GetUniqueKey()))
	})
	if err != nil {
		logger.Error.Println(err)
		return
	}

	r.crontabMapMutex.Lock()
	r.crontabMap[key] = id
	r.crontabMapMutex.Unlock()
}

func (r *PipelineRunReconciler) registerClearCron() {
	if r.clearCron != nil {
		return
	}
	r.clearCron = cron.New(cron.WithLocation(time.Local))
	r.clearCron.AddFunc(opsconstants.ClearCronTab, func() {
		objs := &opsv1.PipelineRunList{}
		err := r.Client.List(context.Background(), objs)
		if err != nil {
			return
		}
		for _, obj := range objs.Items {
			if obj.Status.RunStatus == opsconstants.StatusRunning || obj.Status.RunStatus == opsconstants.StatusEmpty {
				continue
			}
			if obj.GetObjectMeta().GetCreationTimestamp().Add(opsconstants.DefaultTTLSecondsAfterFinished * time.Second).After(time.Now()) {
				continue
			}
			r.Client.Delete(context.Background(), &obj)
		}
	})
	r.clearCron.Start()
}

// buildTaskRunVariables builds variables for TaskRun by filtering pipeline variables
// based on task requirements and merging TaskRef variables
func (r *PipelineRunReconciler) buildTaskRunVariables(pr *opsv1.PipelineRun, t *opsv1.Task, tRef opsv1.TaskRef, ctx context.Context) map[string]string {
	requiredVars := opstask.GetTaskRequiredVariables(t)
	taskResults := r.getTaskResults(pr, ctx)
	vars := make(map[string]string)

	// Filter pipeline variables to only include what task needs
	for k, v := range pr.Spec.Variables {
		if requiredVars[k] {
			vars[k] = opstask.RenderStringWithPathRefs(v, pr.Spec.Variables, taskResults)
		}
	}

	return vars
}

// getTaskResults extracts task results from PipelineRun status
func (r *PipelineRunReconciler) getTaskResults(pr *opsv1.PipelineRun, ctx context.Context) map[string]map[string]string {
	latestPr := &opsv1.PipelineRun{}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: pr.Namespace, Name: pr.Name}, latestPr); err != nil {
		return nil
	}

	taskResults := make(map[string]map[string]string)
	for _, taskStatus := range latestPr.Status.PipelineRunStatus {
		if len(taskStatus.Results) > 0 {
			taskResults[taskStatus.TaskName] = taskStatus.Results
		}
	}
	return taskResults
}

func (r *PipelineRunReconciler) run(logger *opslog.Logger, ctx context.Context, p *opsv1.Pipeline, pr *opsv1.PipelineRun) (err error) {
	runAlways := false
	latestTrOuput := ""
	for _, tRef := range p.Spec.Tasks {
		if runAlways && !tRef.RunAlways {
			continue
		}
		// create taskrun
		t := &opsv1.Task{}
		err = r.Client.Get(ctx, types.NamespacedName{Namespace: pr.Namespace, Name: tRef.TaskRef}, t)
		if err != nil {
			logger.Error.Println(err)
			runAlways = true
			r.commitStatus(logger, ctx, pr, opsconstants.StatusDataInValid, tRef.Name, tRef.TaskRef, &opsv1.TaskRunStatus{
				RunStatus: opsconstants.StatusDataInValid,
			})
			continue
		}
		// patch latest tr ouput var:value to variables (backward compatibility)
		// Todo: support multi vars
		if latestTrOuput != "" {
			latestTrOuputArr := strings.Split(latestTrOuput, ":")
			if len(latestTrOuputArr) == 2 {
				key := strings.TrimSpace(latestTrOuputArr[0])
				value := strings.TrimSpace(latestTrOuputArr[1])
				pr.Spec.Variables[key] = value
			}
		}

		// Build variables for TaskRun: filter pipeline variables by task requirements and merge TaskRef variables
		tr := opsv1.NewTaskRunWithPipelineRun(pr, t, tRef, p)
		tr.Spec.Variables = r.buildTaskRunVariables(pr, t, tRef, ctx)
		err = r.Client.Create(ctx, tr)
		if err != nil {
			logger.Error.Println(err)
			runAlways = true
			r.commitStatus(logger, ctx, pr, opsconstants.StatusDataInValid, tRef.Name, tRef.TaskRef, nil)
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
			r.commitStatus(logger, ctx, pr, opsconstants.StatusRunning, tRef.Name, trRunning.Spec.TaskRef, &trRunning.Status)
			if trRunning.Status.RunStatus == opsconstants.StatusRunning || trRunning.Status.RunStatus == opsconstants.StatusEmpty {
				continue
			} else if trRunning.Status.RunStatus == opsconstants.StatusSuccessed {
				// Extract and store task results
				if tRef.Results != nil && len(tRef.Results) > 0 {
					taskResults := make(map[string]string)
					// Extract results from TaskRunStatus based on TaskRef.Results definition
					// TaskRef.Results is map[resultKey]stepName
					if len(trRunning.Status.TaskRunNodeStatus) == 1 {
						for _, nodeStatus := range trRunning.Status.TaskRunNodeStatus {
							// Build a map of stepName -> stepOutput for quick lookup
							stepOutputs := make(map[string]string)
							for _, step := range nodeStatus.TaskRunStep {
								stepOutputs[step.StepName] = step.StepOutput
							}
							// Extract results according to TaskRef.Results
							for resultKey, stepName := range tRef.Results {
								if output, ok := stepOutputs[stepName]; ok {
									// Try to extract result using special markers
									extractedValue := extractResultFromOutput(output, resultKey)
									if extractedValue != "" {
										taskResults[resultKey] = extractedValue
									} else {
										// Fallback to full output (backward compatibility)
										taskResults[resultKey] = strings.TrimSpace(output)
									}
								}
							}
						}
					}
					// Store results in PipelineRunTaskStatus
					if len(taskResults) > 0 {
						r.updateTaskResults(logger, ctx, pr, tRef.Name, taskResults)
						// Also add results to PipelineRun variables for direct reference
						if pr.Spec.Variables == nil {
							pr.Spec.Variables = make(map[string]string)
						}
						for k, v := range taskResults {
							pr.Spec.Variables[k] = v
						}
					}
				}
				// patch latest tr ouput var:value to variables (backward compatibility)
				// This maintains the existing key:value mechanism
				if len(trRunning.Status.TaskRunNodeStatus) == 1 {
					for _, nodeStatus := range trRunning.Status.TaskRunNodeStatus {
						if len(nodeStatus.TaskRunStep) > 0 {
							latestTrOuput = nodeStatus.TaskRunStep[len(nodeStatus.TaskRunStep)-1].StepOutput
						}
					}
				}
				break
			} else {
				runAlways = true
				break
			}
		}
	}
	finallyStatus := opsconstants.StatusSuccessed
	for _, status := range pr.Status.PipelineRunStatus {
		if status.TaskRunStatus.RunStatus == opsconstants.StatusFailed {
			finallyStatus = opsconstants.StatusFailed
			break
		} else if status.TaskRunStatus.RunStatus == opsconstants.StatusDataInValid {
			finallyStatus = opsconstants.StatusDataInValid
			break
		}
	}
	r.commitStatus(logger, ctx, pr, finallyStatus, "", "", nil)
	// push event
	go opsevent.FactoryPipelineRun(pr.Namespace, pr.Name, opsconstants.Status).Publish(ctx, opsevent.EventPipelineRun{
		PipelineRef:       pr.Spec.PipelineRef,
		Desc:              pr.Spec.Desc,
		Variables:         pr.Spec.Variables,
		PipelineRunStatus: pr.Status,
	})
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// push event
	namespace, err := opsconstants.GetCurrentNamespace()
	if err == nil {
		go opsevent.FactoryController(namespace, opsconstants.PipelineRuns, opsconstants.Setup).Publish(context.TODO(), opsevent.EventController{
			Kind: opsconstants.PipelineRuns,
		})
	}
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &opsv1.PipelineRun{}, ".spec.pipelineRef", func(rawObj client.Object) []string {
		pr := rawObj.(*opsv1.PipelineRun)
		return []string{pr.Spec.PipelineRef}
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
	oldStatus := pr.Status.RunStatus
	for retries := 0; retries < CommitStatusMaxRetries; retries++ {
		latestPr := &opsv1.PipelineRun{}
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
			// Record CRD resource status change metrics - record every status change (value=1 for existing resource)
			if oldStatus != latestPr.Status.RunStatus {
				// Record PipelineRun info metrics (static fields only)
				opsmetrics.RecordPipelineRunInfo(latestPr.Namespace, latestPr.Name, latestPr.Spec.PipelineRef, latestPr.Spec.Crontab, 1)
				// Record PipelineRun status metrics (dynamic fields)
				opsmetrics.RecordPipelineRunStatus(latestPr.Namespace, latestPr.Name, latestPr.Status.RunStatus)
				// Record scheduled task status change if this is a scheduled task (has Crontab)
				if latestPr.Spec.Crontab != "" {
					opsmetrics.RecordPipelineRunInfo(latestPr.Namespace, latestPr.Name, latestPr.Spec.PipelineRef, latestPr.Spec.Crontab, 1)
					opsmetrics.RecordPipelineRunStatus(latestPr.Namespace, latestPr.Name, latestPr.Status.RunStatus)
				}
				// Record PipelineRef status phase change (decrement old status, increment new status)
				if latestPr.Spec.PipelineRef != "" {
					opsmetrics.RecordPipelineRunStatusPhase(latestPr.Namespace, latestPr.Spec.PipelineRef, oldStatus, latestPr.Status.RunStatus)
				}
			}
			return
		}
		if !apierrors.IsConflict(err) {
			logger.Error.Println(err, "update pipelinerun taskrun status error")
			return
		}
		logger.Info.Println("try commit times ", retries+1, "conflict detected, retrying...", err)
		time.Sleep(3 * time.Second)
	}
	logger.Error.Println("update pipelinerun taskrun status failed after retries", err)
	return
}

// updateTaskResults updates the results for a specific task in PipelineRunStatus
func (r *PipelineRunReconciler) updateTaskResults(logger *opslog.Logger, ctx context.Context, pr *opsv1.PipelineRun, taskName string, results map[string]string) error {
	for retries := 0; retries < CommitStatusMaxRetries; retries++ {
		latestPr := &opsv1.PipelineRun{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: pr.GetNamespace(), Name: pr.GetName()}, latestPr)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		// Find and update the task status
		found := false
		for i, taskStatus := range latestPr.Status.PipelineRunStatus {
			if taskStatus.TaskName == taskName {
				if taskStatus.Results == nil {
					taskStatus.Results = make(map[string]string)
				}
				for k, v := range results {
					taskStatus.Results[k] = v
				}
				latestPr.Status.PipelineRunStatus[i] = taskStatus
				found = true
				break
			}
		}
		if !found {
			// Task status not found, create a new one
			latestPr.Status.PipelineRunStatus = append(latestPr.Status.PipelineRunStatus, opsv1.PipelineRunTaskStatus{
				TaskName: taskName,
				Results:  results,
			})
		}
		err = r.Client.Status().Update(ctx, latestPr)
		if err == nil {
			return nil
		}
		if !apierrors.IsConflict(err) {
			logger.Error.Println(err, "update pipelinerun task results error")
			return err
		}
		logger.Info.Println("try update task results times ", retries+1, "conflict detected, retrying...", err)
		time.Sleep(3 * time.Second)
	}
	logger.Error.Println("update pipelinerun task results failed after retries")
	return fmt.Errorf("update pipelinerun task results failed after retries")
}

// extractResultFromOutput extracts result value from step output using special markers
// Supports multiple formats:
// 1. OPS_RESULT:key=value
// 2. OPS_RESULT:{"key":"value"} (JSON format)
// 3. OPS_RESULT:key:value (alternative format)
// Returns empty string if no marker found (fallback to full output)
func extractResultFromOutput(output, resultKey string) string {
	if output == "" || resultKey == "" {
		return ""
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Format 1: OPS_RESULT:key=value
		if strings.HasPrefix(line, "OPS_RESULT:") {
			content := strings.TrimPrefix(line, "OPS_RESULT:")

			// Try key=value format
			if strings.Contains(content, "=") {
				parts := strings.SplitN(content, "=", 2)
				if len(parts) == 2 && strings.TrimSpace(parts[0]) == resultKey {
					return strings.TrimSpace(parts[1])
				}
			}

			// Try key:value format
			if strings.Contains(content, ":") && !strings.HasPrefix(content, "{") {
				parts := strings.SplitN(content, ":", 2)
				if len(parts) == 2 && strings.TrimSpace(parts[0]) == resultKey {
					return strings.TrimSpace(parts[1])
				}
			}

			// Try JSON format: OPS_RESULT:{"key":"value"}
			if strings.HasPrefix(content, "{") {
				// Simple JSON parsing for single key-value pair
				jsonStr := content
				keyPattern := fmt.Sprintf(`"%s"\s*:\s*"([^"]+)"`, resultKey)
				if matched := regexp.MustCompile(keyPattern).FindStringSubmatch(jsonStr); len(matched) > 1 {
					return matched[1]
				}
			}
		}
	}

	return ""
}
