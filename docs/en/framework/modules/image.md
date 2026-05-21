# image Module

Pull images to local directory, push images to remote registry, or copy images between file system and image registry.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| manifests | List of images to operate on | string array | Yes | - |
| platform | Architecture list (e.g., ["linux/amd64", "linux/arm64"]) | string array | No | - |
| policy | Policy for platform filtering: `strict` (all platforms must exist, otherwise error), `warn` (log warning if missing, but continue) | string | No | strict |
| pattern | Regex pattern to match images | string | No | - |
| auths | Registry authentication | Object array | No | - |
| auths.repo | Registry address | string | No | - |
| auths.username | Username | string | No | - |
| auths.password | Password | string | No | - |
| auths.insecure | Whether to skip TLS verification | bool | No | - |
| auths.plain_http | Whether to use HTTP | bool | No | - |
| src | Source image reference (remote registry or local directory, e.g., `docker.io/library/alpine:3.19` or `local:///var/lib/kubekey/images`) | string | No | - |
| dest | Destination (local directory or remote registry, e.g., `local:///tmp/images/` or `hub.kubekey/library/alpine:3.19`) | string | No | - |
| skip_tls_verify | Default whether to skip TLS verification | bool | No | - |

**src/dest format:**
- Remote registry: `registry/repository:tag` (e.g., `docker.io/library/alpine:3.19`)
- Local directory: `local:///absolute/path` (e.g., `local:///var/lib/kubekey/images`)

**Operation types (determined by src and dest):**
- **pull**: `src` = remote registry, `dest` = local directory
- **push**: `src` = local directory, `dest` = remote registry
- **copy**: `src` = local directory, `dest` = local directory

**Available variables for dest:**

- `{{ .module.image.src.reference.registry }}`: registry
- `{{ .module.image.src.reference.repository }}`: repository
- `{{ .module.image.src.reference.reference }}`: reference (e.g., tag)

**Local directory structure example:**

```text
images_dir/
├── registry1/
│   ├── image1/manifests/reference
│   └── image2/manifests/reference
└── registry2/
    └── image1/manifests/reference
```

## Examples

**1. Pull images from remote registry**

```yaml
- name: pull images
  image:
    manifests:
      - "docker.io/kubesphere/ks-apiserver:v4.1.3"
      - "docker.io/kubesphere/ks-controller-manager:v4.1.3"
    platform: [linux/amd64, linux/arm64]
    policy: strict
    src: "docker.io"
    dest: "local:///tmp/images/"
```

**2. Push images to remote registry**

```yaml
- name: push images
  image:
    manifests:
      - "library/alpine:3.19"
    src: "local:///tmp/images/"
    dest: "hub.kubekey/{{ .module.image.src.reference.repository }}:{{ .module.image.src.reference.reference }}"
```

For example:

- `library/alpine:3.19` → `hub.kubekey/library/alpine:3.19`

**3. Copy images between local directories**

```yaml
- name: copy images
  image:
    manifests:
      - "docker.io/calico/apiserver:v3.28.2"
    src: "local:///tmp/images/"
    dest: "local:///tmp/others/images/"
```

**4. Copy images with regex pattern**

```yaml
- name: copy images with pattern
  image:
    pattern: ".*calico.*"
    src: "local:///tmp/images/"
    dest: "local:///tmp/others/images/"
```
