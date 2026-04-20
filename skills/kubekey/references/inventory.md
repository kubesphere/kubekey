# Inventory File Format

Inventory defines cluster node information.

## Complete Example

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
      connector:
        type: ssh
        host: node1
        port: 22
        user: root
        password: "123456"
        # Or use private key: private_key: ~/.ssh/id_rsa
      internal_ipv4: 1.1.1.1

  groups:
    # All Kubernetes nodes
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker

    # Control plane nodes
    kube_control_plane:
      hosts:
        - node1

    # Worker nodes
    kube_worker:
      hosts:
        - node2

    # etcd nodes (when etcd_deployment_type is external)
    etcd:
      hosts:
        - node1

    # Image registry node (optional)
    # image_registry:
    #   hosts:
    #     - node1

    # NFS node (optional)
    # nfs:
    #   hosts:
    #     - node1
```

## Connector Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
spec:
  hosts:
    node1:
      connector:
        # Connection type: ssh, local
        type: ssh
        # Host address
        host: node1
        # Port
        port: 22
        # Username
        user: root
        # Password authentication
        password: "123456"
        # Or use private key authentication
        private_key: ~/.ssh/id_rsa
        # Or provide key content directly
        # private_key_content: ""
```

## Host Variables

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
spec:
  hosts:
    node1:
      # SSH connection configuration
      connector:
        type: ssh
        host: node1
        port: 22
        user: root
        password: "xxx"
      # Node internal IP (for Pod communication)
      internal_ipv4: 192.168.1.10
      # Node external IP (optional)
      # external_ipv4: 1.1.1.1
      # Labels (optional)
      # labels:
      #   node-role.kubernetes.io/control-plane: ""
      # Taints (optional)
      # taints:
      #   - key: node-role.kubernetes.io/control-plane
      #     effect: NoSchedule
```

## Group Reference

| Group Name | Description |
|------------|-------------|
| `k8s_cluster` | All Kubernetes nodes, includes kube_control_plane and kube_worker |
| `kube_control_plane` | Control plane nodes (master) |
| `kube_worker` | Worker nodes |
| `etcd` | etcd nodes (required for external deployment) |
| `image_registry` | Image registry node |
| `nfs` | NFS storage node |

## Multi-Node Example

### HA Cluster

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: ha-cluster
spec:
  hosts:
    master1:
      connector:
        type: ssh
        host: 192.168.1.10
        port: 22
        user: root
        password: "xxx"
      internal_ipv4: 192.168.1.10
    
    master2:
      connector:
        type: ssh
        host: 192.168.1.11
        port: 22
        user: root
        password: "xxx"
      internal_ipv4: 192.168.1.11
    
    master3:
      connector:
        type: ssh
        host: 192.168.1.12
        port: 22
        user: root
        password: "xxx"
      internal_ipv4: 192.168.1.12
    
    worker1:
      connector:
        type: ssh
        host: 192.168.1.20
        port: 22
        user: root
        password: "xxx"
      internal_ipv4: 192.168.1.20
    
    worker2:
      connector:
        type: ssh
        host: 192.168.1.21
        port: 22
        user: root
        password: "xxx"
      internal_ipv4: 192.168.1.21

  groups:
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    
    kube_control_plane:
      hosts:
        - master1
        - master2
        - master3
    
    kube_worker:
      hosts:
        - worker1
        - worker2
    
    etcd:
      hosts:
        - master1
        - master2
        - master3
```
