# NAME
**kk artifact export**: Export a KubeKey offline installation package.

# DESCRIPTION
**kk** will base on the specify manifest file to pull all images, download the specified binaries and Linux repository iso file, then archive them as a KubeKey offline installation package. The export command will download the corresponding binaries from the Internet, so please make sure the network connection is success. And kk will pull the images by `containerd` if the length of manifest's images list is not 0, so make sure kk's work node has `containerd` or a minimum version of 18.09 `docker` installed.

# OPTIONS

## **--manifest, -m**
Path to a manifest file. This option is required.

## **--output, -o**
Path to a output path The default is `kubekey-artifact.tar.gz`.

## **--cri-socket**
Path to the CRI socket to connect. If empty KubeKey will try to auto-detect this value.

## **--download-cmd**
The user defined command to download the necessary binary files. The first param `%s` is output path, the second param `%s`, is the URL. The default is `curl -L -o %s %s`.

## **--debug**
Print detailed information. The default is `false`.

# EXAMPLES
Export a KubeKey artifact named `my-artifact.tar.gz`.
```
$ kk artifact export -m manifest-sample.yaml -o my-artifact.tar.gz
```
Export a KubeKey artifact with a specify CRI socket.
```
$ kk artifact export -m manifest-sample.yaml --cri-socket /run/containerd/containerd.sock 
```