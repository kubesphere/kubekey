# NAME
**kk cri migrate**: migrate your cri smoothly to docker/containerd with this command.

# DESCRIPTION
migrate your cri smoothly to docker/containerd with this command.

# OPTIONS

## **--role**
Which node(worker/master/all) to migrate.

## **--type**
Which cri(docker/containerd) to migrate.

## **--debug**
Print detailed information. The default is `false`.

## **--filename, -f**
Path to a configuration file.This option is required.

## **--yes, -y**
Skip confirm check. The default is `false`.

# EXAMPLES
Migrate your master node's cri smoothly to docker.
```
$ ./kk cri migrate --role master --type docker -f config-sample.yaml
```
Migrate your master node's cri smoothly to containerd.
```
$ ./kk cri migrate --role master --type containerd -f config-sample.yaml
```
Migrate your worker node's cri smoothly to docker.
```
$ ./kk cri migrate --role worker --type docker -f config-sample.yaml
```
Migrate your worker node's cri smoothly to containerd.
```
$ ./kk cri migrate --role worker --type containerd -f config-sample.yaml
```
Migrate all your node's cri smoothly to containerd.
```
$ ./kk cri migrate --role all --type containerd -f config-sample.yaml
```
Migrate all your node's cri smoothly to docker.
```
$ ./kk cri migrate --role all --type docker -f config-sample.yaml
```