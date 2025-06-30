package variable

import (
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
)

// ***************************** MergeFunc ***************************** //

// MergeRemoteVariable merges remote variables to a specific host.
// It takes a map of data and a hostname, and merges the data into the host's RemoteVars
// if the RemoteVars are currently empty. This prevents overwriting existing remote variables.
var MergeRemoteVariable = func(data map[string]any, hostname string) MergeFunc {
	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}

		if hostname == "" {
			return errors.New("when merge source is remote. HostName cannot be empty")
		}
		if _, ok := vv.value.Hosts[hostname]; !ok {
			return errors.Errorf("when merge source is remote. HostName %s not exist", hostname)
		}

		// it should not be changed
		if hv := vv.value.Hosts[hostname]; len(hv.RemoteVars) == 0 {
			hv.RemoteVars = data
			vv.value.Hosts[hostname] = hv
		}

		return nil
	}
}

// MergeRuntimeVariable parses variables using a specific host's context and merges them to the host's runtime variables.
// It takes a YAML node and a list of hostnames, then for each host:
// 1. Gets all variables for the host to create a parsing context
// 2. Parses the YAML node using that context
// 3. Merges the parsed data into the host's RuntimeVars
var MergeRuntimeVariable = func(node yaml.Node, hosts ...string) MergeFunc {
	if node.IsZero() {
		// skip
		return emptyMergeFunc
	}

	return func(v Variable) error {
		for _, hostname := range hosts {
			vv, ok := v.(*variable)
			if !ok {
				return errors.New("variable type error")
			}

			// Avoid nested locking: prepare context for parsing outside locking region
			curVars, err := v.Get(GetAllVariable(hostname))
			if err != nil {
				return err
			}
			ctx, ok := curVars.(map[string]any)
			if !ok {
				return errors.Errorf("host %s variables type error, expect map[string]any", hostname)
			}
			data, err := parseYamlNode(ctx, node)
			if err != nil {
				return err
			}
			hv := vv.value.Hosts[hostname]
			hv.RuntimeVars = CombineVariables(hv.RuntimeVars, data)
			vv.value.Hosts[hostname] = hv
		}

		return nil
	}
}

// MergeHostsRuntimeVariable parses variables using a specific host's context and merges them to multiple hosts' runtime variables.
// It takes a YAML node, a source hostname for context, and a list of target hostnames.
// The function uses the source host's variables as context to parse the YAML node,
// then merges the parsed data into each target host's RuntimeVars.
var MergeHostsRuntimeVariable = func(node yaml.Node, hostname string, hosts ...string) MergeFunc {
	if node.IsZero() {
		// skip
		return emptyMergeFunc
	}

	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}

		// Avoid nested locking: prepare context for parsing outside locking region
		curVars, err := v.Get(GetAllVariable(hostname))
		if err != nil {
			return err
		}
		ctx, ok := curVars.(map[string]any)
		if !ok {
			return errors.Errorf("host %s variables type error, expect map[string]any", hostname)
		}
		data, err := parseYamlNode(ctx, node)
		if err != nil {
			return err
		}
		for _, h := range hosts {
			hv := vv.value.Hosts[h]
			hv.RuntimeVars = CombineVariables(hv.RuntimeVars, data)
			vv.value.Hosts[h] = hv
		}

		return nil
	}
}

// MergeResultVariable parses variables using a specific host's context and sets them as global result variables.
// It takes a YAML node and a hostname, then:
// 1. Gets all variables for the host to create a parsing context
// 2. Parses the YAML node using that context
// 3. Sets the parsed data as the global result variables (accessible across all hosts)
var MergeResultVariable = func(node yaml.Node, hostname string) MergeFunc {
	if node.IsZero() {
		// skip
		return emptyMergeFunc
	}

	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}

		// Avoid nested locking: prepare context for parsing outside locking region
		curVars, err := v.Get(GetAllVariable(hostname))
		if err != nil {
			return err
		}
		ctx, ok := curVars.(map[string]any)
		if !ok {
			return errors.Errorf("host %s variables type error, expect map[string]any", hostname)
		}
		result, err := parseYamlNode(ctx, node)
		if err != nil {
			return err
		}
		vv.value.Result = CombineVariables(vv.value.Result, result)

		return nil
	}
}
