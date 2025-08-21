package modules

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"gopkg.in/yaml.v3"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/utils"
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
func ModuleIncludeVars(ctx context.Context, options ExecOptions) (string, string, error) {
	// get host variable
	vd, err := options.getAllVariables()
	if err != nil {
		return StdoutFailed, StderrGetHostVariable, err
	}
	// check args
	includeVarsByte, err := variable.Extension2String(vd, options.Args)
	if err != nil {
		return StdoutFailed, StderrParseArgument, err
	}
	if len(includeVarsByte) == 0 {
		return StdoutFailed, "input file path wrong", errors.New("input value can not be empty")
	}
	arg := includeVarsArgs{
		includeVars: string(includeVarsByte),
	}
	if !filepath.IsLocal(arg.includeVars) {
		return StdoutFailed, "can not read remote file", errors.New("can not read remote file")
	}
	if !utils.HasSuffixIn(arg.includeVars, []string{"yaml", "yml"}) {
		return StdoutFailed, "input file type wrong", errors.New("input file type wrong")
	}

	var includeVarsFileContent []byte
	if filepath.IsAbs(arg.includeVars) {
		includeVarsFileContent, err = os.ReadFile(arg.includeVars)
		if err != nil {
			return StdoutFailed, "failed to read var file", errors.Wrap(err, "failed to read include variables file")
		}
	} else {
		pj, err := project.New(ctx, options.Playbook, false)
		if err != nil {
			return StdoutFailed, StderrGetPlaybook, err
		}
		fileReadPath := filepath.Join(options.Task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath], _const.VarsDir, arg.includeVars)
		includeVarsFileContent, err = pj.ReadFile(fileReadPath)
		if err != nil {
			return StdoutFailed, "failed to read var file", err
		}
	}

	var node yaml.Node
	// Unmarshal the YAML document into a root node.
	if err := yaml.Unmarshal(includeVarsFileContent, &node); err != nil {
		return StdoutFailed, StderrParseArgument, errors.Wrap(err, "failed to failed to unmarshal YAML")
	}

	if err := options.Variable.Merge(variable.MergeRuntimeVariable([]yaml.Node{node}, options.Host)); err != nil {
		return StdoutFailed, StderrParseArgument, errors.Wrap(err, "failed to merge runtime variables")
	}

	return StdoutSuccess, "", nil
}

func init() {
	utilruntime.Must(RegisterModule("include_vars", ModuleIncludeVars))
}
