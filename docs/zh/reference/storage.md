# 存储与 Multipath 配置

本文档说明 KubeKey 在 Inventory 中配置节点磁盘格式化与 multipath 黑名单的字段规范。

相关实现：

- 磁盘格式化模块：`pkg/modules/storage`
- Multipath 配置模块：`pkg/modules/multipath_conf`
- Multipath 执行角色：`builtin/core/roles/native/multipath`
- Storage 执行角色：`builtin/core/roles/native/storage`
- 预检查：`builtin/core/roles/precheck/storage`

当主机配置了 `storage` 或 `kubernetes.storage_disks` 时，`native` 角色会在其他初始化步骤之前执行磁盘格式化；`precheck` 阶段会校验配置合法性。

---

## 配置入口

KubeKey 支持两种磁盘配置方式，**同一主机只需使用其中一种**：

| 方式 | 字段路径 | 适用场景 |
|------|----------|----------|
| Inventory 标准格式 | `spec.hosts.<name>.storage` | 手写 Inventory、CLI 安装 |
| Web 安装器格式 | `spec.hosts.<name>.kubernetes.storage_disks` | Web 安装器节点存储配置 |

Multipath 配置统一写在 `spec.hosts.<name>.kubernetes.multipath` 下，与上述两种方式可组合使用。

---

## Multipath 配置

在启用 multipath 的环境中，本地数据盘可能被 multipath 错误识别。KubeKey 会在目标节点向 `/etc/multipath.conf` 的 `blacklist` 段追加 `devnode` 规则，将本地磁盘排除在 multipath 管理之外。

### 示例

```yaml
spec:
  hosts:
    node1:
      kubernetes:
        multipath:
          enabled: true
          path: /etc/multipath.conf   # 可选
          backup: true                # 可选
          reload: true                  # 可选
          devnodes:                   # 可选
            - ^sd[a-z]
            - ^vd[a-z]
            - ^xvd[a-z]
            - ^nvme[0-9]n[0-9]
```

### 参数说明

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `kubernetes.multipath.enabled` | Boolean | 是 | `false` | 为 `true` 时执行 multipath 黑名单配置 |
| `kubernetes.multipath.path` | String | 否 | `/etc/multipath.conf` | multipath 配置文件路径 |
| `kubernetes.multipath.backup` | Boolean | 否 | `true` | 修改前是否备份原配置文件 |
| `kubernetes.multipath.reload` | Boolean | 否 | `true` | 配置后是否 reload/restart `multipathd` |
| `kubernetes.multipath.devnodes` | String[] | 否 | 见下方 | 写入 `blacklist` 的 `devnode` 正则规则列表 |

默认 `devnodes` 规则：

```yaml
devnodes:
  - ^sd[a-z]
  - ^vd[a-z]
  - ^xvd[a-z]
  - ^nvme[0-9]n[0-9]
```

### 行为说明

- 若配置文件不存在，会创建仅含空 `blacklist {}` 块的新文件。
- 若已有 `blacklist` 块，仅追加缺失的 `devnode` 规则，不重复写入。
- 修改后会执行 `multipath -t` 校验配置；若存在 `multipathd` 服务则尝试 reload。
- `devnodes` 至少包含一条非空规则，否则模块报错。

---

## Storage 配置（Inventory 标准格式）

在 Inventory 主机变量下通过 `storage` 数组声明多块磁盘，每项描述一次格式化/挂载操作。

### 示例

```yaml
spec:
  hosts:
    node1:
      storage:
        # 多盘 LVM 聚合后挂载
        - device:
            - /dev/sdb1
            - /dev/sdc1
          lvm:
            vg_name: vg_data
            lv_name: lv_data
            lv_size: 100%FREE    # 可选
          filesystem: xfs
          mount_point: /data
          mount_options: prjquota
        # 整盘分区后格式化
        - disk: sdd              # 与 device 二选一或组合使用
          partition: true
          lvm:
            vg_name: vg_data
            lv_name: lv_cache
          filesystem: ext4
          mount_point: /cache
          overwrite: false
```

### 参数说明

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `device` | String / String[] | 与 `disk` 至少填一个 | - | 块设备路径，如 `/dev/sdb1`、`sdc`；支持 JSON 数组字符串 |
| `disk` | String | 与 `device` 至少填一个 | - | 磁盘名（不含 `/dev/` 前缀），如 `sdb`、`nvme0n1` |
| `filesystem` | String | 是 | - | 文件系统类型，仅支持 `ext4`、`xfs`、`btrfs` |
| `partition` | Boolean | 否 | `false` | 对整盘自动创建 GPT 单分区后再格式化 |
| `overwrite` | Boolean | 否 | `false` | 设备上已有文件系统时是否强制覆盖 |
| `mount_point` | String | 否 | - | 挂载点；为空则只格式化不挂载 |
| `mount_options` | String | 否 | `defaults` | 写入 `/etc/fstab` 的挂载选项；未含 `defaults` 时会自动前缀 `defaults,` |
| `lvm` | Object | 否 | - | 启用 LVM 管理；配置后允许多个 `device` |
| `lvm.vg_name` | String | 启用 LVM 时必填 | - | 卷组名，仅允许 `[a-zA-Z0-9._+-]` |
| `lvm.lv_name` | String | 否 | 自动推导 | 逻辑卷名；默认根据 `mount_point` 生成，如 `/data` → `lv_data` |
| `lvm.lv_size` | String | 否 | `100%FREE` | 逻辑卷大小，如 `10G`、`50%FREE`、`100%FREE` |

### 校验规则（precheck）

- 每项至少指定 `disk` 或 `device` 之一。
- `filesystem` 必填，且必须是 `ext4`、`xfs`、`btrfs` 之一（不区分大小写）。
- 若配置了 `lvm`，则 `lvm.vg_name` 不能为空；未配置 `lvm.lv_name` 时会自动推导。

### 执行行为

- **直接格式化**（未配置 `lvm`）：`device`/`disk` 只能指定**一个**设备。
- **LVM 模式**：可对多个物理卷创建/扩展同一卷组，再创建逻辑卷并格式化。
- **分区**：`partition: true` 且目标为整盘时，自动 `mklabel gpt` 并创建单分区；NVMe 设备分区名为 `nvme0n1p1`，其他设备为 `sdb1` 形式。
- **安全保护**：拒绝格式化系统根盘所在磁盘。
- **幂等**：已有文件系统且 `overwrite: false` 时跳过 `mkfs`，但仍会尝试挂载。

---

## Storage 配置（Web 安装器格式）

Web 安装器将磁盘配置写入 `kubernetes.storage_disks`，字段名与 Inventory 标准格式略有不同，由同一 `storage` 模块解析。

### 示例

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

### 参数说明

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `device` | String | 是 | - | 块设备，可写 `vdc` 或 `/dev/vdc` |
| `mountpoint` | String | 是 | - | 挂载点（对应标准格式的 `mount_point`） |
| `filesystem` | String | 否 | `xfs` | 文件系统类型，仅支持 `ext4`、`xfs`、`btrfs` |
| `partition` | Boolean | 否 | `false` | 是否先创建分区 |
| `overwrite` | Boolean | 否 | `true` | 是否覆盖已有文件系统 |
| `mount_option` | String | 否 | `defaults` | 挂载选项（对应标准格式的 `mount_options`） |
| `vg_name` | String | 否 | - | 配置后启用 LVM；未写 `lv_name` 时自动推导 |
| `lv_name` | String | 否 | 自动推导 | 逻辑卷名 |

### 字段映射关系

| Web 安装器 (`storage_disks`) | Inventory 标准格式 (`storage`) |
|------------------------------|--------------------------------|
| `mountpoint` | `mount_point` |
| `mount_option` | `mount_options` |
| `vg_name` / `lv_name`（顶层） | `lvm.vg_name` / `lvm.lv_name`（嵌套） |

### 校验规则（precheck）

- `device` 与 `mountpoint` 必填。
- `filesystem` 若填写，必须是 `ext4`、`xfs`、`btrfs` 之一。

---

## 完整 Inventory 示例

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

## 注意事项

1. **数据破坏性操作**：`overwrite: true` 或首次格式化会清除目标设备数据，请确认设备路径正确。
2. **系统盘保护**：模块会检测并拒绝操作根文件系统所在磁盘，但仍建议事先用 `lsblk` 核对。
3. **LVM 依赖**：LVM 模式要求目标节点已安装 `lvm2` 相关命令（`pvcreate`、`vgcreate`、`lvcreate` 等）。
4. **Multipath 与存储顺序**：`native` 角色依赖中 `native/multipath` 在 `native/storage` 之前执行；仅开 multipath 时也会独立运行，不再依赖 storage 配置。
5. **仅配置其一**：`storage` 与 `kubernetes.storage_disks` 面向不同入口，同一主机通常只配置一种；`multipath` 可单独启用。
