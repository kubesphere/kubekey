# NAME
**kk artifact images push**: Push images to a registry from a KubeKey artifact.

# DESCRIPTION
Push images to a registry from a KubeKey artifact.

# OPTIONS

## **--filename, -f**
Path to a configuration file.

## **--images-dir**
Path to a KubeKey artifact images directory (e.g. ./kubekey/images).

## **--artifact, -a**
Path to a KubeKey artifact.

## **--debug**
Print detailed information. The default is `false`.

# EXAMPLES
Push the image to the private image registry.
```
$ kk artifact images push -f config-sample.yaml -a kubekey-artifact.tar.gz
```
Push the image to the private image registry from a specify directory.
```
$ kk artifact images push -f config-sample.yaml --images-dir ./kubekey/images
```