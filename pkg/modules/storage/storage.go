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
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

var (
	supportedFilesystems = map[string]struct{}{
		"ext4":  {},
		"xfs":   {},
		"btrfs": {},
	}
	lvmNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._+-]+$`)
)

// LVMConfig describes LVM volume management for a storage item.
type LVMConfig struct {
	VGName string
	LVName string
	LVSize string
}

// Item describes a single disk formatting request from inventory host vars.
type Item struct {
	Devices      []string
	Filesystem   string
	Partition    bool
	Overwrite    bool
	MountPoint   string
	MountOptions string
	LVM          *LVMConfig
}

// ModuleStorage formats block devices on the target host.
func ModuleStorage(ctx context.Context, opts internal.ExecOptions) (string, string, error) {
	_, err := opts.GetAllVariables()
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetHostVariable, err
	}

	item, err := parseItem(variable.Extension2Variables(opts.Args))
	if err != nil {
		return internal.StdoutFailed, "invalid storage configuration", err
	}

	conn, err := opts.GetConnector(ctx)
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetConnector, err
	}
	defer conn.Close(ctx)

	script, err := buildScript(item)
	if err != nil {
		return internal.StdoutFailed, "failed to build storage script", err
	}

	stdout, stderr, err := conn.ExecuteCommand(ctx, script)
	if err != nil {
		return string(stdout), string(stderr), err
	}

	return string(stdout), string(stderr), nil
}

func parseItem(args map[string]any) (Item, error) {
	item := Item{}

	devices, err := parseDevices(args)
	if err != nil {
		return Item{}, err
	}
	item.Devices = devices

	if filesystem, err := variable.StringVar(nil, args, _const.VariableStorageFilesystem); err == nil {
		item.Filesystem = strings.ToLower(strings.TrimSpace(filesystem))
	}
	if partition, err := variable.BoolVar(nil, args, _const.VariableStoragePartition); err == nil && partition != nil {
		item.Partition = *partition
	}
	if overwrite, err := variable.BoolVar(nil, args, _const.VariableStorageOverwrite); err == nil && overwrite != nil {
		item.Overwrite = *overwrite
	}
	if mountPoint, err := variable.StringVar(nil, args, _const.VariableStorageMountPoint); err == nil {
		item.MountPoint = strings.TrimSpace(mountPoint)
	}
	if mountOptions, err := variable.StringVar(nil, args, _const.VariableStorageMountOptions); err == nil {
		item.MountOptions = strings.TrimSpace(mountOptions)
	}

	lvm, err := parseLVM(args)
	if err != nil {
		return Item{}, err
	}
	item.LVM = lvm

	if item.Filesystem == "" {
		return Item{}, errors.New("filesystem is required")
	}
	if _, ok := supportedFilesystems[item.Filesystem]; !ok {
		return Item{}, errors.Errorf("unsupported filesystem %q, supported: ext4, xfs, btrfs", item.Filesystem)
	}
	if item.LVM == nil && len(item.Devices) != 1 {
		return Item{}, errors.New("direct formatting requires exactly one device")
	}

	return item, nil
}

func parseDevices(args map[string]any) ([]string, error) {
	devices := make([]string, 0)

	if disk, err := variable.StringVar(nil, args, _const.VariableStorageDisk); err == nil {
		if normalized := normalizeDevicePath(disk); normalized != "" {
			devices = append(devices, normalized)
		}
	}

	deviceVal, found, err := unstructured.NestedFieldNoCopy(args, _const.VariableStorageDevice)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if found && deviceVal != nil {
		parsed, err := parseDeviceValue(deviceVal)
		if err != nil {
			return nil, err
		}
		devices = append(devices, parsed...)
	}

	devices = dedupeDevices(devices)
	if len(devices) == 0 {
		return nil, errors.New("device must specify at least one disk or partition")
	}

	return devices, nil
}

func parseDeviceValue(value any) ([]string, error) {
	switch v := value.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return nil, nil
		}
		if strings.HasPrefix(v, "[") {
			var list []string
			if err := json.Unmarshal([]byte(v), &list); err != nil {
				return nil, errors.Wrap(err, "failed to parse device json array")
			}
			return normalizeDevicePaths(list), nil
		}
		return []string{normalizeDevicePath(v)}, nil
	case []any:
		devices := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, errors.Errorf("unsupported device list item type %T", item)
			}
			if normalized := normalizeDevicePath(s); normalized != "" {
				devices = append(devices, normalized)
			}
		}
		return devices, nil
	case []string:
		return normalizeDevicePaths(v), nil
	default:
		return nil, errors.Errorf("unsupported device type %T", value)
	}
}

func normalizeDevicePaths(paths []string) []string {
	devices := make([]string, 0, len(paths))
	for _, path := range paths {
		if normalized := normalizeDevicePath(path); normalized != "" {
			devices = append(devices, normalized)
		}
	}
	return devices
}

func dedupeDevices(devices []string) []string {
	seen := make(map[string]struct{}, len(devices))
	result := make([]string, 0, len(devices))
	for _, device := range devices {
		if _, ok := seen[device]; ok {
			continue
		}
		seen[device] = struct{}{}
		result = append(result, device)
	}
	return result
}

func parseLVM(args map[string]any) (*LVMConfig, error) {
	lvmVal, found, err := unstructured.NestedFieldNoCopy(args, _const.VariableStorageLVM)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !found || lvmVal == nil {
		return nil, nil
	}

	cfg := &LVMConfig{LVSize: "100%FREE"}

	switch v := lvmVal.(type) {
	case bool:
		if !v {
			return nil, nil
		}
	case map[string]any:
		if vgName, err := variable.StringVar(nil, v, _const.VariableStorageVGName); err == nil {
			cfg.VGName = strings.TrimSpace(vgName)
		}
		if lvName, err := variable.StringVar(nil, v, _const.VariableStorageLVName); err == nil {
			cfg.LVName = strings.TrimSpace(lvName)
		}
		if lvSize, err := variable.StringVar(nil, v, _const.VariableStorageLVSize); err == nil {
			cfg.LVSize = strings.TrimSpace(lvSize)
		}
	default:
		return nil, errors.Errorf("unsupported lvm configuration type %T", lvmVal)
	}

	if vgName, err := variable.StringVar(nil, args, _const.VariableStorageVGName); err == nil && cfg.VGName == "" {
		cfg.VGName = strings.TrimSpace(vgName)
	}
	if lvName, err := variable.StringVar(nil, args, _const.VariableStorageLVName); err == nil && cfg.LVName == "" {
		cfg.LVName = strings.TrimSpace(lvName)
	}
	if lvSize, err := variable.StringVar(nil, args, _const.VariableStorageLVSize); err == nil && cfg.LVSize == "" {
		cfg.LVSize = strings.TrimSpace(lvSize)
	}

	if cfg.VGName == "" && cfg.LVName == "" {
		return nil, nil
	}
	if cfg.VGName == "" || cfg.LVName == "" {
		return nil, errors.New("lvm requires vg_name and lv_name")
	}
	if err := validateLVMName(cfg.VGName, "vg_name"); err != nil {
		return nil, err
	}
	if err := validateLVMName(cfg.LVName, "lv_name"); err != nil {
		return nil, err
	}
	if cfg.LVSize == "" {
		cfg.LVSize = "100%FREE"
	}

	return cfg, nil
}

func validateLVMName(name, field string) error {
	if !lvmNamePattern.MatchString(name) {
		return errors.Errorf("invalid %s %q", field, name)
	}
	return nil
}

func normalizeDevicePath(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, "/dev/") {
		return name
	}
	return "/dev/" + strings.TrimPrefix(name, "/dev/")
}

func formatMountOptions(options string) string {
	options = strings.TrimSpace(options)
	if options == "" {
		return "defaults"
	}
	if strings.Contains(options, "defaults") {
		return options
	}
	return "defaults," + options
}

func buildScript(item Item) (string, error) {
	if item.LVM != nil {
		return buildLVMScript(item)
	}
	return buildFormatScript(item)
}

func buildDeviceResolveScript() string {
	return `
is_partition_device() {
  local dev="$1"
  [ "$(lsblk -dn -o TYPE "$dev" 2>/dev/null | head -1)" = "part" ]
}

resolve_target_device() {
  TARGET_DEV="$DEVICE"
  if is_partition_device "$DEVICE"; then
    echo "using partition device $DEVICE"
    return 0
  fi
  if [ "$PARTITION" != true ]; then
    return 0
  fi
  PART_COUNT=$(lsblk -ln -o TYPE "$DEVICE" 2>/dev/null | grep -c '^part$' || true)
  if [ "$PART_COUNT" -eq 0 ]; then
    parted -s "$DEVICE" mklabel gpt
    parted -s -a optimal "$DEVICE" mkpart primary 0%% 100%%
    partprobe "$DEVICE" 2>/dev/null || true
    udevadm settle 2>/dev/null || sleep 2
  fi
  if [[ "$DEVICE" == *"nvme"* ]]; then
    TARGET_DEV="${DEVICE}p1"
  else
    TARGET_DEV="${DEVICE}1"
  fi
}

refuse_system_disk() {
  local dev="$1"
  local pkname
  pkname=$(lsblk -no PKNAME "$dev" 2>/dev/null | head -1 || true)
  if [ -z "$pkname" ]; then
    pkname=$(basename "$dev")
  fi
  ROOT_SOURCE=$(findmnt -n -o SOURCE / 2>/dev/null || true)
  if [ -z "$ROOT_SOURCE" ]; then
    return 0
  fi
  ROOT_DISK=$(lsblk -no PKNAME "$ROOT_SOURCE" 2>/dev/null | head -1 || true)
  if [ -z "$ROOT_DISK" ]; then
    ROOT_DISK=$(basename "$ROOT_SOURCE")
  fi
  if [ -n "$ROOT_DISK" ] && [ "$pkname" = "$ROOT_DISK" ]; then
    echo "refusing to use system disk $dev" >&2
    exit 1
  fi
}

ensure_block_device() {
  local dev="$1"
  if [ ! -b "$dev" ]; then
    echo "block device $dev does not exist" >&2
    exit 1
  fi
}

refuse_mounted_device() {
  local dev="$1"
  if findmnt -n -S "$dev" >/dev/null 2>&1; then
    echo "device $dev is mounted, refusing to continue" >&2
    exit 1
  fi
}

mount_and_persist() {
  local dev="$1"
  if [ -z "$MOUNT_POINT" ]; then
    return 0
  fi
  mkdir -p "$MOUNT_POINT"
  if mountpoint -q "$MOUNT_POINT"; then
    return 0
  fi
  mount -o "$MOUNT_OPTIONS" "$dev" "$MOUNT_POINT"
  if blkid "$dev" >/dev/null 2>&1; then
    UUID=$(blkid -s UUID -o value "$dev")
    if ! grep -q "$UUID" /etc/fstab; then
      echo "UUID=$UUID $MOUNT_POINT $FILESYSTEM $MOUNT_OPTIONS 0 0" >> /etc/fstab
    fi
  elif ! grep -q "$dev" /etc/fstab; then
    echo "$dev $MOUNT_POINT $FILESYSTEM $MOUNT_OPTIONS 0 0" >> /etc/fstab
  fi
  echo "mounted $dev to $MOUNT_POINT with options $MOUNT_OPTIONS"
}
`
}

func buildFormatScript(item Item) (string, error) {
	if len(item.Devices) == 0 {
		return "", errors.New("device list is empty")
	}

	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail
%s
DEVICE=%q
FILESYSTEM=%q
PARTITION=%t
OVERWRITE=%t
MOUNT_POINT=%q
MOUNT_OPTIONS=%q

ensure_block_device "$DEVICE"
refuse_system_disk "$DEVICE"
resolve_target_device
ensure_block_device "$TARGET_DEV"
refuse_mounted_device "$TARGET_DEV"

FORMAT_DEV="$TARGET_DEV"
if blkid "$FORMAT_DEV" >/dev/null 2>&1; then
  if [ "$OVERWRITE" != true ]; then
    echo "filesystem already exists on $FORMAT_DEV, skipping"
    exit 0
  fi
  echo "overwriting existing filesystem on $FORMAT_DEV"
else
  echo "creating $FILESYSTEM filesystem on $FORMAT_DEV"
fi

mkfs -t "$FILESYSTEM" "$FORMAT_DEV"
mount_and_persist "$FORMAT_DEV"
`, buildDeviceResolveScript(), item.Devices[0], item.Filesystem, item.Partition, item.Overwrite, item.MountPoint, formatMountOptions(item.MountOptions)), nil
}

func quoteBashWords(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = fmt.Sprintf("%q", value)
	}
	return strings.Join(quoted, " ")
}

func buildLVMScript(item Item) (string, error) {
	if len(item.Devices) == 0 {
		return "", errors.New("device list is empty")
	}
	if item.LVM == nil {
		return "", errors.New("lvm configuration is required")
	}

	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail
%s
DEVICES=(%s)
FILESYSTEM=%q
PARTITION=%t
OVERWRITE=%t
MOUNT_POINT=%q
MOUNT_OPTIONS=%q
VG_NAME=%q
LV_NAME=%q
LV_SIZE=%q

for cmd in pvcreate vgcreate vgextend lvcreate lvremove mkfs; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "required command $cmd is not installed" >&2
    exit 1
  fi
done

PV_DEVS=()
for DEVICE in "${DEVICES[@]}"; do
  ensure_block_device "$DEVICE"
  refuse_system_disk "$DEVICE"
  resolve_target_device
  ensure_block_device "$TARGET_DEV"
  refuse_mounted_device "$TARGET_DEV"

  CURRENT_VG=$(pvs --noheadings -o vg_name "$TARGET_DEV" 2>/dev/null | awk '{$1=$1;print}' | head -1 || true)
  if [ -n "$CURRENT_VG" ] && [ "$CURRENT_VG" != "$VG_NAME" ]; then
    echo "device $TARGET_DEV already belongs to volume group $CURRENT_VG" >&2
    exit 1
  fi

  if pvs --noheadings "$TARGET_DEV" >/dev/null 2>&1; then
    echo "device $TARGET_DEV is already a physical volume"
  else
    echo "creating physical volume on $TARGET_DEV"
    pvcreate -y "$TARGET_DEV"
  fi
  PV_DEVS+=("$TARGET_DEV")
done

if vgs --noheadings "$VG_NAME" >/dev/null 2>&1; then
  echo "volume group $VG_NAME already exists, extending with new devices"
  for PV_DEV in "${PV_DEVS[@]}"; do
    CURRENT_VG=$(pvs --noheadings -o vg_name "$PV_DEV" 2>/dev/null | awk '{$1=$1;print}' | head -1 || true)
    if [ -z "$CURRENT_VG" ]; then
      echo "adding $PV_DEV to volume group $VG_NAME"
      vgextend "$VG_NAME" "$PV_DEV"
    fi
  done
else
  echo "creating volume group $VG_NAME with ${#PV_DEVS[@]} physical volume(s)"
  vgcreate "$VG_NAME" "${PV_DEVS[@]}"
fi

LV_PATH="/dev/$VG_NAME/$LV_NAME"
if lvs --noheadings "$VG_NAME/$LV_NAME" >/dev/null 2>&1; then
  if [ "$OVERWRITE" != true ]; then
    echo "logical volume $VG_NAME/$LV_NAME already exists"
  else
    echo "removing existing logical volume $VG_NAME/$LV_NAME"
    lvremove -y "$VG_NAME/$LV_NAME"
    if [[ "$LV_SIZE" =~ ^[0-9]+%% ]] || [[ "$LV_SIZE" =~ ^[0-9]+$ ]]; then
      lvcreate -y -n "$LV_NAME" -l "$LV_SIZE" "$VG_NAME"
    else
      lvcreate -y -n "$LV_NAME" -L "$LV_SIZE" "$VG_NAME"
    fi
  fi
else
  echo "creating logical volume $LV_NAME in $VG_NAME"
  if [[ "$LV_SIZE" =~ ^[0-9]+%% ]] || [[ "$LV_SIZE" =~ ^[0-9]+$ ]]; then
    lvcreate -y -n "$LV_NAME" -l "$LV_SIZE" "$VG_NAME"
  else
    lvcreate -y -n "$LV_NAME" -L "$LV_SIZE" "$VG_NAME"
  fi
fi

FORMAT_DEV="$LV_PATH"
ensure_block_device "$FORMAT_DEV"
refuse_mounted_device "$FORMAT_DEV"

if blkid "$FORMAT_DEV" >/dev/null 2>&1; then
  if [ "$OVERWRITE" != true ]; then
    echo "filesystem already exists on $FORMAT_DEV, skipping mkfs"
  else
    echo "overwriting existing filesystem on $FORMAT_DEV"
    mkfs -t "$FILESYSTEM" "$FORMAT_DEV"
  fi
else
  echo "creating $FILESYSTEM filesystem on $FORMAT_DEV"
  mkfs -t "$FILESYSTEM" "$FORMAT_DEV"
fi

mount_and_persist "$FORMAT_DEV"
`, buildDeviceResolveScript(), quoteBashWords(item.Devices), item.Filesystem, item.Partition, item.Overwrite, item.MountPoint,
		formatMountOptions(item.MountOptions), item.LVM.VGName, item.LVM.LVName, item.LVM.LVSize), nil
}
