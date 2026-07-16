# image_registry

This section describes how to install a private image registry separately. The image registry is installed via docker-compose, supporting both `harbor` and `docker-registry` types.

During installation, the tar tool is relied upon for compressing and decompressing packages, and the Docker service depends on iptables to manage container network rules. Please ensure these commands are pre-installed in the system environment beforehand.

## requirement

- One or more computers running a Linux OS compatible with deb/rpm; for example: Ubuntu or CentOS.
- At least 8 GB of memory per machine; applications will be limited when memory is insufficient.
- At least 4 CPUs on computers used as control plane nodes.
- Full network connectivity between all computers in the cluster. You can use public or private networks.
- When using local storage, computers need 100G of high-speed storage disk space.

## Install Harbor

### Build Inventory
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: # your can set all nodes here. or set nodes on special groups.
    harbor1:
      connetor:
        host: 172.16.66.6
      internal_ipv4: 172.16.66.6
    # harbor2:
    #   connector:
    #     host: 172.16.66.7
    #     private_key: ~/.ssh/id_rsa
    #   internal_ipv4: 172.16.66.6
  groups:
    image_registry:
      hosts:
        - harbor1
  vars:
    zone: cn
    image_registry:
      type: harbor
      auth:
        registry: dockerhub.kubekey.local
        # plain_http: false
        password: Harbor12345

```

| Field | Type | Required | Description |
|------|------|------|------|
| `spec.hosts` | Object | Yes | Host list, key is the host name, value is the host configuration |
| `spec.hosts.<name>.connector` | Object | Yes | Host connection configuration |
| `spec.hosts.<name>.connector.host` | String | Yes | SSH target host IP or domain name |
| `spec.hosts.<name>.connector.private_key` | String | No | SSH private key path, uses system default key by default |
| `spec.hosts.<name>.internal_ipv4` | String | No | Internal IPv4 address of the host, used for /etc/hosts domain resolution |
| `spec.groups` | Object | Yes | Node group configuration |
| `spec.groups.image_registry` | Object | Yes | Image registry node group, specifies which hosts are used to deploy the image registry |
| `spec.groups.image_registry.hosts` | Array | Yes | Image registry node name list |
| `spec.vars` | Object | No | Global variable configuration |
| `spec.vars.zone` | boolean | No | Download zone for files and images. If access to GitHub/GoogleAPIs is restricted, please set this to cn |
| `spec.vars.image_registry` | Object | No | Image registry-related configuration |
| `spec.vars.image_registry.type` | String | No | Image registry type, supports `harbor` or `docker-registry` |
| `spec.vars.image_registry.harbor.http_port` | Integer | No | Harbor HTTP service port. Derived from `auth.registry` when `plain_http=true`, or `80` if no port is specified; empty when `plain_http=false` |
| `spec.vars.image_registry.harbor.https_port` | Integer | No | Harbor HTTPS service port. Derived from `auth.registry` when `plain_http=false`, or `443` if no port is specified; empty when `plain_http=true` |
| `spec.vars.image_registry.auth` | Object | No | Image registry authentication configuration |
| `spec.vars.image_registry.auth.plain_http` | Boolean | No | Whether to use plain HTTP (no TLS). Defaults to `false` |
| `spec.vars.image_registry.auth.registry` | String | No | Image registry domain name, format `host:port/project`; the port is used to derive Harbor listening ports |
| `spec.vars.image_registry.auth.password` | String | No | Image registry login password, defaults to Harbor12345 |

### Installation
Harbor is the default image registry.
1. Precheck before installation
    ```shell
    kk precheck image_registry -i inventory.yaml
    ```
2. Installation

```shell
kk init registry -i inventory.yaml
```

### Harbor High Availability

Harbor HA can be implemented in two ways.

1. All Harbor instances share a single storage service.
Official method, suitable for installation within a Kubernetes cluster. Requires separate PostgreSQL and Redis services.  
Reference: https://goharbor.io/docs/edge/install-config/harbor-ha-helm/

2. Each Harbor has its own storage service.
KubeKey method, suitable for server deployment.
![ha-harbor](../../images/ha-harbor.png)
- load balancer: implemented via Docker Compose deploying keepalived.
- harbor service: implemented via Docker Compose deploying Harbor.
- sync images: achieved using Harbor replication.

#### Add HA configuration in inventory.yaml
```yaml!
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: # your can set all nodes here. or set nodes on special groups.
    harbor1:
      connector:
        host: 172.16.66.6
      internal_ipv4: 172.16.66.6
    harbor2:
      connector:
        host: 172.16.66.7
        private_key: /root/.ssh/id_rsa
      internal_ipv4: 172.16.66.7
  groups:
    image_registry:
      hosts:
        - harbor1
        - harbor2
  vars:
    zone: cn
    image_registry:
      ha_vip: 172.16.66.8
      type: harbor
      auth:
        registry: dockerhub.kubekey.local
        # plain_http: false
```

Inventory field descriptions:

| Field | Type | Required | Description |
|------|------|------|------|
| `spec.hosts` | Object | Yes | Host list, key is the host name, value is the host configuration |
| `spec.hosts.<name>.connector` | Object | Yes | Host connection configuration |
| `spec.hosts.<name>.connector.host` | String | Yes | SSH target host IP or domain name |
| `spec.hosts.<name>.connector.private_key` | String | No | SSH private key path, uses system default key by default |
| `spec.hosts.<name>.internal_ipv4` | String | No | Internal IPv4 address of the host, used for /etc/hosts domain resolution |
| `spec.groups` | Object | Yes | Node group configuration |
| `spec.groups.image_registry` | Object | Yes | Image registry node group, specifies which hosts are used to deploy the image registry |
| `spec.groups.image_registry.hosts` | Array | Yes | Image registry node name list, HA requires 2 or more nodes |
| `spec.vars` | Object | No | Global variable configuration |
| `spec.vars.zone` | String | No | Download zone for files and images. If access to GitHub/GoogleAPIs is restricted, please set this to `cn` |
| `spec.vars.image_registry` | Object | No | Image registry-related configuration |
| `spec.vars.image_registry.ha_vip` | String | Yes (HA scenario) | Virtual IP for load balancing, used as the unified access entry for the image registry |
| `spec.vars.image_registry.type` | String | No | Image registry type, supports `harbor` or `docker-registry` |
| `spec.vars.image_registry.harbor.http_port` | Integer | No | Harbor HTTP service port. Derived from `auth.registry` when `plain_http=true`, or `80` if no port is specified; empty when `plain_http=false` |
| `spec.vars.image_registry.harbor.https_port` | Integer | No | Harbor HTTPS service port. Derived from `auth.registry` when `plain_http=false`, or `443` if no port is specified; empty when `plain_http=true` |
| `spec.vars.image_registry.auth` | Object | No | Image registry authentication configuration |
| `spec.vars.image_registry.auth.plain_http` | Boolean | No | Whether to use plain HTTP (no TLS). Defaults to `false` |
| `spec.vars.image_registry.auth.registry` | String | No | Image registry access domain name, corresponding to the VIP domain name for external client access. Format `host:port/project`; the port is used to derive Harbor listening ports. (The actually deployed Harbor uses inventory_hostname as the internal domain name) |

> **Notes for HA configuration:**
> - Multiple nodes must be set in the `image_registry` group for multi-instance deployment.
> - `ha_vip` must be in the same subnet as the nodes and must not be already in use.

Execute the following command to create a high-availability Harbor cluster:
```shell
./kk init registry -i inventory.yaml 
```

## Install Registry

### Build Inventory

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    # localhost:
    #   connector:
    #     password: 123456
    # node1:
    #   connector:
    #     type: ssh
    #     host: node1
    #     port: 22
    #     user: root
    #     password: 123456
  groups:
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    kube_control_plane:
      hosts:
        - localhost
    kube_worker:
      hosts:
        - localhost
    etcd:
      hosts:
        - localhost
    image_registry:
      hosts:
        - localhost
#    nfs:
#      hosts:
#        - localhost
```

### Build Registry Image Package

KubeKey does not provide an offline registry image package. Manual packaging is required.

```shell
# download registry images
docker pull registry:{{ .docker_registry_version }}
# package image
docker save -o docker-registry-{{ .docker_registry_version }}-linux-{{ .binary_type }}.tgz registry:{{ .docker_registry_version }}
# move image to workdir
mv docker-registry-{{ .docker_registry_version }}-linux-{{ .binary_type }}.tgz {{ .binary_dir }}/image-registry/docker-registry/{{ .docker_registry_version }}/{{ .binary_type }}/
```

- `binary_type`: machine architecture (amd64 or arm64, auto-detected via `gather_fact`)
- `binary_dir`: software package storage path, usually `{{ .work_dir }}/kubekey`.

### Installation

Set `image_registry.type` to `docker-registry` to install the registry.

1. Precheck
```shell
kk precheck image_registry -i inventory.yaml --set image_registry.type=docker-registry --set docker_registry_version=2.8.3,docker_version=24.0.7,dockercompose_version=v2.20.3
```

2. Installation
- Standalone installation
```shell
kk init registry -i inventory.yaml --set image_registry.type=docker-registry --set docker_registry_version=2.8.3,docker_version=24.0.7,dockercompose_version=v2.20.3 --set artifact.artifact_url.docker_registry.amd64=docker-registry-2.8.3-linux.amd64.tgz
```

- Automatic installation during cluster creation
```shell
kk create cluster -i inventory.yaml --set image_registry.type=docker-registry --set docker_registry_version=2.8.3,docker_version=24.0.7,dockercompose_version=v2.20.3 --set artifact.artifact_url.docker_registry.amd64=docker-registry-2.8.3-linux.amd64.tgz
```

### Registry High Availability

![ha-registry](../../images/ha-registry.png)
- load balancer: implemented via Docker Compose deploying keepalived.
- registry service: implemented via Docker Compose deploying registry.
- storage service: Registry HA can be achieved using shared storage. Docker Registry supports multiple storage backends, including:
  - **filesystem**: local storage. By default, Docker Registry uses local disk to store image data. If HA is required, you can mount the local storage directory to NFS or other shared storage. Example:
      ```yaml
      image_registry:
        docker_registry:
          storage:
            filesystem:
              rootdir: /opt/docker-registry/data
              nfs_mount: /repository/docker-registry # optional, mount rootdir to NFS server
      ```
      You need to configure and mount the shared directory on `nfs` nodes to ensure data consistency across all registry instances.
  
  - **azure**: Use Azure Blob Storage as backend. Suitable for Azure cloud deployments. Example:
      ```yaml
      image_registry:
        docker_registry:
          storage:
            azure:
              accountname: <your-account-name>
              accountkey: <your-account-key>
              container: <your-container-name>
      ```
  
  - **gcs**: Use Google Cloud Storage as backend. Suitable for GCP deployments. Example:
      ```yaml
      image_registry:
        docker_registry:
          storage:
            gcs:
              bucket: <your-bucket-name>
              keyfile: /path/to/keyfile.json
      ```
  
  - **s3**: Use Amazon S3 or S3-compatible storage as backend. Suitable for AWS or private clouds supporting S3 protocol. Example:
      ```yaml
      image_registry:
        docker_registry:
          storage:
            s3:
              accesskey: <your-access-key>
              secretkey: <your-secret-key>
              region: <your-region>
              bucket: <your-bucket-name>
      ```

> **Note:**  
> 1. When using shared storage (such as NFS, S3, GCS, Azure Blob), it is recommended to deploy at least 2 or more registry instances, and use load balancing (e.g. keepalived+nginx) to achieve HA access.  
> 2. When configuring shared storage, ensure that each registry node has read/write permissions and network connectivity to the storage.
 
## Uninstall Private Image Registry

```shell
./kk delete registry -i inventory.yaml --all --with-data
```

| Parameter | Description |
|------|------|
| `--all` | Uninstall all other related components, including Docker service, DNS configuration |
| `--with-data` | Also delete the image registry data directory (such as Harbor data, registry storage); if not specified, data is retained |
