# Add Nodes (add_nodes.yaml)

`add_nodes.yaml` is used to add new nodes to an existing Kubernetes cluster, supporting the addition of etcd nodes, worker nodes, or control plane nodes.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.

2. **Pre Install Hook**
   - Import and execute pre-installation scripts from `hook/pre_install.yaml`.
   - Execution flow:
     1. If `pre_install.sh` exists in `scripts_dir`, copy it to all remote nodes at `/etc/kubekey/scripts/pre_install.sh`
     2. For each group listed in `pre_install_groups` (in order), if the current host belongs to that group and `{{ .scripts_dir }}/{{ group }}/pre_install.sh` exists, copy it to `/etc/kubekey/scripts/pre_install_group_{{ group }}.sh`
     3. Copy the host-specific script `pre_install_{{ .inventory_hostname }}.sh` to `/etc/kubekey/scripts/pre_install_{{ .inventory_hostname }}.sh` if it exists
     4. Set script file permissions to 0755
     5. Execute `/etc/kubekey/scripts/pre_install.sh` first if it exists
     6. Execute the copied group scripts in the order defined by `pre_install_groups`
     7. Execute `/etc/kubekey/scripts/pre_install_{{ .inventory_hostname }}.sh` last if it exists

3. **Load Defaults and Precheck**
   - Load default configurations (`defaults`) on all nodes.
   - Execute the `precheck` role to verify that new nodes meet the conditions for joining the cluster.

4. **Resource Preparation**
   - Execute `certs/init` on `localhost` to generate or update certificates.
   - Execute `download` on `localhost` to download required software packages and images.

5. **Node Initialization**
   - Execute the `native` role for all nodes in the `etcd`, `k8s_cluster`, `image_registry`, and `nfs` groups to install base packages and configure the system environment.

6. **etcd Expansion (external mode only)**
   - For nodes in the `etcd` group, execute in sequence:
     - `etcd/prepare`
     - `etcd/backup`
     - `etcd/scaling_up/learner`
     - `etcd/install`
     - `etcd/scaling_up/promote`
     - `etcd/postprocess`
   - The above steps are triggered only when `etcd.deployment_type` is `external` and the node is in the `need_installed_etcd` list.

7. **Sync etcd Config**
   - Execute `kubernetes/sync-etcd-config` on `kube_control_plane` nodes to sync etcd configuration to the control plane.

8. **Container Runtime and Kubernetes Installation**
   - For nodes in the `k8s_cluster` group, execute:
     - `cri`: Install the container runtime.
     - `kubernetes/pre-kubernetes`: Install pre-requisites.
     - `kubernetes/init-kubernetes`: Initialize Kubernetes.
     - `kubernetes/join-kubernetes`: Join the new node to the cluster (triggered only when the node has not yet loaded Kubernetes services).
     - `kubernetes/certs`: Distribute or renew certificates (triggered only on control plane nodes when certificate renewal is enabled).
   - The above roles are filtered using the `add_nodes` list, affecting only nodes that need to be added.

9. **Post Install Hook**
   - Import and execute post-installation scripts from `hook/post_install.yaml`.
   - Execution flow:
     1. If `post_install.sh` exists in `scripts_dir`, copy it to all remote nodes at `/etc/kubekey/scripts/post_install.sh`
     2. For each group listed in `post_install_groups` (in order), if the current host belongs to that group and `{{ .scripts_dir }}/{{ group }}/post_install.sh` exists, copy it to `/etc/kubekey/scripts/post_install_group_{{ group }}.sh`
     3. Copy the host-specific script `post_install_{{ .inventory_hostname }}.sh` to `/etc/kubekey/scripts/post_install_{{ .inventory_hostname }}.sh` if it exists
     4. Set script file permissions to 0755
     5. Execute `/etc/kubekey/scripts/post_install.sh` first if it exists
     6. Execute the copied group scripts in the order defined by `post_install_groups`
     7. Execute `/etc/kubekey/scripts/post_install_{{ .inventory_hostname }}.sh` last if it exists
