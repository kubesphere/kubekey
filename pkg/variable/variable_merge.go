package variable

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
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
			return fmt.Errorf("when merge source is remote. HostName %s not exist", hostname)
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

			depth := 3
			if envDepth, err := strconv.Atoi(os.Getenv(_const.ENV_VARIABLE_PARSE_DEPTH)); err == nil {
				if envDepth != 0 {
					depth = envDepth
				}
			}

			for range depth {
				// merge to specify host
				curVariable, err := v.Get(GetAllVariable(hostname))
				if err != nil {
					return err
				}
				// parse variable
				if err := parseVariable(data, func(s string) (string, error) {
					// parse use total variable. the task variable should not contain template syntax.
					cv, ok := curVariable.(map[string]any)
					if !ok {
						return "", errors.New("variable type error")
					}

					return tmpl.ParseString(CombineVariables(data, cv), s)
				}); err != nil {
					return err
				}
				hv := vv.value.Hosts[hostname]
				hv.RuntimeVars = CombineVariables(hv.RuntimeVars, data)
				vv.value.Hosts[hostname] = hv
			}
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
		// merge to specify host
		curVariable, err := v.Get(GetAllVariable(hostname))
		if err != nil {
			return err
		}
		// parse variable
		if err := parseVariable(data, func(s string) (string, error) {
			// parse use total variable. the task variable should not contain template syntax.
			cv, ok := curVariable.(map[string]any)
			if !ok {
				return "", errors.New("variable type error")
			}

			return tmpl.ParseString(CombineVariables(data, cv), s)
		}); err != nil {
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
