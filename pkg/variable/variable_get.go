package variable

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

// ***************************** GetFunc ***************************** //

// GetHostnames get all hostnames from a group or host
var GetHostnames = func(name []string) GetFunc {
	if len(name) == 0 {
		return emptyGetFunc
	}

	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}
		var hs []string
		for _, n := range name {
			// try parse hostname by Config.
			if pn, err := tmpl.ParseString(Extension2Variables(vv.value.Config.Spec), n); err == nil {
				n = pn
			}
			// add host to hs
			if _, ok := vv.value.Hosts[n]; ok {
				hs = append(hs, n)
			}
			// add group's host to gs
			for gn, gv := range ConvertGroup(vv.value.Inventory) {
				if gn == n {
					if gvd, ok := gv.([]string); ok {
						hs = mergeSlice(hs, gvd)
					}

					break
				}
			}

			// Add the specified host in the specified group to the hs.
			regexForIndex := regexp.MustCompile(`^(.*?)\[(\d+)\]$`)
			if match := regexForIndex.FindStringSubmatch(strings.TrimSpace(n)); match != nil {
				index, err := strconv.Atoi(match[2])
				if err != nil {
					klog.V(4).ErrorS(err, "convert index to int error", "index", match[2])

					return nil, err
				}
				if group, ok := ConvertGroup(vv.value.Inventory)[match[1]].([]string); ok {
					if index >= len(group) {
						return nil, fmt.Errorf("index %v out of range for group %s", index, group)
					}
					hs = append(hs, group[index])
				}
			}

			// add random host in group
			regexForRandom := regexp.MustCompile(`^(.+?)\s*\|\s*random$`)
			if match := regexForRandom.FindStringSubmatch(strings.TrimSpace(n)); match != nil {
				if group, ok := ConvertGroup(vv.value.Inventory)[match[1]].([]string); ok {
					hs = append(hs, group[rand.Intn(len(group))])
				}
			}
		}

		return hs, nil
	}
}

// GetAllVariable get all variable for a given host
var GetAllVariable = func(hostname string) GetFunc {
	// defaultHostVariable set default vars when hostname is "localhost"
	defaultHostVariable := func(hostname string, hostVars map[string]any) {
		if hostname == _const.VariableLocalHost {
			if _, ok := hostVars[_const.VariableIPv4]; !ok {
				hostVars[_const.VariableIPv4] = getLocalIP(_const.VariableIPv4)
			}
			if _, ok := hostVars[_const.VariableIPv6]; !ok {
				hostVars[_const.VariableIPv6] = getLocalIP(_const.VariableIPv6)
			}
		}
		if os, ok := hostVars[_const.VariableOS]; ok {
			// try to set hostname by current actual hostname.
			if osd, ok := os.(map[string]any); ok {
				hostVars[_const.VariableHostName] = osd[_const.VariableOSHostName]
			}
		}
		if _, ok := hostVars[_const.VariableInventoryName]; !ok {
			hostVars[_const.VariableInventoryName] = hostname
		}
		if _, ok := hostVars[_const.VariableHostName]; !ok {
			hostVars[_const.VariableHostName] = hostname
		}
	}

	getHostsVariable := func(v *variable) map[string]any {
		globalHosts := make(map[string]any)
		for hostname := range v.value.Hosts {
			hostVars := make(map[string]any)
			// set groups vars
			for _, gv := range v.value.Inventory.Spec.Groups {
				if slices.Contains(gv.Hosts, hostname) {
					hostVars = CombineVariables(hostVars, Extension2Variables(gv.Vars))
				}
			}
			// find from remote
			hostVars = CombineVariables(hostVars, v.value.Hosts[hostname].RemoteVars)
			// merge from runtime
			hostVars = CombineVariables(hostVars, v.value.Hosts[hostname].RuntimeVars)

			// merge from inventory vars
			hostVars = CombineVariables(hostVars, Extension2Variables(v.value.Inventory.Spec.Vars))
			// merge from inventory host vars
			hostVars = CombineVariables(hostVars, Extension2Variables(v.value.Inventory.Spec.Hosts[hostname]))
			// merge from config
			hostVars = CombineVariables(hostVars, Extension2Variables(v.value.Config.Spec))
			// set default localhost
			defaultHostVariable(hostname, hostVars)
			globalHosts[hostname] = hostVars
		}

		return globalHosts
	}

	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}
		hosts := getHostsVariable(vv)
		hostVars, ok := hosts[hostname].(map[string]any)
		if !ok {
			// cannot found hosts variable.
			return make(map[string]any), nil
		}
		hostVars = CombineVariables(hostVars, map[string]any{
			_const.VariableGlobalHosts: hosts,
		})
		hostVars = CombineVariables(hostVars, map[string]any{
			_const.VariableGroups: ConvertGroup(vv.value.Inventory),
		})

		return hostVars, nil
	}
}

// GetHostMaxLength get the max length for all hosts
var GetHostMaxLength = func() GetFunc {
	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}
		var hostnameMaxLen int
		for k := range vv.value.Hosts {
			hostnameMaxLen = max(len(k), hostnameMaxLen)
		}

		return hostnameMaxLen, nil
	}
}

// GetWorkDir returns the working directory from the configuration.
var GetWorkDir = func() GetFunc {
	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}

		return _const.GetWorkdirFromConfig(vv.value.Config), nil
	}
}
