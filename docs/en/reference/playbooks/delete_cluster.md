# Delete Cluster (delete_cluster.yaml)

`delete_cluster.yaml` is used to uninstall the entire Kubernetes cluster and its related components, supporting selective resource cleanup based on configuration.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.
   - Load the `defaults` role on all nodes.

2. **Uninstall Kubernetes and Container Runtime**
   - For nodes in the `k8s_cluster` group, execute:
     - `uninstall/kubernetes`: Uninstall Kubernetes components.
     - `uninstall/cri`: Uninstall the container runtime (triggered only when `delete.cri` is `true` and the current node does not belong to the `image_registry` group).

3. **Clean Local DNS Configuration**
   - Clean up locally-written DNS (hosts) marker segments on nodes to be uninstalled.
   - Triggered only when `delete.dns` is `true`.

4. **Uninstall etcd (external mode only)**
   - Execute `etcd/scaling_down` for nodes in the `etcd` group.
   - Triggered only when `delete.etcd` is `true` and `etcd.deployment_type` is `external`.

5. **Uninstall Image Registry**
   - Execute `uninstall/image-registry` for nodes in the `image_registry` group.
   - Triggered only when `delete.image_registry` is `true`.

## Notes

- Please confirm that important data has been backed up before deleting the cluster.
- You can control whether to clean up corresponding resources through the `delete` configuration items (e.g. `delete.cri`, `delete.etcd`, `delete.dns`, `delete.image_registry`).
