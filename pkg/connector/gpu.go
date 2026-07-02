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
	"context"
	"strings"

	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// GPUInfo represents GPU information gathered via lspci.
type GPUInfo struct {
	HasNvidiaGPU bool        `json:"hasNvidiaGPU"`
	Devices      []GPUDevice `json:"devices,omitempty"`
}

// GPUDevice represents a detected GPU device.
type GPUDevice struct {
	PCIAddress  string `json:"pciAddress"`
	ClassCode   string `json:"classCode"`
	VendorID    string `json:"vendorId"`
	VendorName  string `json:"vendorName,omitempty"`
	DeviceID    string `json:"deviceId"`
	DeviceName  string `json:"deviceName,omitempty"`
	DriverClass string `json:"driverClass"`
}

// gpuInfoFromLspci runs lspci on the target host and returns parsed GPU info.
func gpuInfoFromLspci(ctx context.Context, workdir string, run commandRunner) (GPUInfo, error) {
	gpu := GPUInfo{Devices: []GPUDevice{}}

	vendorCfg, err := LoadGPUVendorConfig(workdir)
	if err != nil {
		return gpu, err
	}

	stdout, stderr, err := run.ExecuteCommand(ctx, "lspci -mm -nn")
	if err != nil {
		return gpu, err
	}
	if len(stderr) > 0 {
		klog.V(4).InfoS("lspci stderr output", "stderr", string(stderr))
	}

	for _, line := range strings.Split(string(stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		device := parseLspciMM(line, vendorCfg)
		if device == nil {
			continue
		}
		gpu.Devices = append(gpu.Devices, *device)
		if device.DriverClass == "nvidia" {
			gpu.HasNvidiaGPU = true
		}
	}

	return gpu, nil
}

// enrichHostInfoWithGPU appends GPU info to host facts when lspci is available.
func enrichHostInfoWithGPU(ctx context.Context, workdir string, hostInfo map[string]any, run commandRunner) {
	gpu, err := gpuInfoFromLspci(ctx, workdir, run)
	if err != nil {
		klog.V(4).InfoS("skip gpu gathering", "error", err)
		return
	}
	hostInfo[_const.VariableGPU] = gpu
}

func parseLspciMM(line string, vendorCfg *GPUVendorConfig) *GPUDevice {
	fields := extractQuotedFields(line)
	if len(fields) < 3 {
		return nil
	}

	pciAddr := strings.Fields(line)[0]

	className, classCode := splitLastBracketedID(fields[0])
	vendorName, vendorID := splitLastBracketedID(fields[1])
	deviceName, deviceID := splitLastBracketedID(fields[2])
	subVendorName := ""
	subVendorID := ""
	subDeviceName := ""
	subDeviceID := ""
	if len(fields) >= 5 {
		subVendorName, subVendorID = splitLastBracketedID(fields[3])
		subDeviceName, subDeviceID = splitLastBracketedID(fields[4])
	}

	deviceIdentity := pciDeviceIdentity{
		ClassCode:     classCode,
		ClassName:     className,
		VendorID:      vendorID,
		VendorName:    vendorName,
		DeviceID:      deviceID,
		DeviceName:    deviceName,
		SubVendorID:   subVendorID,
		SubVendorName: subVendorName,
		SubDeviceID:   subDeviceID,
		SubDeviceName: subDeviceName,
	}
	if vendorCfg.IsExcludedDevice(deviceIdentity) {
		return nil
	}
	if !vendorCfg.IsGPUClass(classCode, className) {
		return nil
	}

	driverClass := ""
	vendor := vendorCfg.LookupVendor(vendorID, vendorName)
	if vendor != nil {
		driverClass = vendor.DriverClass
		vendorName = vendor.Name
	}

	return &GPUDevice{
		PCIAddress:  pciAddr,
		ClassCode:   classCode,
		VendorID:    vendorID,
		VendorName:  vendorName,
		DeviceID:    deviceID,
		DeviceName:  deviceName,
		DriverClass: driverClass,
	}
}

func splitLastBracketedID(field string) (name, id string) {
	end := strings.LastIndex(field, "]")
	if end == -1 {
		return field, ""
	}
	start := strings.LastIndex(field[:end], "[")
	if start == -1 {
		return field, ""
	}
	id = field[start+1 : end]
	name = strings.TrimSpace(field[:start])
	return
}

func extractQuotedFields(line string) []string {
	var fields []string
	rest := line
	for {
		start := strings.Index(rest, "\"")
		if start == -1 {
			break
		}
		end := strings.Index(rest[start+1:], "\"")
		if end == -1 {
			break
		}
		fields = append(fields, rest[start+1:start+1+end])
		rest = rest[start+1+end+1:]
	}
	return fields
}
