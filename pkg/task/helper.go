/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package task

import (
	"bufio"
	"bytes"
	"context"
	"strings"

	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// getGatherFact get host info
func getGatherFact(ctx context.Context, hostname string, vars variable.Variable) (variable.VariableData, error) {
	v, err := vars.Get(variable.HostVars{HostName: hostname})
	if err != nil {
		klog.V(4).ErrorS(err, "Get host variable error", "hostname", hostname)
		return nil, err
	}
	conn := connector.NewConnector(hostname, v.(variable.VariableData))
	if err := conn.Init(ctx); err != nil {
		klog.V(4).ErrorS(err, "Init connection error", "hostname", hostname)
		return nil, err
	}
	defer conn.Close(ctx)

	// os information
	osVars := make(variable.VariableData)
	var osRelease bytes.Buffer
	if err := conn.FetchFile(ctx, "/etc/os-release", &osRelease); err != nil {
		klog.V(4).ErrorS(err, "Fetch os-release error", "hostname", hostname)
		return nil, err
	}
	osVars["release"] = convertBytesToMap(osRelease.Bytes(), "=")
	kernel, err := conn.ExecuteCommand(ctx, "uname -r")
	if err != nil {
		klog.V(4).ErrorS(err, "Get kernel version error", "hostname", hostname)
		return nil, err
	}
	osVars["kernelVersion"] = string(bytes.TrimSuffix(kernel, []byte("\n")))
	hn, err := conn.ExecuteCommand(ctx, "hostname")
	if err != nil {
		klog.V(4).ErrorS(err, "Get hostname error", "hostname", hostname)
		return nil, err
	}
	osVars["hostname"] = string(bytes.TrimSuffix(hn, []byte("\n")))
	arch, err := conn.ExecuteCommand(ctx, "arch")
	if err != nil {
		klog.V(4).ErrorS(err, "Get arch error", "hostname", hostname)
		return nil, err
	}
	osVars["architecture"] = string(bytes.TrimSuffix(arch, []byte("\n")))

	// process information
	procVars := make(variable.VariableData)
	var cpu bytes.Buffer
	if err := conn.FetchFile(ctx, "/proc/cpuinfo", &cpu); err != nil {
		klog.V(4).ErrorS(err, "Fetch cpuinfo error", "hostname", hostname)
		return nil, err
	}
	procVars["cpuInfo"] = convertBytesToSlice(cpu.Bytes(), ":")
	var mem bytes.Buffer
	if err := conn.FetchFile(ctx, "/proc/meminfo", &mem); err != nil {
		klog.V(4).ErrorS(err, "Fetch meminfo error", "hostname", hostname)
		return nil, err
	}
	procVars["memInfo"] = convertBytesToMap(mem.Bytes(), ":")

	return variable.VariableData{
		"os":      osVars,
		"process": procVars,
	}, nil
}

// convertBytesToMap with split string, only convert line which contain split
func convertBytesToMap(bs []byte, split string) map[string]string {
	config := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewBuffer(bs))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, split, 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}
	return config
}

// convertBytesToSlice with split string. only convert line which contain split.
// group by empty line
func convertBytesToSlice(bs []byte, split string) []map[string]string {
	var config []map[string]string
	currentMap := make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewBuffer(bs))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) > 0 {
			parts := strings.SplitN(line, split, 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				currentMap[key] = value
			}
		} else {
			// If encountering an empty line, add the current map to config and create a new map
			if len(currentMap) > 0 {
				config = append(config, currentMap)
				currentMap = make(map[string]string)
			}
		}
	}

	// Add the last map if not already added
	if len(currentMap) > 0 {
		config = append(config, currentMap)
	}

	return config
}
