# image Module

Pull images to local directory, push images to remote registry, or copy images between file system and image registry.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| pull | Pull from remote registry to local directory | map | No | - |
| pull.images_dir | Local directory for image storage | string | No | - |
| pull.manifests | List of images to pull | string array | Yes | - |
| pull.auths | Remote registry authentication | Object array | No | - |
| pull.auths.repo | Registry address | string | No | - |
| pull.auths.username | Username | string | No | - |
| pull.auths.password | Password | string | No | - |
| pull.auths.insecure | Whether to skip TLS verification | bool | No | - |
| pull.auths.plain_http | Whether to use HTTP | bool | No | - |
| pull.platform | Architecture list (e.g., amd64, arm64) | string array | No | - |
| pull.skip_tls_verify | Default whether to skip TLS verification | bool | No | - |
| push | Push from local directory to remote registry | map | No | - |
| push.images_dir | Local directory for image storage | string | No | - |
| push.auths | Remote registry authentication | Object array | No | - |
| push.auths.repo / .username / .password / .insecure / .plain_http | Same as above | - | - | - |
| push.skip_tls_verify | Default whether to skip TLS verification | bool | No | - |
| push.src_pattern | Regex, filter images to push | string | No | - |
| push.dest | Target image, supports [template syntax](../101-syntax.md) | string | No | - |
| copy | Copy between file system / image registry | map | No | - |
| copy.platform | Architecture list | string array | No | - |
| copy.from | Source | map | No | - |
| copy.from.path | Source directory path | string | No | - |
| copy.from.manifests | Source image list | string array | No | - |
| copy.to | Target | map | No | - |
| copy.to.path | Target path | string | No | - |
| copy.to.pattern | Regex, filter copy targets | string | No | - |

**push.dest** available variables:

- `{{ .module.image.src.reference.registry }}`: registry
- `{{ .module.image.src.reference.repository }}`: repository
- `{{ .module.image.src.reference.reference }}`: reference (e.g., tag)

**Local directory structure example**:

```text
images_dir/
├── registry1/
│   ├── image1/manifests/reference
│   └── image2/manifests/reference
└── registry2/
    └── image1/manifests/reference
```

## Examples

**1. Pull images**

```yaml
- name: pull images
  image:
    pull:
      images_dir: /tmp/images/
      platform: [amd64, arm64]
      manifests:
        - "docker.io/kubesphere/ks-apiserver:v4.1.3"
        - "docker.io/kubesphere/ks-controller-manager:v4.1.3"
        - "docker.io/kubesphere/ks-console:3.19"
```

**2. Push images to remote registry**

```yaml
- name: push images
  image:
    push:
      images_dir: /tmp/images/
      dest: "hub.kubekey/{{ .module.image.src.reference.repository }}:{{ .module.image.src.reference.reference }}"
```

For example:

- `docker.io/kubesphere/ks-apiserver:v4.1.3` → `hub.kubekey/kubesphere/ks-apiserver:v4.1.3`
- `docker.io/kubesphere/ks-console:3.19` → `hub.kubekey/kubesphere/ks-console:3.19`

**3. Copy images between file systems**

```yaml
- name: file to file
  image:
    copy:
      from:
        path: /tmp/images/
        manifests:
          - docker.io/calico/apiserver:v3.28.2
      to:
        path: /tmp/others/images/
```
