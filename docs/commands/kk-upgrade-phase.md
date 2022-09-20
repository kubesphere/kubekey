# NAME
**kk upgade phase**: Upgrade your cluster in phases to a newer version with this command, which is not enough and need to add phase cmds.

# DESCRIPTION
Upgrade your cluster in phases to a newer version with this command,  which is not enough and need to add phase cmds.

# PHASES CMDS
## **binary**: Download the binary and synchronize kubernetes binaries

## **images**: Pull the images before create your cluster

## **nodes**: Upgrade cluster on master nodes and worker nodes to the version you input

## **kubesphere**: Upgrade your kubesphere to a newer version with this command

# OPTIONS

## **--artifact, -a**
Path to a KubeKey artifact.

## **--debug**
Print detailed information. The default is `false`.

## **--download-cmd**
The user defined command to download the necessary binary files. The first param `%s` is output path, the second param `%s`, is the URL. The default is `curl -L -o %s %s`.

## **--filename, -f**
Path to a configuration file.

## **--ignore-err**
Ignore the error message, remove the host which reported error and force to continue. The default is `false`.

## **--skip-pull-images**
Skip pre pull images. The default is `false`.

## **--with-kubernetes**
Specify a supported version of kubernetes. It will override the version of kubernetes in the config file.

## **--with-kubesphere**
Deploy a specific version of kubesphere. It will override the kubesphere `ClusterConfiguration` in the config file with the default value.

## **--yes, -y**
Skip confirm check. The default is `false`.

# EXAMPLES
## Upgrade an `all-in-one` Kubernetes cluster from a specified version
Upgrade the binarys of an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk upgrade phase binary --with-kubernetes v1.23.8
```
Upgrade the images of an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk upgrade phase images --with-kubernetes v1.23.8
```
Upgrade the cluster nodes of an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk upgrade phase nodes --with-kubernetes v1.23.8
```
Upgrade the kubesphere of an `all-in-one` Kubernetes cluster from a specified version if you need.
```
$ kk upgrade phase kubesphere --with-kubesphere v3.3.0
```
## Upgrade an kubernetes cluster from a specified configuration file
Upgrade the binarys from a specified configuration file.
```
$ kk upgrade phase binary -f config-example.yaml
```
Upgrade the images from a specified configuration file.
```
$ kk upgrade phase images -f config-example.yaml
```
Upgrade the cluster nodes from a specified configuration file.
```
$ kk upgrade phase nodes -f config-example.yaml
```
Upgrade the kubesphere from a specified configuration file if you need.
```
$ kk upgrade phase kubesphere -f config-example.yaml
```
## Upgrade a cluster using a KubeKey artifact (in an offline enviroment).
import a KubeKey artifact named `my-artifact.tar.gz`.
```
$ kk artifact import -a my-artifact.tar.gz
```
Upgrade the binarys from a specified configuration file.
```
$ kk upgrade phase binary -f config-example.yaml
```
Upgrade the images from a specified configuration file.
```
$ kk upgrade phase images -f config-example.yaml
```
Upgrade the cluster nodes from a specified configuration file.
```
$ kk upgrade phase nodes -f config-example.yaml
```
Upgrade the kubesphere from a specified configuration file if you need.
```
$ kk upgrade phase kubesphere -f config-example.yaml
```