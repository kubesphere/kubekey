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
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/cache"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type taskController struct {
	client         ctrlclient.Client
	taskReconciler reconcile.Reconciler

	wq            workqueue.RateLimitingInterface
	MaxConcurrent int
}

func (c *taskController) AddTasks(ctx context.Context, o AddTaskOptions) error {
	var nsTasks = &kubekeyv1alpha1.TaskList{}

	if err := c.client.List(ctx, nsTasks, ctrlclient.InNamespace(o.Pipeline.Namespace)); err != nil {
		klog.Errorf("[Pipeline %s] list tasks error: %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), err)
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
		// if tasks has not generated. generate tasks from pipeline
		vars, ok := cache.LocalVariable.Get(string(o.Pipeline.UID))
		if ok {
			o.variable = vars.(variable.Variable)
		} else {
			newVars, err := variable.New(variable.Options{
				Ctx:      ctx,
				Client:   c.client,
				Pipeline: *o.Pipeline,
			})
			if err != nil {
				klog.Errorf("[Pipeline %s] create variable failed: %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), err)
				return err
			}
			cache.LocalVariable.Put(string(o.Pipeline.UID), newVars)
			o.variable = newVars
		}

		klog.V(4).Infof("[Pipeline %s] deal project", ctrlclient.ObjectKeyFromObject(o.Pipeline))
		projectFs, err := project.New(project.Options{Pipeline: o.Pipeline}).FS(ctx, true)
		if err != nil {
			klog.Errorf("[Pipeline %s] deal project error: %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), err)
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
				klog.Errorf("[Pipeline %s] get all host name error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), err)
				return err
			}

			// gather_fact
			if play.GatherFacts {
				for _, h := range ahn.([]string) {
					gfv, err := getGatherFact(ctx, h, o.variable)
					if err != nil {
						klog.Errorf("[Pipeline %s] get gather fact from host %s error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), h, err)
						return err
					}
					if err := o.variable.Merge(variable.HostMerge{
						HostNames:   []string{h},
						LocationUID: "",
						Data:        gfv,
					}); err != nil {
						klog.Errorf("[Pipeline %s] merge gather fact from host %s error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), h, err)
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
					klog.Errorf("[Pipeline %s] convert host by serial error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), err)
					return err
				}
			}

			// split play by hosts group
			for _, h := range hs {
				puid := uuid.NewString()
				if err := o.variable.Merge(variable.LocationMerge{
					Uid:  puid,
					Name: play.Name,
					Type: variable.BlockLocation,
					Vars: play.Vars,
				}); err != nil {
					return err
				}
				hctx := context.WithValue(ctx, _const.CtxBlockHosts, h)
				// generate task from pre tasks
				preTasks, err := c.block2Task(hctx, o, play.PreTasks, nil, puid, variable.BlockLocation)
				if err != nil {
					klog.Errorf("[Pipeline %s] get pre task from  play %s error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), play.Name, err)
					return err
				}
				nsTasks.Items = append(nsTasks.Items, preTasks...)
				// generate task from role
				for _, role := range play.Roles {
					ruid := uuid.NewString()
					if err := o.variable.Merge(variable.LocationMerge{
						ParentID: puid,
						Uid:      ruid,
						Name:     play.Name,
						Type:     variable.BlockLocation,
						Vars:     role.Vars,
					}); err != nil {
						return err
					}
					roleTasks, err := c.block2Task(context.WithValue(hctx, _const.CtxBlockRole, role.Role), o, role.Block, role.When.Data, ruid, variable.BlockLocation)
					if err != nil {
						klog.Errorf("[Pipeline %s] get role from play %s error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), puid, err)
						return err
					}
					nsTasks.Items = append(nsTasks.Items, roleTasks...)
				}
				// generate task from tasks
				tasks, err := c.block2Task(hctx, o, play.Tasks, nil, puid, variable.BlockLocation)
				if err != nil {
					klog.Errorf("[Pipeline %s] get pre task from  play %s error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), puid, err)
					return err
				}
				nsTasks.Items = append(nsTasks.Items, tasks...)
				// generate task from post tasks
				postTasks, err := c.block2Task(hctx, o, play.Tasks, nil, puid, variable.BlockLocation)
				if err != nil {
					klog.Errorf("[Pipeline %s] get pre task from  play %s error %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), puid, err)
					return err
				}
				nsTasks.Items = append(nsTasks.Items, postTasks...)
			}
		}

		for _, task := range nsTasks.Items {
			if err := c.client.Create(ctx, &task); err != nil {
				klog.Errorf("[Pipeline %s] create task %s error: %v", ctrlclient.ObjectKeyFromObject(o.Pipeline), task.Name, err)
				return err
			}
		}
	}

	return nil
}

// block2Task convert ansible block to task
func (k *taskController) block2Task(ctx context.Context, o AddTaskOptions, ats []kkcorev1.Block, when []string, parentLocation string, locationType variable.LocationType) ([]kubekeyv1alpha1.Task, error) {
	var tasks []kubekeyv1alpha1.Task

	for _, at := range ats {
		if !at.Taggable.IsEnabled(o.Pipeline.Spec.Tags, o.Pipeline.Spec.SkipTags) {
			continue
		}
		buid := uuid.NewString()
		if err := o.variable.Merge(variable.LocationMerge{
			Uid:      buid,
			ParentID: parentLocation,
			Type:     locationType,
			Name:     at.Name,
			Vars:     at.Vars,
		}); err != nil {
			return nil, err
		}
		atWhen := append(when, at.When.Data...)

		if len(at.Block) != 0 {
			// add block
			bt, err := k.block2Task(ctx, o, at.Block, atWhen, buid, variable.BlockLocation)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, bt...)

			if len(at.Always) != 0 {
				at, err := k.block2Task(ctx, o, at.Always, atWhen, buid, variable.AlwaysLocation)
				if err != nil {
					return nil, err
				}
				tasks = append(tasks, at...)
			}
			if len(at.Rescue) != 0 {
				rt, err := k.block2Task(ctx, o, at.Rescue, atWhen, buid, variable.RescueLocation)
				if err != nil {
					return nil, err
				}
				tasks = append(tasks, rt...)
			}
		} else {
			task := converter.MarshalBlock(context.WithValue(context.WithValue(ctx, _const.CtxBlockWhen, atWhen), _const.CtxBlockTaskUID, buid),
				at, o.Pipeline)

			for n, a := range at.UnknownFiled {
				data, err := json.Marshal(a)
				if err != nil {
					return nil, err
				}
				if m := modules.FindModule(n); m != nil {
					task.Spec.Module.Name = n
					task.Spec.Module.Args = runtime.RawExtension{Raw: data}
					break
				}
			}
			if task.Spec.Module.Name == "" { // action is necessary for a task
				return nil, fmt.Errorf("no module/action detected in task: %s", task.Name)
			}
			tasks = append(tasks, *task)
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
		klog.Errorf("Queue item %v was not a Request", obj)
		// Return true, don't take a break
		return true
	}

	result, err := k.taskReconciler.Reconcile(ctx, req)
	switch {
	case err != nil:
		k.wq.AddRateLimited(req)
		klog.Errorf("Reconciler error: %v", err)
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
