# Initialize OS (init_os.yaml)

`init_os.yaml` is used to initialize the operating system environment of cluster nodes, including system configuration, package installation, and certificate preparation. It is suitable for node preparation before cluster installation.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.

2. **Load Default Variables**
   - Load the `defaults` role on all nodes.

3. **Certificate and Resource Preparation**
   - Execute on `localhost`:
     - `certs/init`: Initialize and generate certificates required by the cluster.
     - `download`: Download software packages such as Kubernetes and container runtime.

4. **Node System Initialization**
   - Execute the `native` role for nodes in the `etcd`, `k8s_cluster`, `image_registry`, and `nfs` groups to install necessary system packages and complete basic environment configuration.

## Notes

- This playbook itself does not actually deploy Kubernetes or CRI; its main purpose is to complete OS-level node preparation.
- It can be used to batch-initialize node environments before formally executing `create_cluster.yaml`.
