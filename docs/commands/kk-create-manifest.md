# NAME
**kk create manifest**: Create an offline installation package configuration file.

# DESCRIPTION
Create an offline installation package configuration file. This command requires preparing a cluster environment that has been installed a Kubernetes cluster and providing the `kube config` file of the cluster for **kk**. More information about the KubeKey manifest file can be found in the [KubeKey Manifest and Artifact](../manifest_and_artifact.md) and [manifest-example.yaml](../manifest-example.md).

# OPTIONS

## **--debug**
Print detailed information. The default is `false`.

## **--filename, -f**
Specify the manifest file output path. The default is `./manifest-example.yaml`.

## **--kubeconfig**
Specify a kubeconfig file. The default is `$HOME/.kube/config`.

## **--name**
Specify a name of manifest object. The default is `sample`.

# EXAMPLES
Create an example manifest file based on the default `kube config ($HOME/.kube/config)` path.
```
$ kk create manifest
```
Create a manifest file with a specified `kube config` path.
```
$ kk create manifest --kubeconfig /root/.kube/config
```

