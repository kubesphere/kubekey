# NAME
**kk init registry**: Init a local image registry.

# DESCRIPTION
Init a local image registry. More information about the registry can be found [here](../registry.md).

# OPTIONS

## **--artifact, -a**
Path to a KubeKey artifact.

## **--debug**
Print detailed information. The default is `false`.

## **--download-cmd**
The user defined command to download the necessary binary files. The first param `%s` is output path, the second param `%s`, is the URL. The default is `curl -L -o %s %s`.

## **--filename, -f**
Path to a configuration file.

# EXAMPLES
Init a local registry from a specified configuration file.
```
$ kk init registry -f config-example.yaml
```
Init a local registry from a specified configuration file and use a KubeKey artifact.
```
$ kk init registry -f config-example.yaml -a kubekey-artifact.tar.gz
```

