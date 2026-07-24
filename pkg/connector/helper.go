/*
Copyright 2024 The KubeSphere Authors.

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

package connector

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"
)

// blockDevicesFromLsblk runs lsblk on the target host and returns parsed block device trees.
func blockDevicesFromLsblk(ctx context.Context, conn Connector) (any, error) {
	stdout, stderr, err := conn.ExecuteCommand(ctx, "lsblk -J -b -o NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,MODEL")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run lsblk: stderr: %q", string(stderr))
	}
	devices, err := parseLsblkJSON(stdout)
	if err != nil {
		return nil, err
	}

	lvsStdout, _, err := conn.ExecuteCommand(ctx, "lvs --reportformat json -o lv_name,vg_name,lv_path,lv_dm_path 2>/dev/null")
	if err != nil {
		return devices, nil
	}
	if err := enrichBlockDevicesWithLVM(devices, lvsStdout); err != nil {
		klog.V(4).InfoS("skip lvm block device enrichment", "error", err)
	}

	return devices, nil
}

// parseLsblkJSON parses the JSON output of lsblk -J.
func parseLsblkJSON(stdout []byte) (any, error) {
	var data struct {
		Blockdevices []any `json:"blockdevices"`
	}
	if err := json.Unmarshal(stdout, &data); err != nil {
		return nil, errors.Wrap(err, "failed to parse lsblk JSON output")
	}
	return data.Blockdevices, nil
}

func parseLVSJSON(stdout []byte) (map[string]map[string]string, error) {
	var data struct {
		Report []struct {
			LV []struct {
				LVName   string `json:"lv_name"`
				VGName   string `json:"vg_name"`
				LVPath   string `json:"lv_path"`
				LVDMPath string `json:"lv_dm_path"`
			} `json:"lv"`
		} `json:"report"`
	}
	if err := json.Unmarshal(stdout, &data); err != nil {
		return nil, errors.Wrap(err, "failed to parse lvs JSON output")
	}

	result := make(map[string]map[string]string)
	for _, report := range data.Report {
		for _, lv := range report.LV {
			fields := map[string]string{
				"lv_name":    strings.TrimSpace(lv.LVName),
				"vg_name":    strings.TrimSpace(lv.VGName),
				"lv_path":    strings.TrimSpace(lv.LVPath),
				"lv_dm_path": strings.TrimSpace(lv.LVDMPath),
			}
			for _, key := range lvmDeviceKeys(fields) {
				result[key] = fields
			}
		}
	}

	return result, nil
}

func lvmDeviceKeys(fields map[string]string) []string {
	keys := make([]string, 0, 4)
	for _, key := range []string{fields["lv_path"], fields["lv_dm_path"]} {
		if key != "" {
			keys = append(keys, key, baseName(key))
		}
	}
	return keys
}

func baseName(path string) string {
	path = strings.TrimSpace(path)
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func enrichBlockDevicesWithLVM(devices any, stdout []byte) error {
	lvs, err := parseLVSJSON(stdout)
	if err != nil {
		return err
	}

	deviceList, ok := devices.([]any)
	if !ok {
		return nil
	}
	enrichBlockDeviceList(deviceList, lvs)
	return nil
}

func enrichBlockDeviceList(devices []any, lvs map[string]map[string]string) {
	for _, device := range devices {
		deviceMap, ok := device.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := deviceMap["name"].(string); ok {
			if fields, ok := lvs[name]; ok {
				for key, value := range fields {
					if value != "" {
						deviceMap[key] = value
					}
				}
			}
		}
		children, ok := deviceMap["children"].([]any)
		if ok {
			enrichBlockDeviceList(children, lvs)
		}
	}
}

// convertBytesToMap parses the given byte slice into a map[string]string using the provided split string.
// Only lines containing the split string are processed. Each such line is split into key and value at the first occurrence of the split string.
// Leading and trailing spaces are trimmed from both key and value.
// Example input (split = "="):
//
//	FOO=bar
//	BAZ = qux
//
// Result: map[string]string{"FOO": "bar", "BAZ": "qux"}
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

// convertBytesToSlice parses the given byte slice into a slice of map[string]string using the provided split string.
// Only lines containing the split string are processed. Each such line is split into key and value at the first occurrence of the split string.
// Leading and trailing spaces are trimmed from both key and value.
// Groups of key-value pairs are separated by empty lines. Each group is stored as a separate map in the resulting slice.
// Example input (split = ":"):
//
//	foo: bar
//	baz: qux
//
//	hello: world
//
//	Result: []map[string]string{
//	  {"foo": "bar", "baz": "qux"},
//	  {"hello": "world"},
//	}
func convertBytesToSlice(bs []byte, split string) []map[string]string {
	var config []map[string]string
	currentMap := make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewBuffer(bs))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line != "" {
			parts := strings.SplitN(line, split, 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				currentMap[key] = value
			}
		} else if len(currentMap) > 0 {
			// If encountering an empty line, add the current map to config and create a new map
			config = append(config, currentMap)
			currentMap = make(map[string]string)
		}
	}

	// Add the last map if not already added
	if len(currentMap) > 0 {
		config = append(config, currentMap)
	}

	return config
}
