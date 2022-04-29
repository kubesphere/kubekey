# NAME
**kk init os**: Init operating system.

# DESCRIPTION
Init operating system. This command will install `openssl`, `socat`, `conntrack`, `ipset`, `ipvsadm`, `ebtables` and `chrony`  on all the nodes.

# OPTIONS

## **--artifact, -a**
Path to a KubeKey artifact.

## **--debug**
Print detailed information. The default is `false`.

## **--filename, -f**
Path to a configuration file.

# EXAMPLES
Init the operating system from a specified configuration file.
```
$ kk init os -f config-example.yaml
```
Init the operating system from a specified configuration file and use a KubeKey artifact.
```
$ kk init os -f config-example.yaml -a kubekey-artifact.tar.gz
```

