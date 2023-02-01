# NAME
**kk delete cluster**: Delete a cluster.

# DESCRIPTION
Delete a cluster. This command will use the `kubeadm reset` to reset all the nodes. Then, reset network policy, stop `etcd`, remove cluster directory, uninstall Kubernetes certs-auto-renew script and remove internal Loadbalancer module. And [network configurations](../network-configurations.md) on each node will be cleaned up.

# OPTIONS

## **--debug**
Print detailed information. The default is `false`.

## **--filename, -f**
Path to a configuration file.

## **--all, -A**
Delete all CRI(docker/containerd) related files and directories.

# EXAMPLES
Delete an `all-in-one` cluster.
```
$ kk delete cluster
```
Delete a cluster from a specified configuration file.
```
$ kk delete cluster -f config-example.yaml
```
Delete a cluster included CRI related files and directories from a specified configuration file.
```
$ kk delete cluster -f config-example.yaml --all
$ kk delete cluster -f config-example.yaml -A
```

