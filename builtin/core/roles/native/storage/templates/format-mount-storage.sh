#!/bin/bash
# Abort on errors, unset variables, and failed pipes.
set -euo pipefail

# Helper functions. Predicates return 0 for true / 1 for false.
# Return 0 if the given block device is itself a partition.
is_partition_device() { [ "$(lsblk -dn -o TYPE "$1" 2>/dev/null | head -1)" = "part" ]; }

# Resolve the device to be formatted. When DEVICE is already a partition or PARTITION is
# disabled, TARGET_DEV stays DEVICE; otherwise create a single GPT partition and point
# TARGET_DEV at partition 1 (p1 for nvme, otherwise 1).
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

# Abort if the device (or its parent disk) hosts the root filesystem, so we never
# format or destroy the OS disk.
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

# Return 0 only when MOUNT_POINT is fully satisfied by the current configuration:
# it is mounted, from the expected device/LV, and with the expected filesystem.
# Any mismatch means the item must be reprocessed (the idempotent no-op guard).
mount_point_already_used() {
  [ -n "$MOUNT_POINT" ] || return 1
  mountpoint -q "$MOUNT_POINT" || return 1

  local src expected actual_fs
  src=$(findmnt -n -o SOURCE "$MOUNT_POINT" 2>/dev/null || true)
  [ -n "$src" ] || return 1

  # Resolve both the mounted source and the expected target to canonical device
  # paths. LVM mounts report /dev/mapper/<vg>-<lv> or /dev/dm-N while the configured
  # target is /dev/<vg>/<lv>; a whole disk and its partition also resolve
  # differently. Comparing resolved paths avoids false mismatches that would
  # otherwise force an unmount/remount on every run.
  if [ -n "$VG_NAME" ]; then
    expected="/dev/$VG_NAME/$LV_NAME"
  else
    expected="$DEVICE"
    if [ "$PARTITION" = true ] && ! is_partition_device "$DEVICE"; then
      if [[ "$DEVICE" == *"nvme"* ]]; then
        expected="${DEVICE}p1"
      else
        expected="${DEVICE}1"
      fi
    fi
  fi
  [ "$(readlink -f "$src")" = "$(readlink -f "$expected")" ] || return 1

  # The filesystem must match the requested type; otherwise the disk must be rebuilt.
  actual_fs=$(findmnt -n -o FSTYPE "$MOUNT_POINT" 2>/dev/null || true)
  [ "$actual_fs" = "$FILESYSTEM" ] || return 1

  # The mount options must match the requested options; otherwise the mount must
  # be re-created with the correct options. We compare against the fstab entry we
  # persist (column 4) rather than the live kernel options, because the kernel
  # normalizes options ("defaults" expands, "relatime" is added, etc.) which
  # would otherwise cause a false mismatch and force a remount on every run.
  if [ -f /etc/fstab ]; then
    fstab_opts=$(awk -v mp="$MOUNT_POINT" '$2 == mp {print $4; exit}' /etc/fstab 2>/dev/null || true)
    [ "$fstab_opts" = "$MOUNT_OPTIONS" ] || return 1
  else
    return 1
  fi

  return 0
}

# Return 0 if a physical volume (PV) of any volume group is detected on the device or its partitions.
device_has_lvm_allocation() {
  local dev="$1" name pv
  command -v pvs >/dev/null 2>&1 || return 1
  while IFS= read -r name; do
    [ -n "$name" ] || continue
    pv="/dev/$name"
    if pvs --noheadings "$pv" >/dev/null 2>&1; then
      return 0
    fi
  done < <(lsblk -nr -o NAME "$dev" 2>/dev/null || true)
  return 1
}

# Return 0 if the device is already in use: currently mounted somewhere, has a filesystem,
# or is already an LVM physical volume. Used to refuse clobbering existing data.
device_has_existing_allocation() {
  local dev="$1" allocated
  [ -b "$dev" ] || return 1

  if findmnt -rn -S "$dev" >/dev/null 2>&1; then
    return 0
  fi

  allocated=$({ lsblk -nr -o FSTYPE,MOUNTPOINT "$dev" 2>/dev/null || true; } | awk '$1 != "" || $2 != "" { found = 1 } END { print found + 0 }')
  if [ "$allocated" -gt 0 ]; then
    return 0
  fi

  if device_has_lvm_allocation "$dev"; then
    return 0
  fi

  return 1
}

# Derive the force flag for mkfs so that an existing filesystem can be overwritten
# instead of failing with "contains an existing filesystem". Different filesystems
# use different flags: xfs/btrfs need -f, ext2/ext3/ext4 need -F. When MKFS_FORCE is
# empty (unquoted expansion drops it), mkfs runs without a force flag, which is fine
# for a freshly created empty device.
derive_mkfs_force() {
  case "$FILESYSTEM" in
    xfs|btrfs|vxfs) MKFS_FORCE="-f" ;;
    ext2|ext3|ext4) MKFS_FORCE="-F" ;;
    *) MKFS_FORCE="" ;;
  esac
}

# Remove from /etc/fstab any entries referencing the given device, its UUID, or the mount point.
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

# Force-unmount every mount backed by the device (lazy umount) and drop its fstab entry.
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

# Mount $dev at MOUNT_POINT and persist it in /etc/fstab (by UUID when available).
mount_and_persist() {
  local dev="$1"
  [ -n "$MOUNT_POINT" ] || return 0
  mkdir -p "$MOUNT_POINT"
  cleanup_fstab_entries "$dev" "$MOUNT_POINT"
  # Drop any stale mount on the target point before (re)mounting. A previous
  # run may have left the point attached to an old device/LV, and a lazy umount
  # does not release it instantly, so we wait until the point is actually free.
  if mountpoint -q "$MOUNT_POINT"; then
    umount -lf "$MOUNT_POINT"
    for _i in 1 2 3 4 5; do
      mountpoint -q "$MOUNT_POINT" || break
      udevadm settle 2>/dev/null || sleep 1
    done
  fi
  if mountpoint -q "$MOUNT_POINT"; then
    echo "mount point $MOUNT_POINT is still mounted after unmount, cannot proceed" >&2
    exit 1
  fi
  if ! mount -o "$MOUNT_OPTIONS" "$dev" "$MOUNT_POINT"; then
    echo "failed to mount $dev at $MOUNT_POINT" >&2
    exit 1
  fi
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

# Tear down an LVM volume group on pv_dev: unmount and remove its LVs, then remove the VG and PV.
# Refuses if the VG spans more than one PV (would lose data on other disks).
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

# Process a single storage_disks item: derive the LV name if needed, skip when already
# satisfied, then either format+mount the device directly (VG_NAME empty) or build an
# LVM volume group / logical volume and mount it.
process_item() {
  # Derive an LV name when LVM is used but none was specified.
  if [ -n "$VG_NAME" ] && [ -z "$LV_NAME" ]; then
    local _mp_last="${MOUNT_POINT##*/}"
    [ -z "$_mp_last" ] && _mp_last="data"
    LV_NAME="lv_$(basename "$DEVICE")_$_mp_last"
  fi

  if mount_point_already_used; then
    echo "mount point $MOUNT_POINT is already satisfied, skipping storage item"
    return 0
  fi

  # When overwrite is disabled, refuse to rebuild a mount point that is already
  # served by the expected device but with a mismatched filesystem or mount
  # options. We detect this by the live mount source rather than blkid, because
  # blkid can be unreliable on a mounted device and would otherwise let us
  # silently clobber an existing filesystem that the strict idempotent check
  # already flagged as not matching.
  if [ "$OVERWRITE" != true ] && mountpoint -q "$MOUNT_POINT"; then
    local _src _expected_src
    _src=$(findmnt -n -o SOURCE "$MOUNT_POINT" 2>/dev/null || true)
    if [ -n "$VG_NAME" ]; then
      _expected_src="/dev/$VG_NAME/$LV_NAME"
    else
      _expected_src="$DEVICE"
    fi
    if [ -n "$_src" ] && [ "$(readlink -f "$_src")" = "$(readlink -f "$_expected_src")" ]; then
      echo "mount point $MOUNT_POINT is already served by $DEVICE but with different filesystem/options; enable overwrite to rebuild" >&2
      exit 1
    fi
  fi

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

    # The idempotent guard at the top skips the fully-satisfied case. Here we
    # (re)build only what is missing, while refusing to clobber a device that is
    # in use by, or allocated to, something other than this item unless overwrite
    # is enabled.
    if device_has_existing_allocation "$TARGET_DEV"; then
      mnt=$(findmnt -rn -S "$TARGET_DEV" -o TARGET 2>/dev/null || true)
      if [ -n "$mnt" ] && [ "$mnt" != "$MOUNT_POINT" ]; then
        if [ "$OVERWRITE" = true ]; then
          force_unmount_device "$TARGET_DEV"
        else
          echo "device $TARGET_DEV is mounted at $mnt (not $MOUNT_POINT); enable overwrite to take it over" >&2
          exit 1
        fi
      elif [ -z "$mnt" ]; then
        existing_fs=$(blkid -o VALUE -s TYPE "$TARGET_DEV" 2>/dev/null || true)
        if [ "$existing_fs" != "$FILESYSTEM" ] && [ "$OVERWRITE" != true ]; then
          echo "device $TARGET_DEV is already allocated and not formatted as $FILESYSTEM; enable overwrite to rebuild" >&2
          exit 1
        fi
      fi
    fi

    existing_fs=$(blkid -o VALUE -s TYPE "$TARGET_DEV" 2>/dev/null || true)
    if [ -n "$existing_fs" ]; then
      if [ "$existing_fs" = "$FILESYSTEM" ]; then
        echo "filesystem $FILESYSTEM already exists on $TARGET_DEV, keeping it"
      elif [ "$OVERWRITE" = true ]; then
        echo "overwriting existing filesystem ($existing_fs) on $TARGET_DEV with $FILESYSTEM"
        force_unmount_device "$TARGET_DEV"
        derive_mkfs_force
        mkfs -t "$FILESYSTEM" $MKFS_FORCE "$TARGET_DEV"
      else
        echo "existing filesystem ($existing_fs) on $TARGET_DEV does not match requested $FILESYSTEM; enable overwrite to rebuild" >&2
        exit 1
      fi
    else
      echo "creating $FILESYSTEM filesystem on $TARGET_DEV"
      force_unmount_device "$TARGET_DEV"
      derive_mkfs_force
      mkfs -t "$FILESYSTEM" $MKFS_FORCE "$TARGET_DEV"
    fi
    mount_and_persist "$TARGET_DEV"
  else
    # LVM path.
    # Bail out early if the LVM2 user-space tools are missing on the node.
    for cmd in pvcreate pvremove pvs vgcreate vgextend vgremove vgs lvcreate lvremove lvs mkfs wipefs; do
      command -v "$cmd" >/dev/null 2>&1 || { echo "required command $cmd is not installed" >&2; exit 1; }
    done
    PV_DEVS=()
    for DEVICE in "${DEVICES[@]}"; do
      [ -b "$DEVICE" ] || { echo "block device $DEVICE does not exist" >&2; exit 1; }
      refuse_system_disk "$DEVICE"
      resolve_target_device
      [ -b "$TARGET_DEV" ] || { echo "target device $TARGET_DEV does not exist" >&2; exit 1; }
      # Determine the volume group the device currently belongs to (empty if none).
      # Computed unconditionally so the re-run/overwrite branches below always have a value.
      CURRENT_VG=$(pvs --noheadings -o vg_name "$TARGET_DEV" 2>/dev/null | awk '{$1=$1;print}' | head -1 || true)
      if device_has_existing_allocation "$TARGET_DEV"; then
        # A PV belonging to a different volume group without overwrite is a
        # hard conflict and must not be touched. Our own PV (already part of
        # the target VG) falls through so the LV/mount strict checks below can
        # still re-create the filesystem or remount when fs/options drift.
        if [ -n "$CURRENT_VG" ] && [ "$CURRENT_VG" != "$VG_NAME" ] && [ "$OVERWRITE" != true ]; then
          echo "device $TARGET_DEV already belongs to volume group $CURRENT_VG" >&2
          exit 1
        fi
      fi
      force_unmount_device "$TARGET_DEV"
      # Only a PV belonging to a *foreign* volume group needs handling here.
      # Our own PV (CURRENT_VG == VG_NAME) is reused as-is; the LV/mount strict
      # checks further down still re-create the filesystem or remount on drift.
      if [ "$CURRENT_VG" != "$VG_NAME" ]; then
        if [ "$OVERWRITE" = true ]; then
          echo "clearing existing volume group $CURRENT_VG on $TARGET_DEV"
          clear_lvm_device "$CURRENT_VG" "$TARGET_DEV"
        else
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
        pv_vg=$(pvs --noheadings -o vg_name "$PV_DEV" 2>/dev/null | awk '{$1=$1;print}' | head -1 || true)
        if [ -z "$pv_vg" ]; then
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
    existing_fs=$(blkid -o VALUE -s TYPE "$FORMAT_DEV" 2>/dev/null || true)
    if [ -n "$existing_fs" ]; then
      if [ "$existing_fs" = "$FILESYSTEM" ]; then
        echo "filesystem $FILESYSTEM already exists on $FORMAT_DEV, keeping it"
      elif [ "$OVERWRITE" = true ]; then
        echo "overwriting existing filesystem ($existing_fs) on $FORMAT_DEV with $FILESYSTEM"
        derive_mkfs_force
        mkfs -t "$FILESYSTEM" $MKFS_FORCE "$FORMAT_DEV"
      else
        echo "existing filesystem ($existing_fs) on $FORMAT_DEV does not match requested $FILESYSTEM; enable overwrite to rebuild" >&2
        exit 1
      fi
    else
      echo "creating $FILESYSTEM filesystem on $FORMAT_DEV"
      derive_mkfs_force
      mkfs -t "$FILESYSTEM" $MKFS_FORCE "$FORMAT_DEV"
    fi
    mount_and_persist "$FORMAT_DEV"
  fi
}

# Process storage disks (format + mount, optionally via LVM).
{{ if .kubernetes.storage_disks }}
{{ range .kubernetes.storage_disks }}
# Normalize the configured device to an absolute /dev path.
DEVICE="{{ index . "device" }}"
[ -n "$DEVICE" ] || { echo "device is required" >&2; exit 1; }
DEVICE="${DEVICE#/dev/}"
DEVICE="/dev/$DEVICE"
DEVICES=("$DEVICE")

FILESYSTEM="{{ index . "filesystem" | default "xfs" }}"
PARTITION={{ index . "partition" | default false }}
OVERWRITE={{ index . "overwrite" | default true }}
MOUNT_POINT="{{ index . "mountpoint" | default "" }}"
MOUNT_OPTIONS="{{ index . "mount_option" | default "defaults" }}"
VG_NAME="{{ index . "vg_name" | default "" }}"
LV_NAME="{{ index . "lv_name" | default "" }}"
LV_SIZE="100%FREE"

process_item
{{ end }}
{{ end }}
