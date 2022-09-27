# NAME
**kk create phase**: Create your cluster in phases to a newer version with this command, which is not enough and need to add phase cmds.

# DESCRIPTION
Create your cluster in phases to a newer version with this command,  which is not enough and need to add phase cmds.

# PHASES CMDS

| Command | Description |
| - | - |
| kk create phase binary | Download the binaries on the local. |
| kk create phase os | Init the os configure. |
| kk create phase images | Down the container and pull the images before creating your cluster. |
| kk create phase etcd | Install the etcd on the master. |
| kk create phase join | Join the control-plane nodes and worker nodes in the k8s cluster. |
| kk create phase configure | Configure the k8s cluster with plugins, certs and PV. |
| kk create phase kubesphere | Install the kubesphere with the input version. |

# OPTIONS

## **--artifact, -a**
Path to a KubeKey artifact.

## **--certificates-dir**
Specifies where to store or look for all required certificates.

## **--container-manager**
Container manager: docker, crio, containerd and isula. The default is `docker`.

## **--debug**
Print detailed information. The default is `false`.

## **--download-cmd**
The user defined command to download the necessary binary files. The first param `%s` is output path, the second param `%s`, is the URL. The default is `curl -L -o %s %s`.

## **--filename, -f**
Path to a configuration file.

## **--ignore-err**
Ignore the error message, remove the host which reported error and force to continue. The default is `false`.

## **--in-cluster**
Running inside the cluster. The default is `false`.

## **--skip-pull-images**
Skip pre pull images. The default is `false`.

## **--skip-push-images**
Skip pre push images. The default is `false`.

## **--with-kubernetes**
Specify a supported version of kubernetes. It will override the version of kubernetes in the config file.

## **--with-kubesphere**
Deploy a specific version of kubesphere. It will override the kubesphere `ClusterConfiguration` in the config file with the default value.

## **--with-local-storage**
Deploy a local PV provisioner.

## **--with-packages**
Install operating system packages by artifact. The default is `false`.

## **--yes, -y**
Skip confirm check. The default is `false`.

# EXAMPLES
## Create an `all-in-one` pure Kubernetes cluster with default version
Download the binarys of creating an `all-in-one` Kubernetes cluster from a default version.
```
$ kk create phase binary
```
Init the configure os of creating an `all-in-one` Kubernetes cluster from a default version.
```
$ kk create phase os
```
Pull the images of creating an `all-in-one` Kubernetes cluster from a default version.
```
$ kk create phase images
```
Install the etcd of creating an `all-in-one` Kubernetes cluster from a default version.
```
$ kk create phase etcd
```
Init the k8s cluster of creating an `all-in-one` Kubernetes cluster from a default version.
```
$ kk create phase init 
```
Join the nodes to the k8s cluster of creating an `all-in-one` Kubernetes cluster from a default version.
```
$ kk create phase join 
```
Configure the k8s cluster of creating an `all-in-one` Kubernetes cluster from a default version.
```
$ kk create phase configure
```
## Create an `all-in-one` pure Kubernetes cluster with specified version
Download the binarys of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase binary --with-kubernetes v1.22.0
```
Init the configure os of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase os
```
Pull the images of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase images --with-kubernetes v1.22.0
```
Install the etcd of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase etcd
```
Init the k8s cluster of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase init --with-kubernetes v1.22.0
```
Join the nodes to the k8s cluster of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase join --with-kubernetes v1.22.0
```
Configure the k8s cluster of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase configure --with-kubernetes v1.22.0
```
Install the k8s cluster of creating an `all-in-one` Kubernetes cluster from a specified version.
```
$ kk create phase kubesphere --with-kubesphere v3.3.0
```
## Create a kubernetes cluster from a specified configuration file
Download the binarys of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase binary -f config-sample.yaml
```
Init the configure os of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase os
```
Pull the images of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase images -f config-sample.yaml
```
Install the etcd of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase etcd -f config-sample.yaml
```
Init the k8s cluster of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase init -f config-sample.yaml
```
Join the nodes to the k8s cluster of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase join -f config-sample.yaml
```
Configure the k8s cluster of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase configure -f config-sample.yaml
```
Install the k8s cluster of creating a kubernetes cluster from a specified configuration file.
```
$ kk create phase kubesphere -f config-sample.yaml
```
## Create a cluster from the specified configuration file and use the artifact to install operating system packages.
import a KubeKey artifact named `my-artifact.tar.gz`.
```
$ kk artifact import -a my-artifact.tar.gz --with-packages
```
Download the binarys of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase binary -f config-sample.yaml
```
Init the configure os of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase os
```
Pull the images of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase images -f config-sample.yaml
```
Install the etcd of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase etcd -f config-sample.yaml
```
Init the k8s cluster of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase init -f config-sample.yaml
```
Join the nodes to the k8s cluster of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase join -f config-sample.yaml
```
Configure the k8s cluster of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase configure -f config-sample.yaml
```
Install the k8s cluster of creating a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create phase kubesphere -f config-sample.yaml
```