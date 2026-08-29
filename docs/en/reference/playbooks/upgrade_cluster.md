# Upgrade Cluster (upgrade_cluster.yaml)

`upgrade_cluster.yaml` performs a rolling upgrade of an existing Kubernetes cluster. By default, only Kubernetes control plane and worker node binaries are upgraded; other components must be enabled explicitly through the `upgrade` switches.

## Upgrade Switches

The `upgrade` section in `config.yaml` controls whether optional components are upgraded together:

```yaml
upgrade:
  cri: false          # Whether to upgrade the container runtime (docker/containerd)
  etcd: false         # Whether to upgrade the external etcd cluster
  dns: false          # Whether to upgrade CoreDNS / NodeLocalDNS
  image_registry: false
  cni: false          # Whether to upgrade the CNI plugin
  storage_class: false
  nfs: false
```

You can also override them on the command line with `--all` or `--set upgrade.xxx=true`.

## pre_hook

The `pre_hook` allows users to execute scripts on corresponding nodes before the upgrade.

Execution flow:
1. Copy local scripts to remote nodes at `/etc/kubekey/scripts/pre_install_{{ .inventory_hostname }}.sh`
2. Set script file permissions to `0755`
3. Iterate over all `pre_install_*.sh` files in `/etc/kubekey/scripts/` on each remote node and execute them

> **work_dir**: working directory, defaults to the current command execution directory.  
> **inventory_hostname**: the host name defined in the `inventory.yaml` file.

## precheck

The `precheck` phase verifies that the cluster meets the upgrade requirements.

**os_precheck**: OS checks, including:
- **Hostname check**: Verify the hostname format is valid
- **OS version check**: Verify the OS is in the supported distribution list
- **Architecture check**: Verify the system architecture is amd64 or arm64
- **Memory check**: Verify control plane and worker node memory meets the minimum requirement
- **Kernel version check**: Verify the kernel version meets the minimum requirement

**kubernetes_precheck**: Kubernetes-related checks, including:
- **KubeVIP check**: When using `kube-vip` for the control plane endpoint, verify the address is valid and not already in use
- **Kubernetes version check**: Verify the target Kubernetes version meets the minimum requirement
- **Upgrade path check**: In upgrade scenarios, require the installed version to be lower than the target version

**etcd_precheck**: etcd cluster checks, including:
- **Deployment type validation**: Validate `internal` or `external`
- **Version validation**: When `upgrade.etcd=true`, the target version must not be lower than the installed version and must satisfy the minimum etcd version for the target Kubernetes version
- **Disk IO performance check**: Use `fio` to test WAL fsync latency on the etcd data disk

**cri_precheck**: Container runtime checks, including:
- **Container manager check** (run on localhost): Verify the configured container manager is supported (`docker` or `containerd`)
- **Target containerd version check** (run on localhost): When upgrading CRI, verify the target containerd version meets the minimum requirement
- **Installed containerd version check** (run on `k8s_cluster` nodes): When not upgrading CRI, verify the installed containerd version on each node meets the target Kubernetes requirement
- **Docker live-restore check** (run on `k8s_cluster` nodes): When `upgrade.cri=true` and Docker is used, check whether `live-restore` is enabled in `/etc/docker/daemon.json`; print a warning if not

**network_precheck**: Network connectivity checks, including:
- Network interface, CIDR format, dual-stack support, network plugin, address space, etc.

## init

The `init` phase prepares the resources required for the upgrade on `localhost`:
- Load version-specific default variables based on the target Kubernetes version
- Generate or update certificates
- Download binaries and image manifests for Kubernetes, etcd, CRI, CNI, etc.

## upgrade

The `upgrade` phase performs the actual upgrade on each node group in order.

### native

Perform OS-level initialization for nodes in the `etcd`, `k8s_cluster`, `image_registry`, and `nfs` groups, including repository, NTP, DNS, and hostname configuration.

### etcd

Only executed when `upgrade.etcd=true` and `etcd.deployment_type=external`:
- Back up etcd data on the leader node
- Distribute the new etcd binaries
- Restart etcd nodes one by one (`serial: 1`) and wait for health

### cri

Upgrade the container runtime for nodes in the `k8s_cluster` group:
- Perform a full upgrade when `upgrade.cri=true` or the node does not yet have a CRI installed
- Back up the original configuration and binaries
- Sync new binaries, image registry certificates, and systemd service files
- Restart the container runtime service

### kubernetes

1. **pre-kubernetes**: Sync Kubernetes binaries, create directories, synchronize CA / etcd / front-proxy certificates, and apply kubeadm patches.
2. **control plane** (`serial: 1`): Upgrade control plane nodes one by one.
   - The first control plane node runs `kubeadm upgrade apply`
   - Other control plane nodes run `kubeadm upgrade node`
3. **worker**: Upgrade all worker nodes in parallel by running `kubeadm upgrade node` and restarting kubelet.

### cni / storage_class

Executed on a randomly selected control plane node:
- Upgrade the CNI plugin when `upgrade.cni=true`
- Upgrade the StorageClass provisioner when `upgrade.storage_class=true`

## post_hook

The `post_hook` phase executes after the upgrade completes:
1. If security enhancement is enabled (`.security_enhancement=true`), run the `security` role
2. Copy and execute `/etc/kubekey/scripts/post_install_*.sh` scripts

> **work_dir**: working directory, defaults to the current command execution directory.  
> **inventory_hostname**: the host name defined in the `inventory.yaml` file.

## Upgrade Risk Notes

### Worker Nodes Are Upgraded in Parallel

The worker node upgrade play in `upgrade_cluster.yaml` **does not use `serial: 1`**, so all worker nodes are upgraded **in parallel** by default. The playbook also does not automatically run `kubectl drain` or `cordon`.

This means:
- Multiple workers may become `NotReady` at the same time;
- If workload replicas are insufficient or concentrated on certain nodes, services may become briefly unavailable;
- If you need a rolling upgrade, manually `drain` nodes beforehand or control concurrency yourself.

### Upgrading CRI Together (`upgrade.cri=true`)

When `upgrade.cri=true`, the container runtime service is restarted:

- **containerd**: Restarting the containerd daemon usually does not terminate existing containers (containerd reattaches to the shims after restart), but the node will briefly become `NotReady`.
- **Docker**: Whether containers are interrupted depends on whether `live-restore` is enabled:
  - With `live-restore` enabled, restarting dockerd does not interrupt running containers.
  - Without `live-restore`, restarting dockerd may stop running containers.

Therefore, when using Docker with `upgrade.cri=true`, the precheck will inspect `/etc/docker/daemon.json` on each `k8s_cluster` node and print a warning if `live-restore` is not enabled. It is recommended to enable `live-restore` before the upgrade, or ensure that critical workload replicas are distributed across multiple worker nodes.
