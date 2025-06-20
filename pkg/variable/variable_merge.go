package variable

import (
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
)

// ***************************** MergeFunc ***************************** //

// MergeRemoteVariable merge variable to remote.
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

// MergeRuntimeVariable parse variable by specific host and merge to the host.
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

// MergeHostsRuntimeVariable parse variable by specific host and merge to given hosts.
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
