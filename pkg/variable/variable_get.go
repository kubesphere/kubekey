package variable

import (
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

// ***************************** GetFunc ***************************** //

// GetHostnames retrieves all hostnames from specified groups or hosts.
// It supports various hostname patterns including direct hostnames, group names,
// indexed group access (e.g., "group[0]"), and random selection (e.g., "group|random").
// The function also supports template parsing for hostnames using configuration variables.
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
		groupsMap := ConvertGroup(vv.value.Inventory)
		groupsMapAnyValue := make(map[string]any)
		for k, v := range groupsMap {
			groupsMapAnyValue[k] = v
		}
		tmpVariablesWithGroups := CombineVariables(Extension2Variables(vv.value.Config.Spec), groupsMapAnyValue)
		for _, n := range name {
			// Try to parse hostname using configuration variables as template context
			if pn, err := tmpl.ParseFunc(tmpVariablesWithGroups, n, func(b []byte) string { return string(b) }); err == nil {
				n = pn
			}
			// Add direct hostname if it exists in the hosts map
			if _, exists := vv.value.Hosts[n]; exists {
				hs = append(hs, n)
			}
			// Add all hosts from matching groups
			for gn, gv := range groupsMap {
				if gn == n {
					hs = CombineSlice(hs, gv)
					break
				}
			}

			// Handle indexed group access (e.g., "group[0]")
			regexForIndex := regexp.MustCompile(`^(.*?)\[(\d+)\]$`)
			if match := regexForIndex.FindStringSubmatch(strings.TrimSpace(n)); match != nil {
				index, err := strconv.Atoi(match[2])
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert %q to int", match[2])
				}
				if group, ok := ConvertGroup(vv.value.Inventory)[match[1]]; ok {
					if index >= len(group) {
						return nil, errors.Errorf("index %v out of range for group %s", index, group)
					}
					hs = append(hs, group[index])
				}
			}

			// Handle random host selection from group (e.g., "group|random")
			regexForRandom := regexp.MustCompile(`^(.+?)\s*\|\s*random$`)
			if match := regexForRandom.FindStringSubmatch(strings.TrimSpace(n)); match != nil {
				if group, ok := ConvertGroup(vv.value.Inventory)[match[1]]; ok {
					hs = append(hs, group[rand.Intn(len(group))])
				}
			}
		}

		return hs, nil
	}
}

// GetAllVariable retrieves all variables for a given host, including group variables,
// remote variables, runtime variables, inventory variables, and configuration variables.
// It also sets default variables for localhost and provides access to global host and group information.
var GetAllVariable = func(hostname string) GetFunc {
	// getLocalIP retrieves the IPv4 or IPv6 address for the localhost machine
	getLocalIP := func(ipType string) string {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			klog.ErrorS(err, "get network address error")
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipType == _const.VariableIPv4 && ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}

				if ipType == _const.VariableIPv6 && ipNet.IP.To16() != nil && ipNet.IP.To4() == nil {
					return ipNet.IP.String()
				}
			}
		}

		klog.V(4).Infof("cannot get local %s address", ipType)

		return ""
	}

	// defaultHostVariable sets default variables when hostname is "localhost"
	// It automatically detects and sets IPv4/IPv6 addresses and hostname information
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
			// Try to set hostname by current actual hostname from OS information
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

	// getHostsVariable builds a complete variable map for all hosts
	// by combining variables from multiple sources in a specific order
	getHostsVariable := func(v *variable) map[string]any {
		globalHosts := make(map[string]any)
		for hostname := range v.value.Hosts {
			hostVars := make(map[string]any)
			// Set group variables for hosts that belong to groups
			for _, gv := range v.value.Inventory.Spec.Groups {
				if slices.Contains(gv.Hosts, hostname) {
					hostVars = CombineVariables(hostVars, Extension2Variables(gv.Vars))
				}
			}
			// Merge remote variables (variables collected from the actual host)
			hostVars = CombineVariables(hostVars, v.value.Hosts[hostname].RemoteVars)
			// Merge runtime variables (variables set during playbook execution)
			hostVars = CombineVariables(hostVars, v.value.Hosts[hostname].RuntimeVars)

			// Merge inventory-level variables
			hostVars = CombineVariables(hostVars, Extension2Variables(v.value.Inventory.Spec.Vars))
			// Merge host-specific variables from inventory
			hostVars = CombineVariables(hostVars, Extension2Variables(v.value.Inventory.Spec.Hosts[hostname]))
			// Merge configuration variables
			hostVars = CombineVariables(hostVars, Extension2Variables(v.value.Config.Spec))
			// Set default variables for localhost
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
			// Return empty map if host variables cannot be found
			return make(map[string]any), nil
		}
		// Add global hosts information to the host variables
		hostVars = CombineVariables(hostVars, map[string]any{
			_const.VariableGlobalHosts: hosts,
		})
		// Add group information to the host variables
		hostVars = CombineVariables(hostVars, map[string]any{
			_const.VariableGroups: ConvertGroup(vv.value.Inventory),
		})

		return hostVars, nil
	}
}

// GetHostMaxLength calculates the maximum length of all hostnames.
// This is useful for formatting output or determining display widths.
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

// GetResultVariable returns the global result variables.
// This function retrieves the result variables that are set globally and accessible across all hosts.
var GetResultVariable = func() GetFunc {
	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}

		return vv.value.Result[resultKey], nil
	}
}
