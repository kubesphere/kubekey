# image Module

The image module allows users to pull images to a local directory or push images to a remote repository.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| pull | Pull images from a remote repository to a local directory | map | No | - |
| pull.images_dir | Local directory to store images | string | No | - |
| pull.manifests | List of images to pull | array of strings | Yes | - |
| pull.auths | Authentication information for the remote repository | array of objects | No | - |
| pull.auths.repo | Repository address for authentication | string | No | - |
| pull.auths.username | Username for repository authentication | string | No | - |
| pull.auths.password | Password for repository authentication | string | No | - |
| pull.platform | Image platform/architecture | string | No | - |
| pull.skip_tls_verify | Skip TLS verification for the remote repository | bool | No | - |
| push | Push images from a local directory to a remote repository | map | No | - |
| push.images_dir | Local directory storing images | string | No | - |
| push.username | Username for remote repository authentication | string | No | - |
| push.password | Password for remote repository authentication | string | No | - |
| push.skip_tls_verify | Skip TLS verification for the remote repository | bool | No | - |
| push.src_pattern | Regex to filter images in the local directory | map | No | - |
| push.dest | Template syntax for the destination remote repository image | map | No | - |

Each image in the local directory corresponds to a dest image.
```txt
|-- images_dir/
|   |-- registry1/
|   |   |-- image1/
|   |   |   |-- manifests/
|   |   |   |   |-- reference
|   |   |-- image2/
|   |   |   |-- manifests/
|   |   |   |   |-- reference
|   |-- registry2/
|   |   |-- image1/
|   |   |   |-- manifests/
|   |   |   |   |-- reference
```
For each src image, there is a corresponding dest. The dest template supports the following variables:
{{ .module.image.src.reference.registry }}: registry of the local image
{{ .module.image.src.reference.repository }}: repository of the local image
{{ .module.image.src.reference.reference }}: reference of the local image

## Usage Examples

1. Pull images
```yaml
- name: pull images
  image:
    pull:
      images_dir: /tmp/images/
      platform: linux/amd64
      manifests:
       - "docker.io/kubesphere/ks-apiserver:v4.1.3"
       - "docker.io/kubesphere/ks-controller-manager:v4.1.3"
       - "docker.io/kubesphere/ks-console:3.19"
```

2. Push images to a remote repository
```yaml
- name: push images
  push:
    images_dir: /tmp/images/
    dest: hub.kubekey/{{ .module.image.src.reference.repository }}:{{ .module.image.src.reference.reference }}
```
For example:
docker.io/kubesphere/ks-apiserver:v4.1.3 => hub.kubekey/kubesphere/ks-apiserver:v4.1.3
docker.io/kubesphere/ks-controller-manager:v4.1.3 => hub.kubekey/kubesphere/ks-controller-manager:v4.1.3
docker.io/kubesphere/ks-console:3.19 => hub.kubekey/kubesphere/ks-console:v4.1.3