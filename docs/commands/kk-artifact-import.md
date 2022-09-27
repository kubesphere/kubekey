# NAME
**kk artifact import**: Import a KubeKey offline installation package.

# DESCRIPTION
The import command will unarchive the KubeKey offline installation package to get all images, specified binaries and Linux repository iso file.

# OPTIONS

## **--artifact, -a**
Path to a artifact gzip. This option is required.

## **--with-packages**
Install operation system packages by artifact

# EXAMPLES
import a KubeKey artifact named `my-artifact.tar.gz`.
```
$ kk artifact import -a my-artifact.tar.gz
```
import a KubeKey artifact named `my-artifact.tar.gz` and install local repository. 
```
$ kk artifact import -a my-artifact.tar.gz --with-packages true
```