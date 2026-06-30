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

package storage

import (
	"strings"
	"testing"
)

func TestParseDevicesString(t *testing.T) {
	t.Parallel()

	devices, err := parseDevices(map[string]any{
		"device": "/dev/sdb1",
	})
	if err != nil {
		t.Fatalf("parseDevices() error = %v", err)
	}
	if len(devices) != 1 || devices[0] != "/dev/sdb1" {
		t.Fatalf("parseDevices() = %#v", devices)
	}
}

func TestParseDevicesList(t *testing.T) {
	t.Parallel()

	devices, err := parseDevices(map[string]any{
		"device": []any{"/dev/sdb1", "sdc1"},
	})
	if err != nil {
		t.Fatalf("parseDevices() error = %v", err)
	}
	if len(devices) != 2 || devices[0] != "/dev/sdb1" || devices[1] != "/dev/sdc1" {
		t.Fatalf("parseDevices() = %#v", devices)
	}
}

func TestParseDevicesJSONList(t *testing.T) {
	t.Parallel()

	devices, err := parseDevices(map[string]any{
		"device": `["/dev/sdb1","/dev/sdc1"]`,
	})
	if err != nil {
		t.Fatalf("parseDevices() error = %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("parseDevices() = %#v", devices)
	}
}

func TestParseItem(t *testing.T) {
	t.Parallel()

	item, err := parseItem(map[string]any{
		"disk":       "sdb",
		"filesystem": "xfs",
		"partition":  true,
	})
	if err != nil {
		t.Fatalf("parseItem() error = %v", err)
	}
	if len(item.Devices) != 1 || item.Devices[0] != "/dev/sdb" {
		t.Fatalf("Devices = %#v, want [/dev/sdb]", item.Devices)
	}
}

func TestParseItemWithLVM(t *testing.T) {
	t.Parallel()

	item, err := parseItem(map[string]any{
		"device":        []any{"/dev/sdb1", "/dev/sdc1"},
		"filesystem":    "xfs",
		"mount_options": "prjquota",
		"lvm": map[string]any{
			"vg_name": "vg_data",
			"lv_name": "lv_data",
		},
	})
	if err != nil {
		t.Fatalf("parseItem() error = %v", err)
	}
	if item.LVM == nil {
		t.Fatalf("expected LVM config")
	}
	if len(item.Devices) != 2 {
		t.Fatalf("Devices = %#v", item.Devices)
	}
	if formatMountOptions(item.MountOptions) != "defaults,prjquota" {
		t.Fatalf("mount options = %q", formatMountOptions(item.MountOptions))
	}
}

func TestParseItemWithWebInstallerAliases(t *testing.T) {
	t.Parallel()

	item, err := parseItem(map[string]any{
		"device":       "vda1",
		"filesystem":   "xfs",
		"mountpoint":   "/data",
		"mount_option": "prjquota",
		"vg_name":      "vg_data",
	})
	if err != nil {
		t.Fatalf("parseItem() error = %v", err)
	}
	if item.MountPoint != "/data" {
		t.Fatalf("MountPoint = %q, want /data", item.MountPoint)
	}
	if item.MountOptions != "prjquota" {
		t.Fatalf("MountOptions = %q, want prjquota", item.MountOptions)
	}
	if item.LVM == nil || item.LVM.VGName != "vg_data" || item.LVM.LVName != "lv_data" {
		t.Fatalf("LVM = %#v", item.LVM)
	}
}

func TestParseItemDirectFormatRequiresSingleDevice(t *testing.T) {
	t.Parallel()

	_, err := parseItem(map[string]any{
		"device":     []any{"/dev/sdb1", "/dev/sdc1"},
		"filesystem": "xfs",
	})
	if err == nil {
		t.Fatalf("expected error for multiple devices without lvm")
	}
}

func TestFormatMountOptions(t *testing.T) {
	t.Parallel()

	if got := formatMountOptions(""); got != "defaults" {
		t.Fatalf("empty = %q", got)
	}
	if got := formatMountOptions("prjquota"); got != "defaults,prjquota" {
		t.Fatalf("prjquota = %q", got)
	}
}

func TestBuildFormatScript(t *testing.T) {
	t.Parallel()

	script, err := buildFormatScript(Item{
		Devices:      []string{"/dev/sdb"},
		Filesystem:   "ext4",
		Partition:    true,
		MountPoint:   "/data",
		MountOptions: "prjquota",
	})
	if err != nil {
		t.Fatalf("buildFormatScript() error = %v", err)
	}
	for _, want := range []string{`DEVICE="/dev/sdb"`, `MOUNT_OPTIONS="defaults,prjquota"`, `mount_and_persist`} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %q:\n%s", want, script)
		}
	}
}

func TestBuildLVMScript(t *testing.T) {
	t.Parallel()

	script, err := buildLVMScript(Item{
		Devices:    []string{"/dev/sdb1", "/dev/sdc1"},
		Filesystem: "xfs",
		MountPoint: "/data",
		LVM: &LVMConfig{
			VGName: "vg_data",
			LVName: "lv_data",
			LVSize: "100%FREE",
		},
	})
	if err != nil {
		t.Fatalf("buildLVMScript() error = %v", err)
	}
	for _, want := range []string{`DEVICES=(`, `"/dev/sdb1"`, `"/dev/sdc1"`, `vgcreate`, `vgextend`, `mount_and_persist`} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing %q:\n%s", want, script)
		}
	}
}
