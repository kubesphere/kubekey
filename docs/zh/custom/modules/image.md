# image 模块

拉取镜像到本地目录、推送镜像到远程仓库，或在文件系统与镜像仓库间复制镜像。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| pull | 从远程仓库拉取到本地目录 | map | 否 | - |
| pull.images_dir | 镜像存放的本地目录 | 字符串 | 否 | - |
| pull.manifests | 要拉取的镜像列表 | 字符串数组 | 是 | - |
| pull.auths | 远程仓库认证 | Object 数组 | 否 | - |
| pull.auths.repo | 仓库地址 | 字符串 | 否 | - |
| pull.auths.username | 用户名 | 字符串 | 否 | - |
| pull.auths.password | 密码 | 字符串 | 否 | - |
| pull.auths.insecure | 是否跳过 TLS 校验 | bool | 否 | - |
| pull.auths.plain_http | 是否使用 HTTP | bool | 否 | - |
| pull.platform | 架构列表（如 amd64、arm64） | 字符串数组 | 否 | - |
| pull.skip_tls_verify | 默认是否跳过 TLS 校验 | bool | 否 | - |
| push | 从本地目录推送到远程仓库 | map | 否 | - |
| push.images_dir | 镜像存放的本地目录 | 字符串 | 否 | - |
| push.auths | 远程仓库认证 | Object 数组 | 否 | - |
| push.auths.repo / .username / .password / .insecure / .plain_http | 同上 | - | - | - |
| push.skip_tls_verify | 默认是否跳过 TLS 校验 | bool | 否 | - |
| push.src_pattern | 正则，过滤要推送的镜像 | 字符串 | 否 | - |
| push.dest | 目标镜像，支持 [模板语法](../101-syntax.md) | 字符串 | 否 | - |
| copy | 在文件系统 / 镜像仓库间复制 | map | 否 | - |
| copy.platform | 架构列表 | 字符串数组 | 否 | - |
| copy.from | 源 | map | 否 | - |
| copy.from.path | 源目录路径 | 字符串 | 否 | - |
| copy.from.manifests | 源镜像列表 | 字符串数组 | 否 | - |
| copy.to | 目标 | map | 否 | - |
| copy.to.path | 目标路径 | 字符串 | 否 | - |
| copy.to.pattern | 正则，过滤复制目标 | 字符串 | 否 | - |

**push.dest** 可用变量：

- `{{ .module.image.src.reference.registry }}`：registry
- `{{ .module.image.src.reference.repository }}`：repository
- `{{ .module.image.src.reference.reference }}`：reference（如 tag）

**本地目录结构示例**：

```text
images_dir/
├── registry1/
│   ├── image1/manifests/reference
│   └── image2/manifests/reference
└── registry2/
    └── image1/manifests/reference
```

## 示例

**1. 拉取镜像**

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

**2. 推送镜像到远程仓库**

```yaml
- name: push images
  image:
    push:
      images_dir: /tmp/images/
      dest: "hub.kubekey/{{ .module.image.src.reference.repository }}:{{ .module.image.src.reference.reference }}"
```

例如：

- `docker.io/kubesphere/ks-apiserver:v4.1.3` → `hub.kubekey/kubesphere/ks-apiserver:v4.1.3`
- `docker.io/kubesphere/ks-console:3.19` → `hub.kubekey/kubesphere/ks-console:3.19`

**3. 文件系统间复制镜像**

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
