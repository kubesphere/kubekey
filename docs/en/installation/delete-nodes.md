# Delete Cluster Nodes

This section describes how to use KubeKey to safely remove specified nodes from a Kubernetes cluster, including evicting and deleting nodes from Kubernetes, cleaning up etcd members, uninstalling the container runtime, and cleaning up DNS configuration.

## Prerequisites

- An existing Kubernetes cluster deployed by KubeKey.
- Ensure that workloads on the nodes to be deleted have been migrated or backed up.
- When deleting control plane nodes, ensure that at least one available control plane node remains in the cluster.

> **Note**: Web Installer does not currently support deleting cluster nodes. Please use the command line instead.

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

## Delete Nodes

Execute the following command to delete the specified nodes:

```shell
./kk delete nodes node1 node2 -i inventory.yaml -c config.yaml
```

Where `node1 node2` are the names of the nodes to be deleted, which must match the host names defined in `inventory.yaml`.

KubeKey will perform the following operations in sequence:

1. If the node belongs to the etcd cluster and `--all` or related deletion options are configured, the etcd member will be removed first.
2. For Kubernetes nodes, execute `cordon` to prevent new Pod scheduling and `drain` to evict existing workloads.
3. If Calico is used, execute `calicoctl delete node` to clean up network resources.
4. Delete the node from the Kubernetes cluster.
5. Uninstall Kubernetes components and the container runtime (according to configuration).
6. Clean up local DNS hosts configuration (according to configuration).

## Parameter Reference

| Parameter | Description |
|-----------|-------------|
| `-i, --inventory` | Path to the inventory file defining node connection information |
| `-c, --config` | Path to the config file defining key cluster configuration |
| `--with-kubernetes` | Specifies the Kubernetes version |
| `--all` | Deletes all cluster components, including cri, etcd, dns, and image_registry |
| `--with-data` | Also deletes data directories (e.g., harbor data, registry data). Use with caution. |
| `--override` | Removes the deleted nodes from inventory.yaml after successful execution |
| `-a, --artifact` | Path to the offline package, used in air-gapped environments |

## Important Notes

- **Workload Migration**: Before deleting nodes, ensure that workloads on the nodes have been migrated or backed up to avoid data loss.
- **Control Plane Nodes**: Deleting control plane nodes triggers additional safety checks to prevent accidental deletion from causing cluster failure. Be sure to ensure that at least one available control plane node remains after deletion.
- **etcd Node Scaling Down**: When using external etcd mode, ensure that the remaining etcd member count is odd after deleting etcd nodes to maintain cluster high availability.
