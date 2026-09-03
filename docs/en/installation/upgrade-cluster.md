# Upgrade Cluster

This section describes how to use KubeKey to upgrade an existing Kubernetes cluster, including the Kubernetes control plane and worker nodes, and optionally etcd, the container runtime, the CNI plugin, and the StorageClass provisioner.

## Prerequisites

- An existing Kubernetes cluster deployed by KubeKey.
- The target Kubernetes version must be higher than the currently installed version.
- Back up the cluster before upgrading (especially etcd data) so you can recover if the upgrade fails.

> **Note**: Web Installer does not currently support upgrading the cluster. Please use the command line instead.

## Retrieve Current Cluster Configuration Files

If the cluster was installed via the **Web Installer**, you can retrieve the current cluster configuration files as follows.

### Retrieve inventory.yaml

```shell
cp kubekey/runtime/kubekey.kubesphere.io/v1/inventories/default/default.yaml kkv4-inventory.yaml
```

### Retrieve config.yaml

```shell
cat schema/config.json | jq '{spec: .["kubernetes.json"]}' > kkv4-config.json
```

## Upgrade Cluster

By default, `kk upgrade cluster` upgrades only the Kubernetes control plane and worker node binaries (`kubeadm` / `kubelet`); other components must be enabled explicitly.

### Upgrade the Kubernetes Version Only

Set the target version with `--with-kubernetes` (or `kubernetes.kube_version` in `config.yaml`):

```shell
./kk upgrade cluster -i inventory.yaml -c config.yaml --with-kubernetes v1.34.11
```

KubeKey performs a rolling upgrade: the first control plane node runs `kubeadm upgrade apply`, the remaining control plane nodes run `kubeadm upgrade node`, and all worker nodes are upgraded in parallel.

### Upgrade Components Together

Use `--all` to upgrade etcd, the container runtime, the CNI plugin, and the StorageClass provisioner together with Kubernetes:

```shell
./kk upgrade cluster -i inventory.yaml -c config.yaml --with-kubernetes v1.34.11 --all
```

Or enable individual components with `--set`:

```shell
./kk upgrade cluster -i inventory.yaml -c config.yaml --with-kubernetes v1.34.11 --set upgrade.cni=true
```

### Upgrade a Single Component Only

The following subcommands upgrade only the named component and leave the Kubernetes control plane and other components untouched:

```shell
./kk upgrade etcd -i inventory.yaml -c config.yaml
./kk upgrade cri -i inventory.yaml -c config.yaml
./kk upgrade cni -i inventory.yaml -c config.yaml
./kk upgrade storageclass -i inventory.yaml -c config.yaml
```

## Upgrade Switches

The `upgrade` section in `config.yaml` controls whether optional components are upgraded together:

```yaml
upgrade:
  etcd: false          # Whether to upgrade the external etcd cluster
  cri: false           # Whether to upgrade the container runtime (docker/containerd)
  cni: false           # Whether to upgrade the CNI plugin
  storage_class: false # Whether to upgrade the StorageClass provisioner
```

The switches above can also be overridden on the command line with `--all` or `--set upgrade.<component>=true`.

> **Note**: CoreDNS / NodeLocalDNS are upgraded together with Kubernetes and do not require a separate switch. `image_registry` and `nfs` are not supported for upgrade.

## Multi-Minor Upgrades

`kubeadm` only allows a single-minor upgrade at a time. When the target version spans multiple minor versions, KubeKey automatically steps through them one minor at a time, using the recommended patch version for each intermediate minor defined in `cluster_require.kube_upgrade_path`. You can override a specific hop with `--set cluster_require.kube_upgrade_path.v1.24=v1.24.10`.

## Parameter Reference

| Parameter | Description |
|-----------|-------------|
| `-i, --inventory` | Path to the inventory file defining node connection information |
| `-c, --config` | Path to the config file defining key cluster configuration |
| `--with-kubernetes` | Specifies the target Kubernetes version to upgrade to. Defaults to `kubernetes.kube_version` in config |
| `--all` | Upgrades all related components, including etcd, cri, cni and storage_class |
| `--set` | Overrides a value in config. Format: `--set key=val` or `--set k1=v1,k2=v2` |
| `-a, --artifact` | Path to the offline package, used in air-gapped environments |

## Important Notes

- **Rolling Upgrade**: KubeKey upgrades the control plane nodes one by one, but worker nodes are upgraded in parallel and are not automatically `cordon`-ed or `drain`-ed. If workload replicas are insufficient or concentrated on certain nodes, services may become briefly unavailable. Manually `drain` nodes beforehand if a strict rolling upgrade is required.
- **Container Runtime**: When `upgrade.cri=true`, the container runtime service is restarted. With containerd, running containers are usually preserved; with Docker, restarting dockerd may stop running containers unless `live-restore` is enabled.
- **Backup**: Back up the cluster (especially etcd data) before upgrading so you can recover if the upgrade fails.
