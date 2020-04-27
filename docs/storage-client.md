# Storage Client
## NFS
```shell script
# Debian / Ubuntu
apt install nfs-common

# Centos / Redhat
yum install nfs-utils   
```
## Ceph
```shell script
# Debian / Ubuntu
apt install ceph-common

# Centos / Redhat
yum install ceph-common  
```
## GlusterFS
```shell script
# Debian
wget -O - https://download.gluster.org/pub/gluster/glusterfs/01.old-releases/3.12/rsa.pub | apt-key add -
DEBID=$(grep 'VERSION_ID=' /etc/os-release | cut -d '=' -f 2 | tr -d '"') &&
DEBVER=$(grep 'VERSION=' /etc/os-release | grep -Eo '[a-z]+') &&
DEBARCH=$(dpkg --print-architecture) &&
echo deb https://download.gluster.org/pub/gluster/glusterfs/01.old-releases/3.12/LATEST/Debian/${DEBID}/${DEBARCH}/apt ${DEBVER} main > /etc/apt/sources.list.d/gluster.list
apt update
apt install glusterfs-client

# Ubuntu
apt install software-properties-common
add-apt-repository ppa:gluster/glusterfs-3.12
apt update
apt install glusterfs-server

# Centos / Redhat
yum install glusterfs-client 
```
