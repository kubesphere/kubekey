# NAME
**kk upgade**: Upgrade your cluster smoothly to a newer version with this command.

# DESCRIPTION
Upgrade your cluster smoothly to a newer version with this command.

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
Upgrade an `all-in-one`  Kubernetes cluster from a specified version.
```
$ kk upgrade --with-kubernetes v1.22.0
```
Upgrade a cluster from a specified configuration file.
```
$ kk upgrade -f config-example.yaml
```
Upgrade a cluster using a KubeKey artifact (in an offline enviroment).
```
$ kk upgrade -f config-example.yaml -a kubekey-artifact.tar.gz
```