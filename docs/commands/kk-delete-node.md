# NAME
**kk delete node**: Delete a node.

# DESCRIPTION
Delete a node. This command will use the `kubectl drain` to safely evict all pods, and then use `kubectl delete node`  to delete the specified node.

# OPTIONS

## **--debug**
Print detailed information. The default is `false`.

## **--filename, -f**
Path to a configuration file.

# EXAMPLES
Delete a node named `node2` from a specified configuration file.
```
$ kk delete node node2 -f config-example.yaml
```


