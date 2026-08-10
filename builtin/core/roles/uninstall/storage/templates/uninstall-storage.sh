#!/bin/bash
set -uo pipefail

# Remove storage disks provisioned by format-mount-storage.sh.
#
# This script is only generated when delete.data is true, so destroying the
# logical volumes, volume groups and physical volumes and wiping the underlying
# block devices is the intended behaviour: it removes the data together with the
# mounts. Failures are tolerated so cleanup stays idempotent across re-runs.

# Remove fstab entries that reference the given device or mount point.
cleanup_fstab_entries() {
  local dev="$1" mp="${2:-}" uuid="" tmp
  [ -f /etc/fstab ] || return 0
  uuid=$([ -n "$dev" ] && blkid -s UUID -o value "$dev" 2>/dev/null || true)
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

uninstall_item() {
  # Derive the LV name the same way format-mount-storage.sh does, so cleanup
  # matches what was actually created when lv_name was left empty.
  if [ -n "$VG_NAME" ] && [ -z "$LV_NAME" ]; then
    local _mp_last="${MOUNT_POINT##*/}"
    [ -z "$_mp_last" ] && _mp_last="data"
    LV_NAME="lv_$(basename "$DEVICE")_$_mp_last"
  fi

  # Step 1: unmount the mount point and drop its fstab entry.
  if [ -n "$MOUNT_POINT" ]; then
    if mountpoint -q "$MOUNT_POINT"; then
      echo "unmounting $MOUNT_POINT"
      umount -lf "$MOUNT_POINT" || true
    fi
    cleanup_fstab_entries "" "$MOUNT_POINT"
  fi

  if [ -n "$VG_NAME" ]; then
    # LVM path: remove the logical volume, then the whole volume group and its
    # physical volumes once it no longer holds any logical volume.
    if ! command -v lvs >/dev/null 2>&1; then
      echo "lvm2 tools not found, skipping LVM cleanup for $VG_NAME" >&2
      return 0
    fi
    if lvs --noheadings "$VG_NAME/$LV_NAME" >/dev/null 2>&1; then
      echo "removing logical volume $VG_NAME/$LV_NAME"
      lvremove -y "$VG_NAME/$LV_NAME" || true
    fi
    if [ -z "$(lvs --noheadings -o lv_name "$VG_NAME" 2>/dev/null | awk 'NF')" ]; then
      echo "volume group $VG_NAME is empty, removing it and all its physical volumes"
      local pv pvs_list
      # Collect the PVs while the VG still exists; after vgremove the VG-scoped
      # query returns nothing and the PV labels would be left behind.
      # Query all PVs and filter by VG column (the legacy "pvs <vgname>" form is
      # not portable across lvm2 versions and is rejected on some).
      pvs_list=$(pvs --noheadings -o pv_name,vg_name 2>/dev/null | awk -v vg="$VG_NAME" '$2 == vg {print $1}')
      vgremove -y "$VG_NAME" || true
      for pv in $pvs_list; do
        [ -n "$pv" ] || continue
        wipefs -a "$pv" 2>/dev/null || true
        pvremove -ff -y "$pv" 2>/dev/null || true
      done
    fi
  else
    # Direct path: wipe the underlying device to delete its data.
    local dev="${DEVICES[0]:-}"
    if [ -b "$dev" ]; then
      echo "wiping device $dev"
      wipefs -a "$dev" 2>/dev/null || true
    fi
  fi
}

# Remove storage disks (unmount + LVM teardown or device wipe).
{{ if .kubernetes.storage_disks }}
{{ range .kubernetes.storage_disks }}
# Normalize the configured device to an absolute /dev path.
DEVICE="{{ index . "device" }}"
[ -n "$DEVICE" ] || { echo "device is required" >&2; exit 1; }
DEVICE="${DEVICE#/dev/}"
DEVICE="/dev/$DEVICE"
DEVICES=("$DEVICE")

MOUNT_POINT="{{ index . "mountpoint" | default "" }}"
VG_NAME="{{ index . "vg_name" | default "" }}"
LV_NAME="{{ index . "lv_name" | default "" }}"

uninstall_item
{{ end }}
{{ end }}
