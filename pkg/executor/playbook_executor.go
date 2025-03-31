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
	"io"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewPlaybookExecutor return a new playbookExecutor
func NewPlaybookExecutor(ctx context.Context, client ctrlclient.Client, playbook *kkcorev1.Playbook, logOutput io.Writer) Executor {
	// get variable
	v, err := variable.New(ctx, client, *playbook, source.FileSource)
	if err != nil {
		klog.V(5).ErrorS(nil, "convert playbook error", "playbook", ctrlclient.ObjectKeyFromObject(playbook))

		return nil
	}

	return &playbookExecutor{
		option: &option{
			client:    client,
			playbook:  playbook,
			variable:  v,
			logOutput: logOutput,
		},
	}
}

// executor for playbook
type playbookExecutor struct {
	*option
}

// Exec playbook. covert playbook to block and executor it.
func (e playbookExecutor) Exec(ctx context.Context) error {
	klog.V(5).InfoS("deal project", "playbook", ctrlclient.ObjectKeyFromObject(e.playbook))
	pj, err := project.New(ctx, *e.playbook, true)
	if err != nil {
		return err
	}
	// convert to transfer.Playbook struct
	pb, err := pj.MarshalPlaybook()
	if err != nil {
		return err
	}
	for _, play := range pb.Play {
		// check tags
		if !play.Taggable.IsEnabled(e.playbook.Spec.Tags, e.playbook.Spec.SkipTags) {
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
			return err
		}
		// Batch execution, with each batch being a group of hosts run in serial.
		var batchHosts [][]string
		if err := e.dealSerial(play.Serial.Data, hosts, &batchHosts); err != nil {
			return err
		}
		e.dealRunOnce(play.RunOnce, hosts, &batchHosts)
		// exec playbook in each BatchHosts
		if err := e.execBatchHosts(ctx, play, batchHosts); err != nil {
			return err
		}
	}

	return nil
}

// execBatchHosts executor block in play order by: "pre_tasks" > "roles" > "tasks" > "post_tasks"
func (e playbookExecutor) execBatchHosts(ctx context.Context, play kkprojectv1.Play, batchHosts [][]string) error {
	// generate and execute task.
	for _, serials := range batchHosts {
		// each batch hosts should not be empty.
		if len(serials) == 0 {
			return errors.Errorf("host is empty")
		}

		if err := e.variable.Merge(variable.MergeRuntimeVariable(play.Vars, serials...)); err != nil {
			return err
		}
		// generate task from pre tasks
		if err := (blockExecutor{
			option:       e.option,
			hosts:        serials,
			ignoreErrors: play.IgnoreErrors,
			blocks:       play.PreTasks,
			tags:         play.Taggable,
		}.Exec(ctx)); err != nil {
			return err
		}
		// generate task from role
		for _, role := range play.Roles {
			if !kkprojectv1.JoinTag(role.Taggable, play.Taggable).IsEnabled(e.playbook.Spec.Tags, e.playbook.Spec.SkipTags) {
				// if not match the tags. skip
				continue
			}
			if err := e.variable.Merge(variable.MergeRuntimeVariable(role.Vars, serials...)); err != nil {
				return err
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
				return err
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
			return err
		}
		// generate task from post tasks
		if err := (blockExecutor{
			option:       e.option,
			hosts:        serials,
			ignoreErrors: play.IgnoreErrors,
			blocks:       play.PostTasks,
			tags:         play.Taggable,
		}.Exec(ctx)); err != nil {
			return err
		}
	}

	return nil
}

// dealHosts "hosts" argument in playbook. get hostname from kkprojectv1.PlayHost
func (e playbookExecutor) dealHosts(host kkprojectv1.PlayHost, i *[]string) error {
	ahn, err := e.variable.Get(variable.GetHostnames(host.Hosts))
	if err != nil {
		return err
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
func (e playbookExecutor) dealGatherFacts(ctx context.Context, gatherFacts bool, hosts []string) error {
	if !gatherFacts {
		// skip
		return nil
	}
	dealGatherFactsInHost := func(hostname string) error {
		// get host connector
		conn, err := connector.NewConnector(hostname, e.variable)
		if err != nil {
			return err
		}
		if err := conn.Init(ctx); err != nil {
			return err
		}
		defer conn.Close(ctx)

		if gf, ok := conn.(connector.GatherFacts); ok {
			remoteInfo, err := gf.HostInfo(ctx)
			if err != nil {
				return err
			}
			if err := e.variable.Merge(variable.MergeRemoteVariable(remoteInfo, hostname)); err != nil {
				return err
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
func (e playbookExecutor) dealSerial(serial []any, hosts []string, batchHosts *[][]string) error {
	var err error
	*batchHosts, err = converter.GroupHostBySerial(hosts, serial)
	if err != nil {
		return err
	}

	return nil
}

// dealRunOnce argument in playbook. if RunOnce is true. it's always only run in the first hosts.
func (e playbookExecutor) dealRunOnce(runOnce bool, hosts []string, batchHosts *[][]string) {
	if runOnce {
		// runOnce only run in first node
		*batchHosts = [][]string{{hosts[0]}}
	}
}
