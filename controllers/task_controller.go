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

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cron "github.com/robfig/cron/v3"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/kube"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/task"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	Client     client.Client
	Scheme     *runtime.Scheme
	crontabMap map[string]cron.EntryID
	cron       *cron.Cron
}

//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=tasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=tasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.chenshaowen.com,resources=tasks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Task object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if r.crontabMap == nil {
		r.crontabMap = make(map[string]cron.EntryID)
	}
	if r.cron == nil {
		r.cron = cron.New()
		r.cron.Start()
	}

	t := &opsv1.Task{}
	err := r.Client.Get(ctx, req.NamespacedName, t)

	//if delete, stop ticker
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, r.deleteTask(ctx, req.NamespacedName)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	// validate task
	if t.Spec.RuntimeImage == "" {
		t.Spec.RuntimeImage = constants.DefaultRuntimeImage
	}

	// create task
	err = r.createTask(ctx, t)
	if err != nil {
		log.Info(err.Error())
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.Task{}).
		WithEventFilter(
			predicate.Funcs{
				// drop reconcile for status updates
				UpdateFunc: func(e event.UpdateEvent) bool {
					if _, ok := e.ObjectOld.(*opsv1.Task); !ok {
						return true
					}

					oldObject := e.ObjectOld.(*opsv1.Task).DeepCopy()
					newObject := e.ObjectNew.(*opsv1.Task).DeepCopy()

					oldObject.Status = opsv1.TaskStatus{}
					newObject.Status = opsv1.TaskStatus{}

					oldObject.ObjectMeta.ResourceVersion = ""
					newObject.ObjectMeta.ResourceVersion = ""

					return !cmp.Equal(oldObject, newObject)
				},
			},
		).
		Complete(r)
}

func (r *TaskReconciler) deleteTask(ctx context.Context, namespacedName types.NamespacedName) error {
	_, ok := r.crontabMap[namespacedName.String()]
	if ok {
		r.cron.Remove(r.crontabMap[namespacedName.String()])
		delete(r.crontabMap, namespacedName.String())
	}
	return nil
}

func (r *TaskReconciler) createTask(ctx context.Context, t *opsv1.Task) (err error) {
	_, ok := r.crontabMap[t.GetUniqueKey()]
	if ok {
		r.cron.Remove(r.crontabMap[t.GetUniqueKey()])
	}
	typeRef := t.GetSpec().TypeRef
	if typeRef == "host" {
		hostCmd := func() {
			h := &opsv1.Host{}
			err := r.Client.Get(ctx, types.NamespacedName{Namespace: t.GetNamespace(), Name: t.GetSpec().NameRef}, h)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			r.runTaskOnHost(ctx, t, h)
		}
		if t.GetSpec().Crontab != "" {
			r.crontabMap[t.GetUniqueKey()], err = r.cron.AddFunc(t.GetSpec().Crontab, hostCmd)
			if err != nil {
				return err
			}
		} else {
			hostCmd()
		}
	} else if typeRef == "cluster" {
		clusterCmd := func() {
			c := &opsv1.Cluster{}
			err := r.Client.Get(ctx, types.NamespacedName{Namespace: t.GetNamespace(), Name: t.GetSpec().NameRef}, c)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			r.runTaskOnKube(ctx, t, c, t.Spec.NodeName)
		}
		if t.GetSpec().Crontab != "" {
			r.crontabMap[t.GetUniqueKey()], err = r.cron.AddFunc(t.GetSpec().Crontab, clusterCmd)
			if err != nil {
				return err
			}
		} else {
			clusterCmd()
		}
	}
	r.Client.Status().Update(ctx, t)
	return
}

func (r *TaskReconciler) runTaskOnHost(ctx context.Context, t *opsv1.Task, h *opsv1.Host) (err error) {
	hc, err := host.NewHostConnectionBase64(
		h.Spec.Address,
		h.Spec.Port,
		h.Spec.Username,
		h.Spec.Password,
		h.Spec.PrivateKey,
		h.Spec.PrivateKeyPath,
	)
	if err != nil {
		return err
	}
	t.Status.NewOutput()
	t, err = task.RunTaskOnHost(t, hc, option.TaskOption{})
	return
}

func (r *TaskReconciler) runTaskOnKube(ctx context.Context, t *opsv1.Task, c *opsv1.Cluster, nodeName string) (err error) {
	kc, err := kube.NewClusterConnection(c)
	if err != nil {
		return err
	}
	nodes, err := kc.GetNodeByName(nodeName)
	if err != nil || len(nodes.Items) == 0 {
		return err
	}
	t.Status.NewOutput()
	t, err = task.RunTaskOnKube(t, kc, &nodes.Items[0], option.TaskOption{}, option.KubeOption{RuntimeImage: t.Spec.RuntimeImage})
	return
}
