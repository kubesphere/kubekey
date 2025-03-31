package variable

import (
	"github.com/cockroachdb/errors"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
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
// TODO: support merge []byte to preserve yaml definition order
var MergeRuntimeVariable = func(data map[string]any, hosts ...string) MergeFunc {
	if len(data) == 0 || len(hosts) == 0 {
		// skip
		return emptyMergeFunc
	}

	return func(v Variable) error {
		for _, hostname := range hosts {
			vv, ok := v.(*variable)
			if !ok {
				return errors.New("variable type error")
			}
			// parse variable
			if err := parseVariable(data, runtimeVarParser(v, hostname, data)); err != nil {
				return err
			}
			hv := vv.value.Hosts[hostname]
			hv.RuntimeVars = CombineVariables(hv.RuntimeVars, data)
			vv.value.Hosts[hostname] = hv
		}

		return nil
	}
}

// MergeAllRuntimeVariable parse variable by specific host and merge to all hosts.
var MergeAllRuntimeVariable = func(data map[string]any, hostname string) MergeFunc {
	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}
		// parse variable
		if err := parseVariable(data, runtimeVarParser(v, hostname, data)); err != nil {
			return err
		}
		for h := range vv.value.Hosts {
			if _, ok := v.(*variable); !ok {
				return errors.New("variable type error")
			}
			hv := vv.value.Hosts[h]
			hv.RuntimeVars = CombineVariables(hv.RuntimeVars, data)
			vv.value.Hosts[h] = hv
		}

		return nil
	}
}

func runtimeVarParser(v Variable, hostname string, data map[string]any) func(string) (string, error) {
	return func(s string) (string, error) {
		curVariable, err := v.Get(GetAllVariable(hostname))
		if err != nil {
			return "", err
		}
		cv, ok := curVariable.(map[string]any)
		if !ok {
			return "", errors.Errorf("host %s variables type error, expect map[string]any", hostname)
		}

		return tmpl.ParseFunc(
			CombineVariables(data, cv),
			s,
			func(b []byte) string { return string(b) },
		)
	}
}
