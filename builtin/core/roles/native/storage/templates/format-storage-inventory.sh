#!/bin/bash
set -euo pipefail

# Helper functions.
is_partition_device() { [ "$(lsblk -dn -o TYPE "$1" 2>/dev/null | head -1)" = "part" ]; }

resolve_target_device() {
  TARGET_DEV="$DEVICE"
  is_partition_device "$DEVICE" && return 0
  [ "$PARTITION" = true ] || return 0
  PART_COUNT=$(lsblk -ln -o TYPE "$DEVICE" 2>/dev/null | grep -c '^part$' || true)
  if [ "$PART_COUNT" -eq 0 ]; then
    parted -s "$DEVICE" mklabel gpt
    parted -s -a optimal "$DEVICE" mkpart primary 0% 100%
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
  local dev="$1" pkname
  pkname=$(lsblk -no PKNAME "$dev" 2>/dev/null | head -1 || true)
  [ -z "$pkname" ] && pkname=$(basename "$dev")
  ROOT_SOURCE=$(findmnt -n -o SOURCE / 2>/dev/null || true)
  [ -z "$ROOT_SOURCE" ] && return 0
  ROOT_DISK=$(lsblk -no PKNAME "$ROOT_SOURCE" 2>/dev/null | head -1 || true)
  [ -z "$ROOT_DISK" ] && ROOT_DISK=$(basename "$ROOT_SOURCE")
  if [ -n "$ROOT_DISK" ] && [ "$pkname" = "$ROOT_DISK" ]; then
    echo "refusing to use system disk $dev" >&2
    exit 1
  fi
}

cleanup_fstab_entries() {
  local dev="$1" mp="${2:-}" uuid="" tmp
  uuid=$(blkid -s UUID -o value "$dev" 2>/dev/null || true)
  [ -f /etc/fstab ] || return 0
  tmp=$(mktemp)
  awk -v dev="$dev" -v uuid="$uuid" -v mp="$mp" '
    /^[[:space:]]*#/ || NF == 0 { print; next }
    $1 == dev { next }
    uuid != "" && $1 == "UUID=" uuid { next }
    mp != "" && $2 == mp { next }
    { print }
  ' /etc/fstab > "$tmp"
  cat "$tmp" > /etc/fstab
  rm -f "$tmp"
}

force_unmount_device() {
  local dev="$1" target
  local targets
  targets=$(findmnt -rn -S "$dev" -o TARGET 2>/dev/null || true)
  [ -z "$targets" ] && return 0
  while IFS= read -r target; do
    [ -z "$target" ] && continue
    echo "force unmounting $dev from $target"
    umount -lf "$target"
    cleanup_fstab_entries "$dev" "$target"
  done <<< "$targets"
}

mount_and_persist() {
  local dev="$1"
  [ -n "$MOUNT_POINT" ] || return 0
  mkdir -p "$MOUNT_POINT"
  cleanup_fstab_entries "$dev" "$MOUNT_POINT"
  mountpoint -q "$MOUNT_POINT" && umount -lf "$MOUNT_POINT"
  mountpoint -q "$MOUNT_POINT" && { echo "mount point $MOUNT_POINT is still mounted" >&2; exit 1; }
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

clear_lvm_device() {
  local vg_name="$1" pv_dev="$2" pv_count lv_name
  if ! vgs --noheadings "$vg_name" >/dev/null 2>&1; then
    pvremove -ff -y "$pv_dev" 2>/dev/null || true
    wipefs -a "$pv_dev" 2>/dev/null || true
    return 0
  fi
  pv_count=$(pvs --noheadings -o pv_name,vg_name 2>/dev/null | awk -v vg="$vg_name" '$2 == vg {count++} END {print count+0}')
  if [ "$pv_count" -gt 1 ]; then
    echo "volume group $vg_name contains multiple physical volumes, refusing to clear only $pv_dev" >&2
    exit 1
  fi
  while IFS= read -r lv_name; do
    [ -z "$lv_name" ] && continue
    force_unmount_device "/dev/$vg_name/$lv_name"
    lvremove -y "$vg_name/$lv_name"
  done < <(lvs --noheadings -o lv_name "$vg_name" 2>/dev/null | awk '{$1=$1;print}')
  vgremove -y "$vg_name"
  pvremove -ff -y "$pv_dev" 2>/dev/null || true
  wipefs -a "$pv_dev" 2>/dev/null || true
}

process_item() {
  if [ -z "$VG_NAME" ]; then
    # Direct formatting: exactly one device.
    if [ ${#DEVICES[@]} -ne 1 ]; then
      echo "direct formatting requires exactly one device" >&2
      exit 1
    fi
    DEVICE="${DEVICES[0]}"
    [ -b "$DEVICE" ] || { echo "block device $DEVICE does not exist" >&2; exit 1; }
    refuse_system_disk "$DEVICE"
    resolve_target_device
    [ -b "$TARGET_DEV" ] || { echo "target device $TARGET_DEV does not exist" >&2; exit 1; }
    force_unmount_device "$TARGET_DEV"
    FORMAT_DEV="$TARGET_DEV"
    if blkid "$FORMAT_DEV" >/dev/null 2>&1; then
      if [ "$OVERWRITE" != true ]; then
        echo "filesystem already exists on $FORMAT_DEV, skipping"
        return 0
      fi
      echo "overwriting existing filesystem on $FORMAT_DEV"
    else
      echo "creating $FILESYSTEM filesystem on $FORMAT_DEV"
    fi
    mkfs -t "$FILESYSTEM" "$FORMAT_DEV"
    mount_and_persist "$FORMAT_DEV"
  else
    # LVM path.
    for cmd in pvcreate pvremove pvs vgcreate vgextend vgremove vgs lvcreate lvremove lvs mkfs wipefs; do
      command -v "$cmd" >/dev/null 2>&1 || { echo "required command $cmd is not installed" >&2; exit 1; }
    done
    PV_DEVS=()
    for DEVICE in "${DEVICES[@]}"; do
      [ -b "$DEVICE" ] || { echo "block device $DEVICE does not exist" >&2; exit 1; }
      refuse_system_disk "$DEVICE"
      resolve_target_device
      [ -b "$TARGET_DEV" ] || { echo "target device $TARGET_DEV does not exist" >&2; exit 1; }
      force_unmount_device "$TARGET_DEV"
      CURRENT_VG=$(pvs --noheadings -o vg_name "$TARGET_DEV" 2>/dev/null | awk '{$1=$1;print}' | head -1 || true)
      if [ -n "$CURRENT_VG" ]; then
        if [ "$OVERWRITE" = true ]; then
          echo "clearing existing volume group $CURRENT_VG on $TARGET_DEV"
          clear_lvm_device "$CURRENT_VG" "$TARGET_DEV"
        elif [ "$CURRENT_VG" != "$VG_NAME" ]; then
          echo "device $TARGET_DEV already belongs to volume group $CURRENT_VG" >&2
          exit 1
        fi
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
        force_unmount_device "$LV_PATH"
        lvremove -y "$VG_NAME/$LV_NAME"
        if [[ "$LV_SIZE" =~ ^[0-9]+% ]] || [[ "$LV_SIZE" =~ ^[0-9]+$ ]]; then
          lvcreate -y -n "$LV_NAME" -l "$LV_SIZE" "$VG_NAME"
        else
          lvcreate -y -n "$LV_NAME" -L "$LV_SIZE" "$VG_NAME"
        fi
      fi
    else
      echo "creating logical volume $LV_NAME in $VG_NAME"
      if [[ "$LV_SIZE" =~ ^[0-9]+% ]] || [[ "$LV_SIZE" =~ ^[0-9]+$ ]]; then
        lvcreate -y -n "$LV_NAME" -l "$LV_SIZE" "$VG_NAME"
      else
        lvcreate -y -n "$LV_NAME" -L "$LV_SIZE" "$VG_NAME"
      fi
    fi
    FORMAT_DEV="$LV_PATH"
    [ -b "$FORMAT_DEV" ] || { echo "logical volume $FORMAT_DEV does not exist" >&2; exit 1; }
    force_unmount_device "$FORMAT_DEV"
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
  fi
}

# Process inventory standard storage items.
{{ if .storage }}
{{ range .storage }}
DEVICES=()
{{ $disk := index . "disk" | default "" }}{{ if $disk }}DEVICES+=("{{ $disk }}"){{ end }}
{{ $device := index . "device" | default "" }}
{{ if kindIs "string" $device }}{{ if $device }}DEVICES+=("{{ $device }}"){{ end }}{{ else }}{{ range $device }}DEVICES+=("{{ . }}"){{ end }}{{ end }}
for i in "${!DEVICES[@]}"; do
  d="${DEVICES[$i]}"
  d="${d#/dev/}"
  DEVICES[$i]="/dev/$d"
done

FILESYSTEM="{{ index . "filesystem" }}"
PARTITION={{ index . "partition" | default false }}
OVERWRITE={{ index . "overwrite" | default false }}
MOUNT_POINT="{{ index . "mount_point" | default "" }}"
MOUNT_OPTIONS="{{ index . "mount_options" | default "defaults" }}"
VG_NAME="{{ with index . "lvm" }}{{ index . "vg_name" | default "" }}{{ end }}"
LV_NAME="{{ with index . "lvm" }}{{ index . "lv_name" | default "" }}{{ end }}"
LV_SIZE="{{ with index . "lvm" }}{{ index . "lv_size" | default "100%FREE" }}{{ end }}"

[ -n "$MOUNT_OPTIONS" ] || MOUNT_OPTIONS="defaults"
if [ -z "$LV_NAME" ] && [ -n "$VG_NAME" ]; then
  name="${MOUNT_POINT#/}"
  name="${name:-${DEVICES[0]#/dev/}}"
  name="${name//\//_}"
  name="${name//-/_}"
  LV_NAME="lv_${name:-data}"
fi

process_item
{{ end }}
{{ end }}
