package internal

import (
	"fmt"
	"math"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/version"
)

var Template = template.New("kubekey").Funcs(funcMap())

func funcMap() template.FuncMap {
	var f = sprig.TxtFuncMap()

	delete(f, "env")
	delete(f, "expandenv")
	// add custom function
	f["toYaml"] = toYAML
	f["versionAtLeast"] = versionAtLeast
	f["versionLessThan"] = versionLessThan
	f["ipInCIDR"] = ipInCIDR
	f["pow"] = pow

	return f
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// versionAtLeast tests if the "version2" is at least equal to a given minimum "version1".
func versionAtLeast(version1, version2 string) (bool, error) {
	v1, err := version.ParseGeneric(version1)
	if err != nil {
		return false, fmt.Errorf("convert %s to version error: %w", version1, err)
	}
	v2, err := version.ParseGeneric(version2)
	if err != nil {
		return false, fmt.Errorf("convert %s to version error: %w", version2, err)
	}
	return v2.AtLeast(v1), nil
}

// versionLessThan tests if the "version2" is less than a given "version1".
func versionLessThan(version1, version2 string) (bool, error) {
	v1, err := version.ParseGeneric(version1)
	if err != nil {
		return false, fmt.Errorf("convert %s to version error: %w", version1, err)
	}
	v2, err := version.ParseGeneric(version2)
	if err != nil {
		return false, fmt.Errorf("convert %s to version error: %w", version2, err)
	}
	return v2.LessThan(v1), nil
}

// ipInCIDR get the IP of a specific location within the cidr range
func ipInCIDR(index int, cidr string) (string, error) {
	var ips = make([]string, 0)
	for _, s := range strings.Split(cidr, ",") {
		ips = append(ips, parseIp(s)...)
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
