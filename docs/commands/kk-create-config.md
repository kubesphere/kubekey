# NAME
**kk create config**: Create cluster configuration file

# DESCRIPTION
Create cluster configuration file. More information about the configuration file can be found in the [config-example.yaml](../config-example.md).

# OPTIONS

## **--debug**
Print detailed information. The default is `false`.

## **--filename, -f**
Specify the configuration file output path. The default is `./config-example.yaml`.

## **--from-cluster**
Create a configuration based on existing cluster. Usually used in conjunction with the ``--kubeconfig`` option.

## **--kubeconfig**
Specify a kubeconfig file.

## **--name**
Specify a name of cluster object. The default is `sample`.

## **--with-kubernetes**
Specify a supported version of kubernetes.

## **--with-kubesphere**
Deploy a specific version of kubesphere. It will generate the kubesphere `ClusterConfiguration` in the config file with the default value.

# EXAMPLES
Create an example configuration file with default Kubernetes version.
```
$ kk create config
```
Create a Kubernetes and KubeSphere example configuration file with a specified version.
```
$ kk create config --with-kubernetes v1.21.5 --with-kubesphere v3.2.1
```
Create a example configuration file based on existing cluster
```
$ kk create config --from-cluster --kubeconfig ~/.kube/config
```
