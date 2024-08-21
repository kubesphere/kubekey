package internal

import (
	"math"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

// Template parse file or vars which defined in project.
var Template = template.New("kubekey").Funcs(funcMap())

func funcMap() template.FuncMap {
	var f = sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")
	// add custom function
	f["toYaml"] = toYAML
	f["ipInCIDR"] = ipInCIDR
	f["pow"] = pow

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

	return strings.TrimSuffix(string(data), "\n")
}

// ipInCIDR get the IP of a specific location within the cidr range
func ipInCIDR(index int, cidr string) (string, error) {
	var ips = make([]string, 0)
	for _, s := range strings.Split(cidr, ",") {
		ips = append(ips, parseIP(s)...)
	}

	if index < 0 {
		index = max(len(ips)+index, 0)
	}
	index = max(index, 0)
	index = min(index, len(ips)-1)

	return ips[index], nil
}

// pow Get the "pow" power of "base". (base ** pow)
func pow(base, pow float64) (float64, error) {
	return math.Pow(base, pow), nil
}
