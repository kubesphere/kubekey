# image 模块

拉取镜像到本地目录、推送镜像到远程仓库，或在文件系统与镜像仓库间复制镜像。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| manifests | 要操作的镜像列表 | 字符串数组 | 是 | - |
| platform | 架构列表（如 ["linux/amd64", "linux/arm64"]） | 字符串数组 | 否 | - |
| policy | 平台过滤策略：`strict`（所有平台必须存在，否则报错），`warn`（缺失时记录警告但继续执行） | 字符串 | 否 | strict |
| pattern | 正则匹配镜像 | 字符串 | 否 | - |
| auths | 仓库认证信息 | Object 数组 | 否 | - |
| auths.repo | 仓库地址 | 字符串 | 否 | - |
| auths.username | 用户名 | 字符串 | 否 | - |
| auths.password | 密码 | 字符串 | 否 | - |
| auths.insecure | 是否跳过 TLS 校验 | bool | 否 | - |
| auths.plain_http | 是否使用 HTTP | bool | 否 | - |
| src | 源镜像引用（远程仓库或本地目录，如 `docker.io/library/alpine:3.19` 或 `local:///var/lib/kubekey/images`） | 字符串 | 否 | - |
| dest | 目标位置（本地目录或远程仓库，如 `local:///tmp/images/` 或 `hub.kubekey/library/alpine:3.19`） | 字符串 | 否 | - |
| skip_tls_verify | 默认是否跳过 TLS 校验 | bool | 否 | - |

**src/dest 格式：**
- 远程仓库：`registry/repository:tag`（如 `docker.io/library/alpine:3.19`）
- 本地目录：`local:///绝对路径`（如 `local:///var/lib/kubekey/images`）

**操作类型（由 src 和 dest 决定）：**
- **pull**：从远程仓库拉取到本地目录（`src` = 远程仓库，`dest` = 本地目录）
- **push**：从本地目录推送到远程仓库（`src` = 本地目录，`dest` = 远程仓库）
- **copy**：本地目录间复制（`src` = 本地目录，`dest` = 本地目录）

**dest 可用变量：**

- `{{ .module.image.src.reference.registry }}`：registry
- `{{ .module.image.src.reference.repository }}`：repository
- `{{ .module.image.src.reference.reference }}`：reference（如 tag）

**本地目录结构示例：**

```text
images_dir/
├── registry1/
│   ├── image1/manifests/reference
│   └── image2/manifests/reference
└── registry2/
    └── image1/manifests/reference
```

## 示例

**1. 从远程仓库拉取镜像**

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

**2. 推送镜像到远程仓库**

```yaml
- name: push images
  image:
    manifests:
      - "library/alpine:3.19"
    src: "local:///tmp/images/"
    dest: "hub.kubekey/{{ .module.image.src.reference.repository }}:{{ .module.image.src.reference.reference }}"
```

例如：

- `library/alpine:3.19` → `hub.kubekey/library/alpine:3.19`

**3. 本地目录间复制镜像**

```yaml
- name: copy images
  image:
    manifests:
      - "docker.io/calico/apiserver:v3.28.2"
    src: "local:///tmp/images/"
    dest: "local:///tmp/others/images/"
```

**4. 使用正则模式复制镜像**

```yaml
- name: copy images with pattern
  image:
    pattern: ".*calico.*"
    src: "local:///tmp/images/"
    dest: "local:///tmp/others/images/"
```
