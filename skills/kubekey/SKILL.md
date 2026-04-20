---
name: kubekey
description: |
  KubeKey (kk) Usage Guide. For Kubernetes cluster lifecycle management including creation, scaling and deletion.
  Use this skill whenever the user mentions kubekey, kk commands, cluster setup, cluster management, adding nodes,
  deleting nodes, creating inventory files, configuring clusters, installing Kubernetes, or deploying K8s.
  Also use this skill when the user asks about inventory.yaml, config.yaml formats, CNI configuration, container runtime setup,
  offline installation, or artifact operations. Provides detailed guidance on kk subcommand usage, config file formats, inventory structure and common workflows.
---

# KubeKey Skill

KubeKey is a lightweight Kubernetes cluster lifecycle management tool for creating, scaling and deleting clusters.

## Installation

### Binary Installation

```bash
# Linux/macOS - Latest version (stable)
curl -sfL https://get-kk.kubesphere.io | KK_ONLY=true sh -

# Download a specific pre-release version (e.g., v4.0.5)
# The get-kk script may not include pre-releases. Use the GitHub release directly:
curl -sfL https://github.com/kubesphere/kubekey/releases/download/v4.0.5/downloadKubekey.sh | KK_ONLY=true sh -

# China mirror (faster in China)
curl -sfL https://get-kk.kubesphere.io | KK_ONLY=true KKZONE=cn sh -
```

**Windows**: Use [WSL](https://docs.microsoft.com/en-us/windows/wsl/) or download from [GitHub Releases](https://github.com/kubesphere/kubekey/releases)

The script downloads `kk` binary to current directory. Move it to a directory in your `$PATH`:

```bash
# Move to /usr/local/bin
sudo mv kk /usr/local/bin/

# Verify
kk version
```

## Quick Start

### 1. Basic Cluster Creation Flow

```bash
# Step 1: Create inventory file (defines nodes)
kk create inventory -o inventory.yaml

# Step 2: Create config file (defines software: K8s version, network, storage, etc.)
# --with-kubernetes: Specify Kubernetes version (run 'kk create config --help' to see the default)
kk create config --with-kubernetes v1.33.7 -o config.yaml

# Step 3: Create cluster
# Method 1: Use config.yaml file
kk create cluster -i inventory.yaml -c config.yaml

# Method 2: Use default config for K8s version (no config.yaml file generated)
kk create cluster -i inventory.yaml --with-kubernetes v1.33.7
```

### kk create config Command

Generates default config.yaml with cluster software configuration (K8s version, network, storage, etc.).

**Common Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `--with-kubernetes` | Specify Kubernetes version | `--with-kubernetes v1.33.7` |
| `-o, --output` | Output file path | `-o config.yaml` |

**Generated Config Structure:**
- `kubernetes`: K8s version, cluster name, API Server, etc.
- `cni`: CNI plugin, Pod/Service CIDR
- `cri`: Container runtime configuration
- `storage`: Storage class configuration
- `image_registry`: Image registry configuration

### kk create inventory Command

Generates default inventory.yaml file.

**Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `-o, --output` | Output file path | `-o inventory.yaml` |

The generated inventory contains template nodes that need to be edited with actual node information.

### kk create cluster Command

Creates cluster based on inventory and config.

Two methods:
1. `kk create cluster -i inventory.yaml -c config.yaml` - Use config file
2. `kk create cluster -i inventory.yaml --with-kubernetes v1.33.7` - Use default config for K8s version (no file generated)

**Common Parameters:**

| Parameter | Description |
|-----------|-------------|
| `-i, --inventory` | Inventory file path |
| `-c, --config` | Config file path |
| `--with-kubernetes` | Specify K8s version directly (uses default config, no file generated) |
| `-f, --force` | Force execution (overwrite existing config) |
| `-v` | Log level (klog mechanism, higher number = more verbose) |

## Inventory File

For detailed format and examples, see: `references/inventory.md`

Includes:
- Complete inventory examples
- SSH/local connector configuration
- Host variables (IP, labels, taints)
- Group definitions
- HA cluster multi-node examples

## Complete Config Reference

For detailed configuration, see: `references/config.md`

Includes:
- Basic structure (cluster_require, certs, image_registry, native)
- Kubernetes configuration (apiserver, controller-manager, scheduler, kubelet, kube-proxy)
- CNI network configuration (calico, cilium, flannel, kubeovn)
- CRI container runtime configuration
- etcd configuration (internal/external)
- DNS configuration (CoreDNS, NodeLocalDNS)
- Storage class configuration (local, nfs)
- Image registry configuration (harbor, docker-registry)
- Complete configuration examples

## Command Reference

### Generate Config

```bash
# Generate node config (defines hosts, SSH connections, node roles, etc.) - defines cluster spec
kk create inventory -o inventory.yaml

# Generate software config (defines K8s version, network plugin, storage, etc.)
kk create config --with-kubernetes v1.33.7 -o config.yaml
```

### Init (optional, can be merged into create cluster)

```bash
# Initialize OS environment (install dependencies, configure network, etc.)
kk init os -i inventory.yaml -c config.yaml

# Install private image registry
kk init registry -i inventory.yaml -c config.yaml

# Node precheck
kk precheck -i inventory.yaml -c config.yaml
# Can specify check items: etcd, os, network, cri, nfs
kk precheck etcd os network cri nfs -i inventory.yaml -c config.yaml
```

### Create Cluster

```bash
# Method 1: Use config.yaml file
kk create cluster -i inventory.yaml -c config.yaml

# Method 2: Use default config for K8s version (no config.yaml file generated)
kk create cluster -i inventory.yaml --with-kubernetes v1.33.7
```

### Add Nodes

Step 1: Define new node in inventory.yaml hosts section

Step 2: Add to corresponding groups
```bash
# Method 1: Add nodes directly in inventory.yaml groups
kk add nodes -i inventory.yaml -c config.yaml

# Method 2: Add nodes via command line parameters
kk add nodes -i inventory.yaml -c config.yaml --control-plane node2,node3 --worker node4,node5
```

### Delete Nodes

```bash
kk delete nodes node1 node2 -i inventory.yaml -c config.yaml
```

### Update Certificates

```bash
kk certs renew -i inventory.yaml -c config.yaml
```

### Delete Cluster

```bash
kk delete cluster -i inventory.yaml -c config.yaml
```

### Offline Installation (Artifact)

For air-gapped environments, use artifact commands to manage offline packages:

```bash
# Export artifact package containing required packages and images
kk artifact export --with-kubernetes v1.33.7 -c config.yaml

# Push artifact images to a private registry
kk artifact images -i inventory.yaml -c config.yaml --push

# Pull artifact images to local directory
kk artifact images -i inventory.yaml -c config.yaml --pull
```

## How to Help Users

When a user asks about KubeKey or cluster operations, follow this approach:

1. **Identify the goal**: Determine if they want to create, scale, delete or configure a cluster.
2. **Check prerequisites**: Ask for node information (IPs, SSH credentials, roles) if creating or scaling a cluster.
3. **Guide step by step**: Direct them to generate templates first (`kk create inventory/config`), edit files, then execute.
4. **Point to references**: For specific config options (CNI, CRI, storage), direct them to `references/config.md`.
5. **Verify with precheck**: Recommend running `kk precheck` before cluster creation to avoid failures.

## Common Workflows

### Cluster Creation Process

**Important: Collect node information first before starting.**

#### Step 1: Collect Node Information

Prepare the following information:

| Field | Description | Example |
|-------|-------------|---------|
| Node Name | Unique identifier | node1, master1, worker1 |
| IP Address | Node network address | 192.168.1.10 |
| SSH Port | Default 22 | 22 |
| SSH User | Login username | root |
| Auth Method | Password or private key path | password: "xxx" or private_key: ~/.ssh/id_rsa |
| Node Role | control_plane/worker/etcd | Choose based on requirements |

#### Step 2: Generate Inventory Template

```bash
kk create inventory -o inventory.yaml
```

#### Step 3: Edit inventory.yaml

Fill in node information into inventory.yaml, refer to format in `references/inventory.md`.

#### Step 4: Generate Config

Generate config.yaml based on your desired configuration.

For specific configuration examples tailored to your Kubernetes version, CNI plugin, storage type, etc., refer to `references/config.md`.

```bash
# Example: Generate based on your provided configuration
kk create config --with-kubernetes v1.33.7 -o config.yaml
```

#### Step 5: Create Cluster

```bash
# Method 1: Use config.yaml file
kk create cluster -i inventory.yaml -c config.yaml

# Method 2: Use default config (no config.yaml file generated)
kk create cluster -i inventory.yaml --with-kubernetes v1.33.7
```

### Create HA Cluster

inventory.yaml needs to contain:
- 3 etcd nodes
- 3 control plane nodes
- Multiple worker nodes

## Troubleshooting

### Common Issues

1. **SSH connection failed**: Check SSH configuration in inventory
2. **Node does not meet requirements**: Run `kk precheck` for details

### Debug Mode

Use -v parameter to specify log level (klog mechanism, higher number = more verbose):

```bash
# Example: -v 5 prints debug logs
kk create cluster -i inventory.yaml -c config.yaml -v 5
```

### macOS Compatibility

When running kk on macOS, you may encounter the following issues due to the default zsh shell and macOS sudo configuration:

#### 1. `zsh: no matches found: *.sh`

**Symptom**: Playbook fails with glob pattern errors like:
```
zsh:1: no matches found: /etc/kubekey/scripts/pre_install_*.sh
```

**Root Cause**: macOS default shell is zsh, which has different glob behavior from bash.

**Fix**: Set SHELL to bash before running kk commands:
```bash
export SHELL=/bin/bash
kk create cluster -i inventory.yaml -c config.yaml
```

#### 2. `command not found: mkdir`

**Symptom**: Commands like `mkdir` or `rm` fail on localhost with:
```
zsh:4: command not found: mkdir
```

**Root Cause**: macOS sudo's `secure_path` is `/usr/local/bin` by default, which does not include `/bin` where mkdir is located.

**Fix**: Update sudo secure_path to include standard directories. On macOS, `visudo` is at `/usr/sbin/visudo`:

```bash
# Edit sudoers (use full path on macOS)
sudo /usr/sbin/visudo

# Find the line:
#   Defaults    secure_path = /usr/local/bin
# Change it to:
#   Defaults    secure_path = /usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin
```

**Note**: The sudoers change takes effect immediately and persists across reboots. Revert after use if desired.

## Get Help

```bash
# View help
kk --help

# View subcommand help
kk create cluster --help
kk add nodes --help

# View version
kk version
```
