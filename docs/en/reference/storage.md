# Storage and Multipath Configuration

This document describes how to configure disk formatting and multipath blacklisting in KubeKey Inventory.

Related implementation:

- Disk formatting role: `builtin/core/roles/native/storage`
- Multipath configuration role: `builtin/core/roles/native/multipath`
- Precheck role: `builtin/core/roles/precheck/storage`

When a host defines `storage` or `kubernetes.storage_disks`, the `native` role formats disks before other initialization steps. The `precheck` role validates the configuration first.

---

## Configuration Entry Points

KubeKey supports two disk configuration formats. **Use only one per host**:

| Format | Field path | Use case |
|--------|------------|----------|
| Inventory standard | `spec.hosts.<name>.storage` | Hand-written Inventory, CLI installs |
| Web installer | `spec.hosts.<name>.kubernetes.storage_disks` | Web installer node storage settings |

Multipath settings are always defined under `spec.hosts.<name>.kubernetes.multipath` and can be combined with either disk format.

---

## Multipath Configuration

In environments with multipath enabled, local data disks may be incorrectly managed by multipath. KubeKey appends `devnode` rules to the `blacklist` section of `/etc/multipath.conf` on target nodes to exclude local disks.

### Example

```yaml
spec:
  hosts:
    node1:
      kubernetes:
        multipath:
          enabled: true
          path: /etc/multipath.conf   # optional
          backup: true                # optional
          reload: true                # optional
          devnodes:                   # optional
            - ^sd[a-z]
            - ^vd[a-z]
            - ^xvd[a-z]
            - ^nvme[0-9]n[0-9]
```

### Parameters

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `kubernetes.multipath.enabled` | Boolean | Yes | `false` | When `true`, configure multipath blacklist rules |
| `kubernetes.multipath.path` | String | No | `/etc/multipath.conf` | Path to multipath configuration file |
| `kubernetes.multipath.backup` | Boolean | No | `true` | Back up the existing config before modification |
| `kubernetes.multipath.reload` | Boolean | No | `true` | Reload or restart `multipathd` after changes |
| `kubernetes.multipath.devnodes` | String[] | No | See below | Regex rules written as `devnode` entries in `blacklist` |

Default `devnodes` rules:

```yaml
devnodes:
  - ^sd[a-z]
  - ^vd[a-z]
  - ^xvd[a-z]
  - ^nvme[0-9]n[0-9]
```

### Behavior

- Creates a new file with an empty `blacklist {}` block if the config file does not exist.
- Appends only missing `devnode` rules to an existing `blacklist` block; does not duplicate rules.
- Runs `multipath -t` to validate the config and reloads `multipathd` when available.
- Fails if `devnodes` is empty after trimming.

---

## Storage Configuration (Inventory Standard Format)

Define multiple disks under the host-level `storage` array. Each item describes one format/mount operation.

### Example

```yaml
spec:
  hosts:
    node1:
      storage:
        # Multiple disks aggregated via LVM
        - device:
            - /dev/sdb1
            - /dev/sdc1
          lvm:
            vg_name: vg_data
            lv_name: lv_data
            lv_size: 100%FREE    # optional
          filesystem: xfs
          mount_point: /data
          mount_options: prjquota
        # Whole disk with automatic partitioning
        - disk: sdd              # Use disk and/or device
          partition: true
          lvm:
            vg_name: vg_data
            lv_name: lv_cache
          filesystem: ext4
          mount_point: /cache
          overwrite: false
```

### Parameters

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `device` | String / String[] | At least one of `disk` or `device` | - | Block device path, e.g. `/dev/sdb1`, `sdc`; JSON array strings are supported |
| `disk` | String | At least one of `disk` or `device` | - | Disk name without `/dev/` prefix, e.g. `sdb`, `nvme0n1` |
| `filesystem` | String | Yes | - | Filesystem type: `ext4`, `xfs`, or `btrfs` only |
| `partition` | Boolean | No | `false` | Create a single GPT partition on a whole disk before formatting |
| `overwrite` | Boolean | No | `false` | Force overwrite when a filesystem already exists |
| `mount_point` | String | No | - | Mount point; if empty, only format without mounting |
| `mount_options` | String | No | `defaults` | Mount options written to `/etc/fstab`; `defaults,` is prepended if `defaults` is missing |
| `lvm` | Object | No | - | Enable LVM management; allows multiple `device` entries |
| `lvm.vg_name` | String | Required when LVM is enabled | - | Volume group name; must match `[a-zA-Z0-9._+-]` |
| `lvm.lv_name` | String | No | Auto-derived | Logical volume name; defaults from `mount_point`, e.g. `/data` → `lv_data` |
| `lvm.lv_size` | String | No | `100%FREE` | Logical volume size, e.g. `10G`, `50%FREE`, `100%FREE` |

### Precheck Rules

- Each item must specify at least one of `disk` or `device`.
- `filesystem` is required and must be one of `ext4`, `xfs`, or `btrfs` (case-insensitive).
- When `lvm` is set, `lvm.vg_name` must be non-empty; `lvm.lv_name` is auto-derived when omitted.

### Runtime Behavior

- **Direct format** (no `lvm`): exactly **one** device is allowed.
- **LVM mode**: multiple physical volumes can be added to the same volume group before creating and formatting a logical volume.
- **Partitioning**: with `partition: true` on a whole disk, creates a GPT label and a single partition; NVMe devices use the `p1` suffix (e.g. `nvme0n1p1`), others use `sdb1` style names.
- **Safety**: refuses to format the disk hosting the root filesystem.
- **Idempotency**: skips `mkfs` when a filesystem already exists and `overwrite` is `false`, but still attempts to mount.

---

## Storage Configuration (Web Installer Format)

The web installer stores disk settings in `kubernetes.storage_disks`. Field names differ slightly from the Inventory standard format, but the same role handles both.

### Example

```yaml
spec:
  hosts:
    node1:
      kubernetes:
        storage_disks:
          - device: vda1
            mountpoint: /data
            filesystem: xfs
            partition: false
            overwrite: true
            mount_option: prjquota
            vg_name: vg_data
            lv_name: lv_data
          - device: vda2
            mountpoint: /data2
            filesystem: xfs
```

### Parameters

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `device` | String | Yes | - | Block device; `vdc` or `/dev/vdc` |
| `mountpoint` | String | Yes | - | Mount point (maps to `mount_point` in the standard format) |
| `filesystem` | String | No | `xfs` | Filesystem type: `ext4`, `xfs`, or `btrfs` only |
| `partition` | Boolean | No | `false` | Create a partition before formatting |
| `overwrite` | Boolean | No | `true` | Overwrite an existing filesystem |
| `mount_option` | String | No | `defaults` | Mount options (maps to `mount_options`) |
| `vg_name` | String | No | - | Enables LVM when set; `lv_name` is auto-derived if omitted |
| `lv_name` | String | No | Auto-derived | Logical volume name |

### Field Mapping

| Web installer (`storage_disks`) | Inventory standard (`storage`) |
|---------------------------------|--------------------------------|
| `mountpoint` | `mount_point` |
| `mount_option` | `mount_options` |
| `vg_name` / `lv_name` (top-level) | `lvm.vg_name` / `lvm.lv_name` (nested) |

### Precheck Rules

- `device` and `mountpoint` are required.
- If `filesystem` is set, it must be one of `ext4`, `xfs`, or `btrfs`.

---

## Full Inventory Example

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
      connector:
        host: 192.168.1.10
        user: root
        private_key: ~/.ssh/id_rsa
      kubernetes:
        multipath:
          enabled: true
        storage_disks:
          - device: vdc
            filesystem: xfs
            mountpoint: /data
            vg_name: vg_data
            lv_name: lv_data
  groups:
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    kube_control_plane:
      hosts:
        - node1
    kube_worker:
      hosts:
        - node1
```

---

## Notes

1. **Destructive operations**: `overwrite: true` or first-time formatting erases data on the target device. Verify device paths with `lsblk` first.
2. **Root disk protection**: The role refuses to operate on the root filesystem disk, but manual verification is still recommended.
3. **LVM dependencies**: LVM mode requires `lvm2` tools (`pvcreate`, `vgcreate`, `lvcreate`, etc.) on the target node.
4. **Execution order**: In the `native` role dependencies, `native/multipath` runs before `native/storage`. Multipath can run on its own and no longer depends on storage configuration.
5. **Use one format**: `storage` and `kubernetes.storage_disks` target different entry points; configure only one per host. `multipath` can be enabled independently.
