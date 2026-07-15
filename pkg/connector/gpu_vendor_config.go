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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// GPUVendorConfig is the top-level config loaded from gpu_vendors.yaml.
type GPUVendorConfig struct {
	GPUClassCodes    []string         `yaml:"gpu_class_codes"`
	GPUClassKeywords []string         `yaml:"gpu_class_keywords"`
	DeviceExclusions []GPUDeviceMatch `yaml:"device_exclusions"`
	Vendors          []GPUVendorEntry `yaml:"vendors"`
	classCodeSet     map[string]bool
	classKeywordList []string
	vendorIDMap      map[string]*GPUVendorEntry
	vendorList       []GPUVendorEntry
}

// GPUDeviceMatch describes a PCI device match rule from gpu_vendors.yaml.
type GPUDeviceMatch struct {
	ClassCode            string `yaml:"classCode"`
	VendorID             string `yaml:"vendorId"`
	DeviceID             string `yaml:"deviceId"`
	SubVendorID          string `yaml:"subVendorId"`
	SubDeviceID          string `yaml:"subDeviceId"`
	ClassNamePattern     string `yaml:"classNamePattern"`
	VendorNamePattern    string `yaml:"vendorNamePattern"`
	DeviceNamePattern    string `yaml:"deviceNamePattern"`
	SubVendorNamePattern string `yaml:"subVendorNamePattern"`
	SubDeviceNamePattern string `yaml:"subDeviceNamePattern"`
}

type pciDeviceIdentity struct {
	ClassCode     string
	ClassName     string
	VendorID      string
	VendorName    string
	DeviceID      string
	DeviceName    string
	SubVendorID   string
	SubVendorName string
	SubDeviceID   string
	SubDeviceName string
}

// GPUVendorEntry represents a single GPU vendor rule.
type GPUVendorEntry struct {
	VendorID    string `yaml:"vendorId"`
	NamePattern string `yaml:"namePattern"`
	Name        string `yaml:"name"`
	DriverClass string `yaml:"driverClass"`
	Priority    int    `yaml:"priority"`
}

// LoadGPUVendorConfig loads GPU vendor config from {workdir}/config/scanner/gpu_vendors.yaml.
func LoadGPUVendorConfig(workdir string) (*GPUVendorConfig, error) {
	configPath, data, err := readGPUVendorConfig(workdir)
	if err != nil {
		return nil, err
	}

	return parseGPUVendorConfig(configPath, data)
}

func readGPUVendorConfig(workdir string) (string, []byte, error) {
	paths := []string{
		filepath.Join(workdir, "config", "scanner", _const.GPUVendorConfig),
		filepath.Join(workdir, _const.GatherFactsDir, _const.GPUVendorConfig),
	}
	missing := make([]string, 0, len(paths))
	for _, configPath := range paths {
		data, err := os.ReadFile(configPath)
		if err == nil {
			return configPath, data, nil
		}
		if !os.IsNotExist(err) {
			return "", nil, fmt.Errorf("read gpu vendor config %s: %w", configPath, err)
		}
		missing = append(missing, configPath)
	}

	return "", nil, fmt.Errorf("read gpu vendor config: not found in %s", strings.Join(missing, ", "))
}

func parseGPUVendorConfig(configPath string, data []byte) (*GPUVendorConfig, error) {
	var cfg GPUVendorConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse gpu vendor config %s: %w", configPath, err)
	}

	cfg.classCodeSet = make(map[string]bool, len(cfg.GPUClassCodes))
	for _, code := range cfg.GPUClassCodes {
		cfg.classCodeSet[strings.ToLower(code)] = true
	}

	cfg.classKeywordList = make([]string, len(cfg.GPUClassKeywords))
	for i, kw := range cfg.GPUClassKeywords {
		cfg.classKeywordList[i] = strings.ToLower(kw)
	}

	cfg.vendorIDMap = make(map[string]*GPUVendorEntry, len(cfg.Vendors))
	cfg.vendorList = make([]GPUVendorEntry, len(cfg.Vendors))
	for i, v := range cfg.Vendors {
		cfg.vendorList[i] = v
		if v.VendorID != "" {
			cfg.vendorIDMap[strings.ToLower(v.VendorID)] = &cfg.vendorList[i]
		}
		if cfg.vendorList[i].NamePattern == "" {
			cfg.vendorList[i].NamePattern = strings.ToLower(v.Name)
		} else {
			cfg.vendorList[i].NamePattern = strings.ToLower(v.NamePattern)
		}
	}

	return &cfg, nil
}

// IsGPUClass checks if a device is a GPU class.
func (c *GPUVendorConfig) IsGPUClass(classCode, className string) bool {
	if classCode != "" && c.classCodeSet[strings.ToLower(classCode)] {
		return true
	}
	lower := strings.ToLower(className)
	for _, kw := range c.classKeywordList {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// IsExcludedDevice reports whether a broad class match is explicitly excluded.
func (c *GPUVendorConfig) IsExcludedDevice(device pciDeviceIdentity) bool {
	for _, rule := range c.DeviceExclusions {
		if rule.matches(device) {
			return true
		}
	}
	return false
}

func (m GPUDeviceMatch) matches(device pciDeviceIdentity) bool {
	if !matchID(m.ClassCode, device.ClassCode) {
		return false
	}
	if !matchID(m.VendorID, device.VendorID) {
		return false
	}
	if !matchID(m.DeviceID, device.DeviceID) {
		return false
	}
	if !matchID(m.SubVendorID, device.SubVendorID) {
		return false
	}
	if !matchID(m.SubDeviceID, device.SubDeviceID) {
		return false
	}
	if !matchPattern(m.ClassNamePattern, device.ClassName) {
		return false
	}
	if !matchPattern(m.VendorNamePattern, device.VendorName) {
		return false
	}
	if !matchPattern(m.DeviceNamePattern, device.DeviceName) {
		return false
	}
	if !matchPattern(m.SubVendorNamePattern, device.SubVendorName) {
		return false
	}
	if !matchPattern(m.SubDeviceNamePattern, device.SubDeviceName) {
		return false
	}
	return true
}

func matchID(pattern, value string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return true
	}
	return pattern == strings.ToLower(strings.TrimSpace(value))
}

func matchPattern(pattern, value string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), pattern)
}

// LookupVendor identifies a vendor.
func (c *GPUVendorConfig) LookupVendor(vendorID, vendorName string) *GPUVendorEntry {
	if vendorID != "" {
		if v, ok := c.vendorIDMap[strings.ToLower(vendorID)]; ok {
			return v
		}
	}
	lower := strings.ToLower(vendorName)
	for i := range c.vendorList {
		if strings.Contains(lower, c.vendorList[i].NamePattern) {
			return &c.vendorList[i]
		}
	}
	return nil
}
