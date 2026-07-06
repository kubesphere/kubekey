package variable

import (
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// EnsureLocalHost registers the implicit localhost host and syncs runtime vars
// from inventory hosts so local-only plays (certs/init, download) can render templates.
var EnsureLocalHost = func() MergeFunc {
	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}
		ensureImplicitLocalHost(vv.value.Hosts)
		syncLocalHostRuntimeVars(vv.value.Hosts)
		return nil
	}
}

func syncLocalHostRuntimeVars(hosts map[string]host) {
	lh, ok := hosts[_const.VariableLocalHost]
	if !ok {
		return
	}

	var reference map[string]any
	referenceLen := 0
	for name, h := range hosts {
		if name == _const.VariableLocalHost {
			continue
		}
		if len(h.RuntimeVars) > referenceLen {
			referenceLen = len(h.RuntimeVars)
			reference = h.RuntimeVars
		}
	}
	if reference == nil {
		return
	}

	lh.RuntimeVars = CombineVariables(reference, lh.RuntimeVars)
	hosts[_const.VariableLocalHost] = lh
}

func syncImplicitLocalHostAfterRuntimeMerge(hosts map[string]host) {
	ensureImplicitLocalHost(hosts)
	syncLocalHostRuntimeVars(hosts)
}

// ***************************** MergeFunc ***************************** //

// MergeRemoteVariable merges the provided data as remote variables into the specified hosts.
// For each hostname, if the host exists and its RemoteVars is empty, it sets RemoteVars to the provided data.
// If the host does not exist, it returns an error.
var MergeRemoteVariable = func(data map[string]any, hostnames ...string) MergeFunc {
	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}

		for _, hostname := range hostnames {
			if _, ok := vv.value.Hosts[hostname]; !ok {
				return errors.Errorf("when merge source is remote. HostName %s not exist", hostname)
			}
			// always update remote variable
			hv := vv.value.Hosts[hostname]
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
var MergeRuntimeVariable = func(nodes []yaml.Node, hosts ...string) MergeFunc {
	if len(nodes) == 0 {
		// skip
		return emptyMergeFunc
	}

	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}

		for _, hostname := range hosts {
			// Avoid nested locking: prepare context for parsing outside locking region
			curVars, err := v.Get(GetAllVariable(hostname))
			if err != nil {
				return err
			}
			ctx, ok := curVars.(map[string]any)
			if !ok {
				return errors.Errorf("host %s variables type error, expect map[string]any", hostname)
			}
			for _, node := range nodes {
				if node.IsZero() {
					continue
				}
				data, err := parseYamlNode(ctx, copyNode(node))
				if err != nil {
					return err
				}
				hv := vv.value.Hosts[hostname]
				hv.RuntimeVars = CombineVariables(hv.RuntimeVars, data)
				vv.value.Hosts[hostname] = hv
			}
		}
		syncImplicitLocalHostAfterRuntimeMerge(vv.value.Hosts)

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
		syncImplicitLocalHostAfterRuntimeMerge(vv.value.Hosts)

		return nil
	}
}

// MergeResultVariable parses variables using a specific host's context and sets them as global result variables.
// It takes a YAML node and a hostname, then:
// 1. Gets all variables for the host to create a parsing context
// 2. Parses the YAML node using that context
// 3. Sets the parsed data as the global result variables (accessible across all hosts)
var MergeResultVariable = func(result any) MergeFunc {
	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}

		vv.value.Result = CombineVariables(vv.value.Result, map[string]any{resultKey: result})

		return nil
	}
}
