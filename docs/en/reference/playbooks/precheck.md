# Cluster Precheck (precheck.yaml)

`precheck.yaml` is used to perform environment and condition checks on nodes before cluster installation or expansion, ensuring that requirements for the operating system, Kubernetes, network, etcd, container runtime, and storage are met.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes (tag: `always`).

2. **Load Default Variables and Checks**
   - Load the `defaults` role on all nodes (tag: `always`).
   - Execute the `precheck` role to complete the following sub-item checks:
     - **OS check**: hostname compliance, supported distributions, system architecture, memory, kernel version.
     - **Kubernetes check**: IP address configuration, KubeVIP validity, Kubernetes version compatibility, installed Kubernetes version match.
     - **Network check**: network interfaces, CIDR format, dual-stack support, network plugin validity, available address space.
     - **etcd check**: deployment type validation, disk IO performance, installed etcd detection.
     - **Container runtime check**: container manager support, containerd minimum version.
     - **NFS check**: NFS server node uniqueness.
     - **Image registry check**: whether required software (Docker, Docker Compose) is configured.

## Filtering by category

Precheck supports filtering checks by category (tag), so you can run or skip specific categories:

- **Whitelist (run only the listed categories)**: `kk precheck --tags cri,os` runs only the container-runtime (cri) and OS (os) checks; all other categories are skipped. Positional arguments work too: `kk precheck cri os`.
- **Blacklist (skip the listed categories)**: `kk precheck --skip-tags cri` runs every category except the container runtime. `--skip-tags` is the recommended way to "turn off one category while keeping the rest".

When no filter is given, `kk precheck` runs all categories. Roles tagged `always` (e.g. `native/root`, `defaults`) always execute regardless of filtering.

Available category tags: `artifact`, `cri`, `cni`, `storageclass`, `os`, `network`, `storage`, `kubernetes`, `etcd`, `nfs`.

## Notes

- If any check fails, the playbook will stop execution and return the corresponding error message.
- It is recommended to always run this playbook before formal installation or adding nodes to avoid mid-process failures.
