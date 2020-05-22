# Storage Client
## NFS
```shell script
# Debian / Ubuntu
apt install nfs-common

# Centos / Redhat
yum install nfs-utils   
```
> Recommended nfs server configuration:  *(rw,insecure,sync,no_subtree_check,no_root_squash)
## Ceph
```shell script
# Debian / Ubuntu
apt install ceph-common

# Centos / Redhat
yum install ceph-common  
```
## GlusterFS

  * The following kernel modules must be loaded:
  
      1. dm_snapshot
      2. dm_mirror
      3. dm_thin_pool
     
      For kernel modules, `lsmod | grep <name>` will show you if a given module is present, and `modprobe <name>` will load 
      a given  module.

 * Each node requires that the `mount.glusterfs` command is available. 
  
 * GlusterFS client version installed on nodes should be as close as possible to the version of the server.
 * Take `glusterfs 7.x` as an example.
```shell script
# Debian
wget -O - https://download.gluster.org/pub/gluster/glusterfs/7/rsa.pub | apt-key add -
DEBID=$(grep 'VERSION_ID=' /etc/os-release | cut -d '=' -f 2 | tr -d '"')
DEBVER=$(grep 'VERSION=' /etc/os-release | grep -Eo '[a-z]+')
DEBARCH=$(dpkg --print-architecture)
echo deb https://download.gluster.org/pub/gluster/glusterfs/LATEST/Debian/${DEBID}/${DEBARCH}/apt ${DEBVER} main > /etc/apt/sources.list.d/gluster.list
apt update
apt install glusterfs-client

# Ubuntu
apt install software-properties-common
add-apt-repository ppa:gluster/glusterfs-7
apt update
apt install glusterfs-client

# Centos / Redhat
yum install glusterfs-fuse
```
