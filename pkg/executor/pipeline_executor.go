/*
Copyright 2024 The KubeSphere Authors.

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

package executor

import (
	"context"
	"errors"
	"fmt"
	"io"

	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kkprojectv1 "github.com/kubesphere/kubekey/v4/pkg/apis/project/v1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewPipelineExecutor return a new pipelineExecutor
func NewPipelineExecutor(ctx context.Context, client ctrlclient.Client, pipeline *kkcorev1.Pipeline, logOutput io.Writer) Executor {
	// get variable
	v, err := variable.New(ctx, client, *pipeline, source.FileSource)
	if err != nil {
		klog.V(5).ErrorS(nil, "convert playbook error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))

		return nil
	}

	return &pipelineExecutor{
		option: &option{
			client:    client,
			pipeline:  pipeline,
			variable:  v,
			logOutput: logOutput,
		},
	}
}

// executor for pipeline
type pipelineExecutor struct {
	*option
}

// Exec pipeline. covert playbook to block and executor it.
func (e pipelineExecutor) Exec(ctx context.Context) error {
	klog.V(5).InfoS("deal project", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
	pj, err := project.New(ctx, *e.pipeline, true)
	if err != nil {
		return fmt.Errorf("deal project error: %w", err)
	}

	// convert to transfer.Playbook struct
	pb, err := pj.MarshalPlaybook()
	if err != nil {
		return fmt.Errorf("convert playbook error: %w", err)
	}

	for _, play := range pb.Play {
		// check tags
		if !play.Taggable.IsEnabled(e.pipeline.Spec.Tags, e.pipeline.Spec.SkipTags) {
			// if not match the tags. skip
			continue
		}
		// hosts should contain all host's name. hosts should not be empty.
		var hosts []string
		if err := e.dealHosts(play.PlayHost, &hosts); err != nil {
			klog.V(4).ErrorS(err, "deal hosts error, skip this playbook", "hosts", play.PlayHost)

			continue
		}
		// when gather_fact is set. get host's information from remote.
		if err := e.dealGatherFacts(ctx, play.GatherFacts, hosts); err != nil {
			return fmt.Errorf("deal gather_facts argument error: %w", err)
		}
		// Batch execution, with each batch being a group of hosts run in serial.
		var batchHosts [][]string
		if err := e.dealSerial(play.Serial.Data, hosts, &batchHosts); err != nil {
			return fmt.Errorf("deal serial argument error: %w", err)
		}
		e.dealRunOnce(play.RunOnce, hosts, &batchHosts)
		// exec pipeline in each BatchHosts
		if err := e.execBatchHosts(ctx, play, batchHosts); err != nil {
			return fmt.Errorf("exec batch hosts error: %v", err)
		}
	}

	return nil
}

// execBatchHosts executor block in play order by: "pre_tasks" > "roles" > "tasks" > "post_tasks"
func (e pipelineExecutor) execBatchHosts(ctx context.Context, play kkprojectv1.Play, batchHosts [][]string) any {
	// generate and execute task.
	for _, serials := range batchHosts {
		// each batch hosts should not be empty.
		if len(serials) == 0 {
			klog.V(5).ErrorS(nil, "Host is empty", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

			return errors.New("host is empty")
		}

		if err := e.variable.Merge(variable.MergeRuntimeVariable(play.Vars, serials...)); err != nil {
			return fmt.Errorf("merge variable error: %w", err)
		}
		// generate task from pre tasks
		if err := (blockExecutor{
			option:       e.option,
			hosts:        serials,
			ignoreErrors: play.IgnoreErrors,
			blocks:       play.PreTasks,
			tags:         play.Taggable,
		}.Exec(ctx)); err != nil {
			return fmt.Errorf("execute pre-tasks from play error: %w", err)
		}
		// generate task from role
		for _, role := range play.Roles {
			if !role.Taggable.IsEnabled(e.pipeline.Spec.Tags, e.pipeline.Spec.SkipTags) {
				// if not match the tags. skip
				continue
			}
			if err := e.variable.Merge(variable.MergeRuntimeVariable(role.Vars, serials...)); err != nil {
				return fmt.Errorf("merge variable error: %w", err)
			}
			// use the most closely configuration
			ignoreErrors := role.IgnoreErrors
			if ignoreErrors == nil {
				ignoreErrors = play.IgnoreErrors
			}
			// role is block.
			if err := (blockExecutor{
				option:       e.option,
				hosts:        serials,
				ignoreErrors: ignoreErrors,
				blocks:       role.Block,
				role:         role.Role,
				when:         role.When.Data,
				tags:         kkprojectv1.JoinTag(role.Taggable, play.Taggable),
			}.Exec(ctx)); err != nil {
				return fmt.Errorf("execute role-tasks error: %w", err)
			}
		}
		// generate task from tasks
		if err := (blockExecutor{
			option:       e.option,
			hosts:        serials,
			ignoreErrors: play.IgnoreErrors,
			blocks:       play.Tasks,
			tags:         play.Taggable,
		}.Exec(ctx)); err != nil {
			return fmt.Errorf("execute tasks error: %w", err)
		}
		// generate task from post tasks
		if err := (blockExecutor{
			option:       e.option,
			hosts:        serials,
			ignoreErrors: play.IgnoreErrors,
			blocks:       play.PostTasks,
			tags:         play.Taggable,
		}.Exec(ctx)); err != nil {
			return fmt.Errorf("execute post-tasks error: %w", err)
		}
	}

	return nil
}

// dealHosts "hosts" argument in playbook. get hostname from kkprojectv1.PlayHost
func (e pipelineExecutor) dealHosts(host kkprojectv1.PlayHost, i *[]string) error {
	ahn, err := e.variable.Get(variable.GetHostnames(host.Hosts))
	if err != nil {
		return fmt.Errorf("getHostnames error: %w", err)
	}

	if h, ok := ahn.([]string); ok {
		*i = h
	}
	if len(*i) == 0 { // if hosts is empty skip this playbook
		return errors.New("hosts is empty")
	}

	return nil
}

// dealGatherFacts "gather_facts" argument in playbook. get host remote info and merge to variable
func (e pipelineExecutor) dealGatherFacts(ctx context.Context, gatherFacts bool, hosts []string) error {
	if !gatherFacts {
		// skip
		return nil
	}

	dealGatherFactsInHost := func(hostname string) error {
		v, err := e.variable.Get(variable.GetParamVariable(hostname))
		if err != nil {
			klog.V(5).ErrorS(err, "get host variable error", "hostname", hostname)

			return err
		}

		connectorVars := make(map[string]any)
		if c1, ok := v.(map[string]any)[_const.VariableConnector]; ok {
			if c2, ok := c1.(map[string]any); ok {
				connectorVars = c2
			}
		}
		// get host connector
		conn, err := connector.NewConnector(hostname, connectorVars)
		if err != nil {
			klog.V(5).ErrorS(err, "new connector error", "hostname", hostname)

			return err
		}
		if err := conn.Init(ctx); err != nil {
			klog.V(5).ErrorS(err, "init connection error", "hostname", hostname)

			return err
		}
		defer conn.Close(ctx)

		if gf, ok := conn.(connector.GatherFacts); ok {
			remoteInfo, err := gf.HostInfo(ctx)
			if err != nil {
				klog.V(5).ErrorS(err, "gatherFacts from connector error", "hostname", hostname)

				return err
			}
			if err := e.variable.Merge(variable.MergeRemoteVariable(remoteInfo, hostname)); err != nil {
				klog.V(5).ErrorS(err, "merge gather fact error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "host", hostname)

				return fmt.Errorf("merge gather fact error: %w", err)
			}
		}

		return nil
	}

	for _, hostname := range hosts {
		if err := dealGatherFactsInHost(hostname); err != nil {
			return err
		}
	}

	return nil
}

// dealSerial "serial" argument in playbook.
func (e pipelineExecutor) dealSerial(serial []any, hosts []string, batchHosts *[][]string) error {
	var err error
	*batchHosts, err = converter.GroupHostBySerial(hosts, serial)
	if err != nil {
		return fmt.Errorf("group host by serial error: %w", err)
	}

	return nil
}

// dealRunOnce argument in playbook. if RunOnce is true. it's always only run in the first hosts.
func (e pipelineExecutor) dealRunOnce(runOnce bool, hosts []string, batchHosts *[][]string) {
	if runOnce {
		// runOnce only run in first node
		*batchHosts = [][]string{{hosts[0]}}
	}
}
