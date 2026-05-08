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

## Notes

- If any check fails, the playbook will stop execution and return the corresponding error message.
- It is recommended to always run this playbook before formal installation or adding nodes to avoid mid-process failures.
