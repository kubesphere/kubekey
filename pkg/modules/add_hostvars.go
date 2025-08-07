package modules

import (
	"context"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
Module: add_hostvars

Description:
- Adds or updates host variables for one or more hosts.
- Accepts a YAML mapping with "hosts" (string or array of strings) and "vars" (mapping of variables).
- Similar in spirit to set_fact.go, but operates on multiple hosts.

Example Usage in Playbook Task:
  - name: Add custom variables to hosts
    add_hostvars:
      hosts: ["host1", "host2"]
      vars:
        custom_var: "value"
        another_var: 42

Return Values:
- On success: returns empty stdout and stderr.
- On failure: returns error message in stderr.
*/

// addHostvarsArgs holds the parsed arguments for the add_hostvars module.
type addHostvarsArgs struct {
	hosts []string  // List of hosts to which variables will be added
	vars  yaml.Node // Variables to add, as a YAML node
}

// newAddHostvarsArgs parses the raw module arguments and returns an addHostvarsArgs struct.
// The arguments must be a YAML mapping with "hosts" and "vars" keys.
// "hosts" can be a string or a sequence of strings.
// "vars" must be a mapping node.
func newAddHostvarsArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*addHostvarsArgs, error) {
	var node yaml.Node
	// Unmarshal the YAML document into a root node.
	if err := yaml.Unmarshal(raw.Raw, &node); err != nil {
		return nil, err
	}
	// The root node should be a document node with a single mapping node as its content.
	if len(node.Content) != 1 && node.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("module argument format error")
	}
	args := &addHostvarsArgs{}
	// Iterate over the mapping node's key-value pairs.
	for i := 0; i < len(node.Content[0].Content); i += 2 {
		keyNode := node.Content[0].Content[i]
		valueNode := node.Content[0].Content[i+1]
		switch keyNode.Value {
		case "hosts":
			var val any
			if err := valueNode.Decode(&val); err != nil {
				return nil, errors.New("cannot decode \"hosts\"")
			}
			args.hosts, _ = variable.StringSliceVar(vars, map[string]any{"hosts": val}, "hosts")
		case "vars":
			// Store the "vars" node for later processing.
			args.vars = *valueNode
		}
	}

	// Validate that hosts and vars are not empty.
	if len(args.hosts) == 0 {
		return nil, errors.New("\"hosts\" should be string or string array")
	}
	if args.vars.IsZero() {
		return nil, errors.New("\"vars\" should not be empty")
	}

	return args, nil
}

// ModuleAddHostvars handles the "add_hostvars" module, merging variables into the specified hosts.
// Returns empty stdout and stderr on success, or error message in stderr on failure.
func ModuleAddHostvars(ctx context.Context, options ExecOptions) (string, string, error) {
	// Get all host variables (for context, not used directly here).
	ha, err := options.getAllVariables()
	if err != nil {
		return StdoutFailed, StderrGetHostVariable, err
	}
	// Parse module arguments.
	args, err := newAddHostvarsArgs(ctx, options.Args, ha)
	if err != nil {
		return StdoutFailed, StderrParseArgument, err
	}
	ahn, err := options.Variable.Get(variable.GetHostnames(args.hosts))
	if err != nil {
		return StdoutFailed, "failed to get hostnames", err
	}
	hosts, ok := ahn.([]string)
	if !ok {
		return StdoutFailed, "failed to get actual hosts from given \"hosts\"", errors.Errorf("failed to get actual hosts from given \"hosts\"")
	}

	// Merge the provided variables into the specified hosts.
	if err := options.Variable.Merge(variable.MergeHostsRuntimeVariable(args.vars, options.Host, hosts...)); err != nil {
		return StdoutFailed, "failed to add_hostvars", errors.Wrap(err, "failed to add_hostvars")
	}

	return StdoutSuccess, "", nil
}

func init() {
	utilruntime.Must(RegisterModule("add_hostvars", ModuleAddHostvars))
}
