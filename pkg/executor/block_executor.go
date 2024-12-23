package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type blockExecutor struct {
	*option

	// playbook level config
	hosts        []string // which hosts will run playbook
	ignoreErrors *bool    // IgnoreErrors for playbook
	// blocks level config
	blocks []kkprojectv1.Block
	role   string   // role name of blocks
	when   []string // when condition for blocks
	tags   kkprojectv1.Taggable
}

// Exec block. convert block to task and executor it.
func (e blockExecutor) Exec(ctx context.Context) error {
	for _, block := range e.blocks {
		hosts := e.dealRunOnce(block.RunOnce)
		tags := e.dealTags(block.Taggable)
		ignoreErrors := e.dealIgnoreErrors(block.IgnoreErrors)
		when := e.dealWhen(block.When)

		// // check tags
		if !tags.IsEnabled(e.pipeline.Spec.Tags, e.pipeline.Spec.SkipTags) {
			// if not match the tags. skip
			continue
		}

		// merge variable which defined in block
		if err := e.variable.Merge(variable.MergeRuntimeVariable(block.Vars, hosts...)); err != nil {
			klog.V(5).ErrorS(err, "merge variable error", "pipeline", e.pipeline, "block", block.Name)

			return err
		}

		switch {
		case len(block.Block) != 0:
			if err := e.dealBlock(ctx, hosts, ignoreErrors, when, tags, block); err != nil {
				klog.V(5).ErrorS(err, "deal block error", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

				return err
			}
		case block.IncludeTasks != "":
			// do nothing. include tasks has converted to blocks.
		default:
			if err := e.dealTask(ctx, hosts, when, block); err != nil {
				klog.V(5).ErrorS(err, "deal task error", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

				return err
			}
		}
	}

	return nil
}

// dealRunOnce "run_once" argument in block.
// If RunOnce is true, it's always only run in the first host.
// Otherwise, return hosts which defined in parent block.
func (e blockExecutor) dealRunOnce(runOnce bool) []string {
	hosts := e.hosts
	if runOnce {
		// runOnce only run in first node
		hosts = hosts[:1]
	}

	return hosts
}

// dealIgnoreErrors "ignore_errors" argument in block.
// if ignore_errors not defined in block, set it which defined in parent block.
func (e blockExecutor) dealIgnoreErrors(ie *bool) *bool {
	if ie == nil {
		ie = e.ignoreErrors
	}

	return ie
}

// dealTags "tags" argument in block. block tags inherits parent block
func (e blockExecutor) dealTags(taggable kkprojectv1.Taggable) kkprojectv1.Taggable {
	return kkprojectv1.JoinTag(taggable, e.tags)
}

// dealWhen argument in block. block when inherits parent block.
func (e blockExecutor) dealWhen(when kkprojectv1.When) []string {
	w := e.when
	for _, d := range when.Data {
		if !slices.Contains(w, d) {
			w = append(w, d)
		}
	}

	return w
}

// dealBlock "block" argument has defined in block. execute order is: block -> rescue -> always
// If rescue is defined, execute it when block execute error.
// If always id defined, execute it.
func (e blockExecutor) dealBlock(ctx context.Context, hosts []string, ignoreErrors *bool, when []string, tags kkprojectv1.Taggable, block kkprojectv1.Block) error {
	var errs error
	// exec block
	if err := (blockExecutor{
		option:       e.option,
		hosts:        hosts,
		ignoreErrors: ignoreErrors,
		role:         e.role,
		blocks:       block.Block,
		when:         when,
		tags:         tags,
	}.Exec(ctx)); err != nil {
		klog.V(5).ErrorS(err, "execute tasks from block error", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
		errs = errors.Join(errs, err)
	}
	// if block exec failed exec rescue
	if e.pipeline.Status.Phase == kkcorev1.PipelinePhaseFailed && len(block.Rescue) != 0 {
		if err := (blockExecutor{
			option:       e.option,
			hosts:        hosts,
			ignoreErrors: ignoreErrors,
			blocks:       block.Rescue,
			role:         e.role,
			when:         when,
			tags:         tags,
		}.Exec(ctx)); err != nil {
			klog.V(5).ErrorS(err, "execute tasks from rescue error", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
			errs = errors.Join(errs, err)
		}
	}
	// exec always after block
	if len(block.Always) != 0 {
		if err := (blockExecutor{
			option:       e.option,
			hosts:        hosts,
			ignoreErrors: ignoreErrors,
			blocks:       block.Always,
			role:         e.role,
			when:         when,
			tags:         tags,
		}.Exec(ctx)); err != nil {
			klog.V(5).ErrorS(err, "execute tasks from always error", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
			errs = errors.Join(errs, err)
		}
	}
	// when execute error. return
	return errs
}

// dealTask "block" argument is not defined in block.
func (e blockExecutor) dealTask(ctx context.Context, hosts []string, when []string, block kkprojectv1.Block) error {
	task := converter.MarshalBlock(e.role, hosts, when, block)
	// complete module by unknown field
	for n, a := range block.UnknownField {
		data, err := json.Marshal(a)
		if err != nil {
			klog.V(5).ErrorS(err, "Marshal unknown field error", "field", n, "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

			return err
		}
		if m := modules.FindModule(n); m != nil {
			task.Spec.Module.Name = n
			task.Spec.Module.Args = runtime.RawExtension{Raw: data}

			break
		}
	}
	if task.Spec.Module.Name == "" { // action is necessary for a task
		klog.V(5).ErrorS(nil, "No module/action detected in task", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

		return fmt.Errorf("no module/action detected in task: %s", task.Name)
	}
	// complete by pipeline
	task.GenerateName = e.pipeline.Name + "-"
	task.Namespace = e.pipeline.Namespace
	if err := ctrl.SetControllerReference(e.pipeline, task, e.client.Scheme()); err != nil {
		klog.V(5).ErrorS(err, "Set controller reference error", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

		return err
	}

	if err := (&taskExecutor{option: e.option, task: task, taskRunTimeout: 60 * time.Minute}).Exec(ctx); err != nil {
		klog.V(5).ErrorS(err, "exec task error", "block", block.Name, "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

		return err
	}

	return nil
}
