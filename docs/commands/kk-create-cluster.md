# NAME
**kk create cluster**: Add nodes to the cluster according to the new nodes information from the specified configuration file.

# DESCRIPTION
Add nodes to the cluster according to the new nodes information from the specified configuration file. You need to add new node's information to the cluster config file first, then apply the changes.

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
Create an `all-in-one` pure Kubernetes cluster with default version.
```
$ kk create cluster
```
Create a Kubernetes and KubeSphere cluster with a specified version.
```
$ kk create cluster --with-kubernetes v1.21.5 --with-kubesphere v3.2.1
```
Create a cluster using the configuration file.
```
$ kk create cluster -f config-sample.yaml
```
Create a cluster from the specified configuration file and use the artifact to install operating system packages.
```
$ kk create cluster -f config-sample.yaml -a kubekey-artifact.tar.gz --with-packages
```