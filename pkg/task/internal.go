/*
Copyright 2023 The KubeSphere Authors.

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

package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	cgcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type taskController struct {
	schema         *runtime.Scheme
	client         ctrlclient.Client
	taskReconciler reconcile.Reconciler

	variableCache cgcache.Store

	wq            workqueue.RateLimitingInterface
	MaxConcurrent int
}

func (c *taskController) AddTasks(ctx context.Context, o AddTaskOptions) error {
	var nsTasks = &kubekeyv1alpha1.TaskList{}

	if err := c.client.List(ctx, nsTasks, ctrlclient.InNamespace(o.Pipeline.Namespace)); err != nil {
		klog.V(4).ErrorS(err, "List tasks error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
		return err
	}
	defer func() {
		// add task to workqueue
		for _, task := range nsTasks.Items {
			c.wq.Add(ctrl.Request{ctrlclient.ObjectKeyFromObject(&task)})
		}
		converter.CalculatePipelineStatus(nsTasks, o.Pipeline)
	}()

	// filter by ownerReference
	for i := len(nsTasks.Items) - 1; i >= 0; i-- {
		var hasOwner bool
		for _, ref := range nsTasks.Items[i].OwnerReferences {
			if ref.UID == o.Pipeline.UID && ref.Kind == "Pipeline" {
				hasOwner = true
			}
		}

		if !hasOwner {
			nsTasks.Items = append(nsTasks.Items[:i], nsTasks.Items[i+1:]...)
		}
	}

	if len(nsTasks.Items) == 0 {
		vars, ok, err := c.variableCache.GetByKey(string(o.Pipeline.UID))
		if err != nil {
			klog.V(4).ErrorS(err, "Get variable from store error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
			return err
		}
		// if tasks has not generated. generate tasks from pipeline
		//vars, ok := cache.LocalVariable.Get(string(o.Pipeline.UID))
		if ok {
			o.variable = vars.(variable.Variable)
		} else {
			nv, err := variable.New(variable.Options{
				Ctx:      ctx,
				Client:   c.client,
				Pipeline: *o.Pipeline,
			})
			if err != nil {
				klog.V(4).ErrorS(err, "Create variable error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
				return err
			}
			if err := c.variableCache.Add(nv); err != nil {
				klog.V(4).ErrorS(err, "Add variable to store error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
				return err
			}
			o.variable = nv
		}

		klog.V(4).InfoS("deal project", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
		projectFs, err := project.New(project.Options{Pipeline: o.Pipeline}).FS(ctx, true)
		if err != nil {
			klog.V(4).ErrorS(err, "Deal project error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
			return err
		}

		// convert to transfer.Playbook struct
		pb, err := converter.MarshalPlaybook(projectFs, o.Pipeline.Spec.Playbook)
		if err != nil {
			return err
		}

		for _, play := range pb.Play {
			if !play.Taggable.IsEnabled(o.Pipeline.Spec.Tags, o.Pipeline.Spec.SkipTags) {
				continue
			}
			// convert Hosts (group or host) to all hosts
			ahn, err := o.variable.Get(variable.Hostnames{Name: play.PlayHost.Hosts})
			if err != nil {
				klog.V(4).ErrorS(err, "Get all host name error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
				return err
			}

			// gather_fact
			if play.GatherFacts {
				for _, h := range ahn.([]string) {
					gfv, err := getGatherFact(ctx, h, o.variable)
					if err != nil {
						klog.V(4).ErrorS(err, "Get gather fact error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "host", h)
						return err
					}
					if err := o.variable.Merge(variable.HostMerge{
						HostNames:   []string{h},
						LocationUID: "",
						Data:        gfv,
					}); err != nil {
						klog.V(4).ErrorS(err, "Merge gather fact error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "host", h)
						return err
					}
				}
			}

			var hs [][]string
			if play.RunOnce {
				// runOnce only run in first node
				hs = [][]string{{ahn.([]string)[0]}}
			} else {
				// group hosts by serial. run the playbook by serial
				hs, err = converter.GroupHostBySerial(ahn.([]string), play.Serial.Data)
				if err != nil {
					klog.V(4).ErrorS(err, "Group host by serial error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
					return err
				}
			}

			// split play by hosts group
			for _, h := range hs {
				puid := uuid.NewString()
				if err := o.variable.Merge(variable.LocationMerge{
					UID:  puid,
					Name: play.Name,
					Type: variable.BlockLocation,
					Vars: play.Vars,
				}); err != nil {
					return err
				}
				if len(h) == 0 {
					err := fmt.Errorf("host is empty")
					klog.V(4).ErrorS(err, "Host is empty", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline))
					return err
				}
				hctx := context.WithValue(ctx, _const.CtxBlockHosts, h)
				// generate task from pre tasks
				preTasks, err := c.createTasks(hctx, o, play.PreTasks, nil, puid, variable.BlockLocation)
				if err != nil {
					klog.V(4).ErrorS(err, "Get pre task from  play error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "play", play.Name)
					return err
				}
				nsTasks.Items = append(nsTasks.Items, preTasks...)
				// generate task from role
				for _, role := range play.Roles {
					ruid := uuid.NewString()
					if err := o.variable.Merge(variable.LocationMerge{
						ParentUID: puid,
						UID:       ruid,
						Name:      play.Name,
						Type:      variable.BlockLocation,
						Vars:      role.Vars,
					}); err != nil {
						return err
					}
					roleTasks, err := c.createTasks(context.WithValue(hctx, _const.CtxBlockRole, role.Role), o, role.Block, role.When.Data, ruid, variable.BlockLocation)
					if err != nil {
						klog.V(4).ErrorS(err, "Get role task from  play error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "play", play.Name, "role", role.Role)
						return err
					}
					nsTasks.Items = append(nsTasks.Items, roleTasks...)
				}
				// generate task from tasks
				tasks, err := c.createTasks(hctx, o, play.Tasks, nil, puid, variable.BlockLocation)
				if err != nil {
					klog.V(4).ErrorS(err, "Get task from  play error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "play", play.Name)
					return err
				}
				nsTasks.Items = append(nsTasks.Items, tasks...)
				// generate task from post tasks
				postTasks, err := c.createTasks(hctx, o, play.Tasks, nil, puid, variable.BlockLocation)
				if err != nil {
					klog.V(4).ErrorS(err, "Get post task from  play error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "play", play.Name)
					return err
				}
				nsTasks.Items = append(nsTasks.Items, postTasks...)
			}
		}
	}

	return nil
}

// createTasks convert ansible block to task
func (k *taskController) createTasks(ctx context.Context, o AddTaskOptions, ats []kkcorev1.Block, when []string, puid string, locationType variable.LocationType) ([]kubekeyv1alpha1.Task, error) {
	var tasks []kubekeyv1alpha1.Task
	for _, at := range ats {
		if !at.Taggable.IsEnabled(o.Pipeline.Spec.Tags, o.Pipeline.Spec.SkipTags) {
			continue
		}

		uid := uuid.NewString()
		atWhen := append(when, at.When.Data...)
		if len(at.Block) != 0 {
			// add block
			block, err := k.createTasks(ctx, o, at.Block, atWhen, uid, variable.BlockLocation)
			if err != nil {
				klog.V(4).ErrorS(err, "Get block task from block error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name)
				return nil, err
			}
			tasks = append(tasks, block...)

			if len(at.Always) != 0 {
				always, err := k.createTasks(ctx, o, at.Always, atWhen, uid, variable.AlwaysLocation)
				if err != nil {
					klog.V(4).ErrorS(err, "Get always task from block error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name)
					return nil, err
				}
				tasks = append(tasks, always...)
			}
			if len(at.Rescue) != 0 {
				rescue, err := k.createTasks(ctx, o, at.Rescue, atWhen, uid, variable.RescueLocation)
				if err != nil {
					klog.V(4).ErrorS(err, "Get rescue task from block error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name)
					return nil, err
				}
				tasks = append(tasks, rescue...)
			}
		} else {
			task := converter.MarshalBlock(context.WithValue(ctx, _const.CtxBlockWhen, atWhen), at)
			// complete by pipeline
			task.GenerateName = o.Pipeline.Name + "-"
			task.Namespace = o.Pipeline.Namespace
			if err := controllerutil.SetControllerReference(o.Pipeline, task, k.schema); err != nil {
				klog.V(4).ErrorS(err, "Set controller reference error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name)
				return nil, err
			}
			// complete module by unknown field
			for n, a := range at.UnknownFiled {
				data, err := json.Marshal(a)
				if err != nil {
					klog.V(4).ErrorS(err, "Marshal unknown field error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name, "field", n)
					return nil, err
				}
				if m := modules.FindModule(n); m != nil {
					task.Spec.Module.Name = n
					task.Spec.Module.Args = runtime.RawExtension{Raw: data}
					break
				}
			}
			if task.Spec.Module.Name == "" { // action is necessary for a task
				klog.V(4).ErrorS(nil, "No module/action detected in task", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name)
				return nil, fmt.Errorf("no module/action detected in task: %s", task.Name)
			}
			// create task
			if err := k.client.Create(ctx, task); err != nil {
				klog.V(4).ErrorS(err, "Create task error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name)
				return nil, err
			}
			uid = string(task.UID)
			tasks = append(tasks, *task)
		}
		// add location to variable
		if err := o.variable.Merge(variable.LocationMerge{
			UID:       uid,
			ParentUID: puid,
			Type:      locationType,
			Name:      at.Name,
			Vars:      at.Vars,
		}); err != nil {
			klog.V(4).ErrorS(err, "Merge block to variable error", "pipeline", ctrlclient.ObjectKeyFromObject(o.Pipeline), "block", at.Name)
			return nil, err
		}
	}
	return tasks, nil
}

// Start task controller, deal task in work queue
func (k *taskController) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		k.wq.ShutDown()
	}()
	// deal work queue
	wg := &sync.WaitGroup{}
	for i := 0; i < k.MaxConcurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k.processNextWorkItem(ctx) {
			}
		}()
	}
	<-ctx.Done()
	wg.Wait()
	return nil
}

func (k *taskController) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := k.wq.Get()
	if shutdown {
		return false
	}

	defer k.wq.Done(obj)

	req, ok := obj.(ctrl.Request)
	if !ok {
		// As the item in the workqueue is actually invalid, we call
		// Forget here else we'd go into a loop of attempting to
		// process a work item that is invalid.
		k.wq.Forget(obj)
		klog.V(4).ErrorS(nil, "Queue item was not a Request", "request", req)
		// Return true, don't take a break
		return true
	}

	result, err := k.taskReconciler.Reconcile(ctx, req)
	switch {
	case err != nil:
		k.wq.AddRateLimited(req)
		klog.V(4).ErrorS(err, "Reconciler error", "request", req)
	case result.RequeueAfter > 0:
		// The result.RequeueAfter request will be lost, if it is returned
		// along with a non-nil error. But this is intended as
		// We need to drive to stable reconcile loops before queuing due
		// to result.RequestAfter
		k.wq.Forget(obj)
		k.wq.AddAfter(req, result.RequeueAfter)
	case result.Requeue:
		k.wq.AddRateLimited(req)
	default:
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		k.wq.Forget(obj)
	}
	return true
}

func (k *taskController) NeedLeaderElection() bool {
	return true
}
