package include_vars

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"gopkg.in/yaml.v3"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
Module: include_vars

Description:
- Adds or updates host variables for one or more hosts.

Example Usage in Playbook Task:
  - name: Add custom variables to hosts
    include_vars: path/file.yaml

Return Values:
- On success: returns empty stdout and stderr.
- On failure: returns error message in stderr.
*/

type includeVarsArgs struct {
	includeVars string
}

// ModuleIncludeVars handle the "include_vars" module ,add other var files into playbook
func ModuleIncludeVars(ctx context.Context, opts internal.ExecOptions) (string, string, error) {
	// get host variable
	vd, err := opts.GetAllVariables()
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetHostVariable, err
	}
	// check args
	includeVarsByte, err := variable.Extension2String(vd, opts.Args)
	if err != nil {
		return internal.StdoutFailed, internal.StderrParseArgument, err
	}
	if len(includeVarsByte) == 0 {
		return internal.StdoutFailed, "input file path wrong", errors.New("input value can not be empty")
	}
	arg := includeVarsArgs{
		includeVars: string(includeVarsByte),
	}
	if filepath.Ext(arg.includeVars) != ".yaml" && filepath.Ext(arg.includeVars) != ".yml" {
		return internal.StdoutFailed, "invalid file type", errors.Errorf("unsupported file extension %q: only .yaml and .yml files are allowed", arg.includeVars)
	}

	var includeVarsFileContent []byte
	if filepath.IsAbs(arg.includeVars) {
		includeVarsFileContent, err = os.ReadFile(arg.includeVars)
		if err != nil {
			return internal.StdoutFailed, "failed to read var file", errors.Wrap(err, "failed to read include variables file")
		}
	} else {
		pj, err := project.New(ctx, opts.Playbook, false)
		if err != nil {
			return internal.StdoutFailed, internal.StderrGetPlaybook, err
		}
		fileReadPath := filepath.Join(opts.Task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath], _const.VarsDir, arg.includeVars)
		includeVarsFileContent, err = pj.ReadFile(fileReadPath)
		if err != nil {
			return internal.StdoutFailed, "failed to read var file", err
		}
	}

	var node yaml.Node
	// Unmarshal the YAML document into a root node.
	if err := yaml.Unmarshal(includeVarsFileContent, &node); err != nil {
		return internal.StdoutFailed, internal.StderrParseArgument, errors.Wrap(err, "failed to failed to unmarshal YAML")
	}

	if err := opts.Merge(variable.MergeRuntimeVariable([]yaml.Node{node}, opts.Host)); err != nil {
		return internal.StdoutFailed, internal.StderrParseArgument, errors.Wrap(err, "failed to merge runtime variables")
	}

	return internal.StdoutSuccess, "", nil
}
