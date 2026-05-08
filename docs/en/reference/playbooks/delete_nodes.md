# Delete Nodes (delete_nodes.yaml)

`delete_nodes.yaml` is used to safely remove specified nodes from the cluster, including eviction and deletion of nodes from Kubernetes, etcd decommissioning, CRI uninstallation, and DNS cleanup.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.
   - Load the `defaults` role on all nodes.

2. **etcd Node Scaling Down (external mode only)**
   - For nodes in the `etcd` group that need to be uninstalled, execute in sequence:
     - `etcd/prepare`
     - `etcd/scaling_down`
     - `etcd/postprocess`
   - Triggered only when `delete.etcd` is `true` and the node is in the `need_uninstall_etcd` list.

3. **Ensure Control Plane Nodes Are Available**
   - Execute pre-checks on `kube_control_plane` nodes to ensure at least one control plane node remains after deletion.
   - Also execute `kubernetes/sync-etcd-config` to sync etcd configuration to the remaining control plane nodes.

4. **Remove Node from Kubernetes Cluster**
   - For nodes in the `k8s_cluster` group to be deleted, execute pre-tasks:
     - `kubectl cordon`: Prevent new Pods from being scheduled to the node.
     - `kubectl drain`: Evict workloads from the node.
     - If using Calico, execute `calicoctl delete node`.
     - `kubectl delete node`: Delete the node from the cluster.
   - Then execute:
     - `uninstall/kubernetes`: Uninstall Kubernetes components.
     - `uninstall/cri`: Uninstall the container runtime (if `delete.cri` is configured and the node does not belong to the `image_registry` group).

5. **Clean Local DNS Configuration**
   - Clean up locally-written hosts marker segments by KubeKey.
   - Triggered only when `delete.dns` is `true` and the node is in the `delete_nodes` list.

6. **Uninstall etcd and Image Registry**
   - Execute the `etcd` role on corresponding nodes to uninstall etcd (if `delete.etcd` is enabled).
   - Execute `uninstall/image-registry` on `image_registry` nodes (if `delete.image_registry` is enabled).

## Notes

- Before deleting a node, ensure that workloads on the node have been migrated or backed up.
- Deleting control plane nodes will trigger additional safety checks to prevent accidental deletion that could cause cluster loss of control.
