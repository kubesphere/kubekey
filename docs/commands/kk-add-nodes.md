# NAME
**kk add nodes**: Add nodes to the cluster according to the new nodes information from the specified configuration file.

# DESCRIPTION
Add nodes to the cluster according to the new nodes information from the specified configuration file. You need to add new node's information to the cluster config file first, then apply the changes.

# OPTIONS

## **--filename, -f**
Path to a configuration file.

## **--skip-pull-images**
Skip pre pull images. The default is `false`.

## **--container-manager**
Container manager: docker, crio, containerd and isula. The default is `docker`.

## **--download-cmd**
The user defined command to download the necessary binary files. The first param `%s` is output path, the second param `%s`, is the URL. The default is `curl -L -o %s %s`.

## **--artifact, -a**
Path to a KubeKey artifact.

## **--with-packages**
Install operating system packages by artifact. The default is `false`.

## **--in-cluster**
Running inside the cluster. The default is `false`.

## **--debug**
Print detailed information. The default is `false`.

## **--yes, -y**
Skip confirm check. The default is `false`.

## **--ignore-err**
Ignore the error message, remove the host which reported error and force to continue. The default is `false`.

# EXAMPLES
Add nodes from the specified configuration file.
```
$ kk add nodes -f config-sample.yaml
```
Add nodes from the specified configuration file and use the artifact to install operating system packages.
```
$ kk add nodes -f config-sample.yaml -a kubekey-artifact.tar.gz --with-packages
```