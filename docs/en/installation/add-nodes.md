# Add Cluster Nodes

This section describes how to use KubeKey to add new nodes to an existing Kubernetes cluster, including control plane nodes, worker nodes, and etcd nodes.

## Prerequisites

- An existing Kubernetes cluster deployed by KubeKey.
- New nodes prepared to join the cluster and meeting the system requirements (see [Install Kubernetes](README.md)).

> **Note**: Web Installer does not currently support adding cluster nodes. Please use the command line instead.

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

## Method 1: Add Nodes via inventory.yaml Groups

This method requires that the nodes to be added are already defined in `inventory.yaml` with their connection information and assigned to the corresponding groups (e.g., `kube_control_plane`, `kube_worker`, `etcd`).

1. Ensure `inventory.yaml` contains the new nodes' connection information and group assignments.

   Example:

   ```yaml
   spec:
     hosts:
       node1:
         connector:
           type: ssh
           host: 192.168.1.101
           port: 22
           user: root
           password: 123456
     groups:
       kube_control_plane:
         hosts:
           - localhost
           - node1
       kube_worker:
         hosts:
           - localhost
           - node1
       etcd:
         hosts:
           - localhost
   ```

2. Execute the following command to add the nodes:

   ```shell
   ./kk add nodes -i inventory.yaml -c config.yaml
   ```

   KubeKey will automatically detect nodes defined in `inventory.yaml` that are not yet part of the cluster and install them according to their assigned group roles.

## Method 2: Add Nodes via Command-Line Flags

This method only requires the nodes' connection information to be defined in `inventory.yaml`, without pre-assigning them to groups. You specify the node roles via command-line flags and can use `--override` to automatically update `inventory.yaml`.

1. Ensure `inventory.yaml` defines the connection information for the nodes to be added.

   Example:

   ```yaml
   spec:
     hosts:
       node1:
         connector:
           type: ssh
           host: 192.168.1.101
           port: 22
           user: root
           password: 123456
       node2:
         connector:
           type: ssh
           host: 192.168.1.102
           port: 22
           user: root
           password: 123456
   ```

2. Execute the following command to add nodes and specify their roles:

   ```shell
   ./kk add nodes --control-plane node1 --worker node2 -i inventory.yaml -c config.yaml --override
   ```

   > **PS**: In an offline environment, you can add the `--set download.fetch=false` parameter to prevent downloading resources from the internet.

   - `--control-plane`: Specifies the hostnames to be added as control plane nodes. Multiple nodes are separated by commas.
   - `--worker`: Specifies the hostnames to be added as worker nodes. Multiple nodes are separated by commas.
   - `--etcd`: Specifies the hostnames to be added as etcd nodes. Multiple nodes are separated by commas.
   - `--override`: After successful execution, automatically adds the nodes to the corresponding groups and updates `inventory.yaml`.

## Parameter Reference

| Parameter | Description |
|-----------|-------------|
| `-i, --inventory` | Path to the inventory file defining node connection information |
| `-c, --config` | Path to the config file defining key cluster configuration |
| `--with-kubernetes` | Specifies the Kubernetes version. Defaults to the current cluster version |
| `--control-plane` | Specifies the list of nodes to be added as control plane nodes |
| `--worker` | Specifies the list of nodes to be added as worker nodes |
| `--etcd` | Specifies the list of nodes to be added as etcd nodes |
| `--override` | Overwrites and updates the inventory.yaml file after successful execution |
| `-a, --artifact` | Path to the offline package, used in air-gapped environments |
