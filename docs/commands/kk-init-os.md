# NAME
**kk init os**: Init operating system.

# DESCRIPTION
Init operating system. This command will install `openssl`, `socat`, `conntrack`, `ipset`, `ebtables` and `chrony`  on all the nodes.

# OPTIONS

## **--debug**
Print detailed information. The default is `false`.

## **--filename, -f**
Path to a configuration file.

# EXAMPLES
Init the operationg system from a specified configuration file.
```
$ kk init os -f config-example.yaml
```


