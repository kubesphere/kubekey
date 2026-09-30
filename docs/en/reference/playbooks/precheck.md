# Cluster Precheck (precheck.yaml)

`precheck.yaml` is used to perform environment and condition checks on nodes before cluster installation or expansion, ensuring that requirements for the operating system, Kubernetes, network, etcd, container runtime, and storage are met.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes (tag: `always`).

2. **Load Default Variables and Checks**
   - Load the `defaults` role on all nodes (tag: `always`).
   - Execute the `precheck` role to complete the following sub-item checks:
     - **OS check**: hostname compliance, supported distributions, system architecture, memory, kernel version, and whether the CPU provides the x86-64-v2 microarchitecture level that Calico requires.
     - **Kubernetes check**: IP address configuration, KubeVIP validity, Kubernetes version compatibility, installed Kubernetes version match.
     - **Network check**: network interfaces, CIDR format, dual-stack support, network plugin validity, available address space.
     - **etcd check**: deployment type validation, disk IO performance, installed etcd detection.
     - **Container runtime check**: container manager support, containerd minimum version, containerd v2 compatibility with the node glibc and kernel (nodes with glibc older than 2.35 must set `cri.containerd.static_binary=true`).
     - **NFS check**: NFS server node uniqueness.
     - **Image registry check**: whether required software (Docker, Docker Compose) is configured.

## Notes

- If any check fails, the playbook will stop execution and return the corresponding error message.
- It is recommended to always run this playbook before formal installation or adding nodes to avoid mid-process failures.

## Known Issues

- **runc masked-path mount fails on Ubuntu 20.04 (5.4 kernel)**: runc 1.3.6 and later mount masked `/proc` paths with a tmpfs carrying `nr_blocks=1,nr_inodes=1` (opencontainers/runc#5275). The official Ubuntu 20.04 5.4 kernel carries a Canonical-private `mm/shmem.c` patch that enforces a lower bound of 2 on `nr_inodes` and rejects the mount, so every pod fails to start with `can't mask dir "/proc/acpi": ... invalid argument` (opencontainers/runc#5348). The patch is absent from mainline/stable kernels and from Ubuntu 22.04+ (5.15+), so Debian, RHEL/CentOS, and Ubuntu 22.04+ nodes are unaffected. runc v1.5.1+ ships a fallback using `nr_inodes=2` that resolves this. KubeKey's default runc (v1.3.6 / v1.4.3 in containerd mode) triggers this on Ubuntu 20.04 + 5.4 kernel nodes; the precheck does not guard against it yet, so upgrade runc to v1.5.1+ or the distribution on affected nodes.
