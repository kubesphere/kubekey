package tmpl

import (
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"

	"github.com/kubesphere/kubekey/v4/pkg/utils"
)

// default function docs: http://masterminds.github.io/sprig

func funcMap() template.FuncMap {
	var f = sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")
	// add custom function
	f["toYaml"] = toYAML
	f["fromYaml"] = fromYAML
	f["ipInCIDR"] = ipInCIDR
	f["ipFamily"] = ipFamily
	f["pow"] = pow
	f["subtractList"] = subtractList
	f["fileExist"] = fileExist
	f["unquote"] = unquote
	f["getStringSlice"] = getStringSlice

	return f
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v any) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}

	return strings.TrimSpace(string(data))
}

// fromYAML takes a YAML string, unmarshals it into an interface{}, and returns the result.
// If there is an error during unmarshaling, it will be returned along with nil for the value.
func fromYAML(v string) (any, error) {
	var output any
	err := yaml.Unmarshal([]byte(v), &output)

	return output, err
}

// ipInCIDR takes a comma-separated list of CIDR strings, parses each one to extract IPs using parseIP,
// and returns a combined slice of all IPs found. Returns an error only if parseIP fails (not shown here).
func ipInCIDR(cidr string) ([]string, error) {
	var ips = make([]string, 0)
	for _, s := range strings.Split(cidr, ",") {
		ips = append(ips, utils.ParseIP(s)...)
	}

	return ips, nil
}

// ipFamily returns the IP family (IPv4 or IPv6) of the given IP address or IP cidr.
func ipFamily(addrOrCIDR string) (string, error) {
	// from IP address
	var ip = net.ParseIP(addrOrCIDR)
	if ip == nil {
		// from IP cidr
		ipFromCIDR, _, err := net.ParseCIDR(addrOrCIDR)
		if err != nil {
			return "Invalid", errors.Errorf("%s is not ip or cidr", addrOrCIDR)
		}
		ip = ipFromCIDR
	}

	if ip.To4() != nil {
		return "IPv4", nil
	}

	return "IPv6", nil
}

// pow Get the "pow" power of "base". (base ** pow)
func pow(base, pow float64) (float64, error) {
	return math.Pow(base, pow), nil
}

// subtractList returns a new list containing elements from list a that are not in list b.
// It first creates a set from list b for O(1) lookups, then builds a result list by
// including only elements from a that don't exist in the set.
func subtractList(a, b []any) ([]any, error) {
	set := make(map[any]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}

	result := make([]any, 0, len(a))
	for _, v := range a {
		if _, exists := set[v]; !exists {
			result = append(result, v)
		}
	}

	return result, nil
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func unquote(input any) string {
	if input == nil {
		return ""
	}
	inputStr, ok := input.(string)
	if !ok {
		return ""
	}
	output, err := strconv.Unquote(inputStr)
	if err != nil {
		return inputStr
	}
	return output
}

func getStringSlice(d map[string][]string, key string) []string {
	if val, ok := d[key]; ok {
		return val
	}
	return nil
}
