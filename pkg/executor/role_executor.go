package executor

import (
	"context"
	"slices"

	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// roleExecutor is responsible for executing a role within a playbook.
// It manages the execution of role dependencies, variable merging, and block execution.
type roleExecutor struct {
	*option

	// playbook level config
	hosts []string // which hosts will run playbook
	// blocks level config
	role         kkprojectv1.Role
	ignoreErrors *bool // IgnoreErrors for role

	when []string // when condition for merge
	tags kkprojectv1.Taggable
}

// Exec executes the role, including its dependencies and blocks.
// It checks tags, merges variables, and recursively executes dependent roles and blocks.
func (e roleExecutor) Exec(ctx context.Context) error {
	// check tags: skip execution if tags do not match
	if !e.tags.IsEnabled(e.playbook.Spec.Tags, e.playbook.Spec.SkipTags) {
		// if not match the tags. skip
		return nil
	}
	// merge variables defined in the role for the current hosts
	if err := e.variable.Merge(variable.MergeRuntimeVariable(e.role.Vars.Nodes, e.hosts...)); err != nil {
		return err
	}
	// deal dependency role: execute all role dependencies recursively
	for _, dep := range e.role.RoleDependency {
		// recursively execute the dependency role
		if err := (roleExecutor{
			option:       e.option,
			role:         dep,
			hosts:        e.hosts,
			ignoreErrors: e.dealIgnoreErrors(dep.IgnoreErrors),
			when:         e.dealWhen(dep.When),
			tags:         e.dealTags(dep.Taggable),
		}.Exec(ctx)); err != nil {
			return err
		}
	}

	// execute the blocks defined in the role
	return (blockExecutor{
		option:       e.option,
		hosts:        e.hosts,
		ignoreErrors: e.ignoreErrors,
		blocks:       e.role.Block,
		role:         e.role.Role,
		when:         e.role.When.Data,
		tags:         e.tags,
	}.Exec(ctx))
}

// dealTags merges the provided taggable with the current tags.
// "tags" argument in block. block tags inherits parent block.
func (e roleExecutor) dealTags(taggable kkprojectv1.Taggable) kkprojectv1.Taggable {
	return kkprojectv1.JoinTag(taggable, e.tags)
}

// dealIgnoreErrors returns the ignore_errors value for the block.
// If ignore_errors is not defined in the block, it uses the value from the parent block.
func (e roleExecutor) dealIgnoreErrors(ie *bool) *bool {
	if ie == nil {
		ie = e.ignoreErrors
	}

	return ie
}

// dealWhen merges the provided when conditions with the current ones.
// Block when inherits parent block.
func (e roleExecutor) dealWhen(when kkprojectv1.When) []string {
	w := e.when
	for _, d := range when.Data {
		if !slices.Contains(w, d) {
			w = append(w, d)
		}
	}

	return w
}
