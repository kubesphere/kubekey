# KubeKey Manifest And Artifact
KubeKey v2.0.0 (hereinafter kk) adds the concepts of `manifest` and `artifact` to provide a solution for users to deploy Kubernetes clusters offline enviroment. In the past, users had to prepare deployment tools, images' `tar` files, and other related binaries, and each user has a different version of Kubernetes to deploy and different images to deploy. Now with kk, you only need to use the `manifest` file to define what you need for the cluster environment to be deployed offline, and then use that `manifest` to export the `artifact` file to complete the preparation. Then offline installation requires only kk and `artifact` for quick and easy deployment of image registry (docker-registry or harbor) and Kubernetes clusters in your environment.

## What is the KubeKey Manifest?
A `manifest` is a text file that describes information about the current Kubernetes cluster and defines what needs to be included in the `artifact` . There are currently two ways to generate this file：
* Manually creating and writing the file from a template.
* Generate the file from an existing cluster using the kk command.

The first way requires more information about the different fields in this configuration file, see [manifest-example.yaml](./manifest-example.md).

The following is only for the second way of generating files using the kk command.

### Usage
> Note:
> This way requires preparing a cluster environment that has been installed a Kubernetes cluster and providing the `./kube/config` file of the cluster for kk.

Commands:
```
./kk create manifest
```
By default, kk will use the `$HOME/.kube/config` file, or you can specify a `config` file:
```
./kk create manifest --kubeconfig config
```
After execution, the `manifest-sample.yaml` file will be generated in the current directory. The contents of the `manifest-sample.yaml` file can then be modified to export the desired `artifact` file later.

### Principle
kk connects to the corresponding Kubernetes cluster via the `kubeconfig` file and then checks out the following information in the cluster environment:
* Node architecture
* Node operating system
* The images on the node
* Kubernetes version
* CRI information

After that, the description of the current cluster will be written to the `manifest` file. Besides, other undetectable files (e.g. ETCD cluster information, image regsitry, etc.) will be written to the `manifest` file according to the default values recommended by kk.

## What is the KubeKey Artifact?
The `artifact` is a `.tar.gz` package containing the images' `tar` file and other related binaries, exported from the specified `manifest` file. An `artifact` can be specified in the kk `init registry`, `create cluster`, `add node` and `upgrade cluster` commands. kk will automatically unpack the `artifact` and will use the unpacked file directly when executing the command.

### Usage
#### Export Artifact
> Note:
> 1. The export command will download the corresponding binaries from the Internet, so please make sure the network connection is success.
> 2. Make sure kk's work node has containerd or a minimum version of 18.09 docker installed.
> 3. kk will parse the image's name in the image list, if the mirror in the image's name needs authentication information, you can configure it in the `.registry.auths` field in the `manifest` file.
> 4. If the `artifact` file to be exported contains OS dependency files (e.g. conntarck, chrony, etc.), you can configure the corresponding ISO dependency download URL address in the `.repostiory.iso.url` in the `operationSystems` field.

* Export
```
./kk artifact export -m manifest-sample.yaml
```
After execution, the `kubekey-artifact.tar.gz` file will be generated in the current directory.

#### Use Artifact
> Note:
> 1. In an offline environment, you need to use kk to generate the `config-sample.yaml` file and configure the corresponding information before using the `artifact`.
> 2. In an offline environment, the `artifact` image will be pushed to the private image registry by default when you use the `create cluster` and `upgrade cluster` commands. If the private image registry needs authentication information, you can configure it in the `.spec.registry.auths` field in the `config-sample.yaml` file.

* Initialize the image registry, the related configuration can be found in [container image registry](./registry.md).
 ```
./kk init registry -f config-sample.yaml -a kubekey-artifact.tar.gz
```
* Create the cluster.
> Note: In an offline environment, you need to configure private image registry information for cluster image management, please refer to [config-sample.yaml](./config-example.md) and [container image registry](./registry.md).

```
./kk create cluster -f config-sample.yaml -a kubekey-artifact.tar.gz
```

* Create the cluster and install the OS dependencies (requires the `artifact` to contain the OS dependency files for the nodes in the target cluster).
```
./kk create cluster -f config-sample.yaml -a kubekey-artifact.tar.gz --with-packages
```
* Add nodes.
```
./kk add nodes -f config-sample.yaml -a kubekey-artifact.tar.gz
```
* Upgrade the cluster.
```
./kk upgrade -f config-sample.yaml -a kubekey-artifact.tar.gz
```