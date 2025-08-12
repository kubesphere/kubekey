# image 模块

image模块允许用户下载镜像到本地目录或上传镜像到远程目录。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| pull | 把镜像从远程仓库中拉取到本地目录 | map | 否 | - |
| pull.images_dir | 镜像存放的本地目录 | 字符串 | 否 | - |
| pull.manifests | 需要拉取的镜像列表 | 字符串数组 | 是 | - |
| pull.username | 用于认证远程仓库的用户 | 字符串 | 否 | - |
| pull.password | 用于认证远程仓库的密码 | 字符串 | 否 | - |
| pull.platform | 镜像的架构信息 | 字符串 | 否 | - |
| pull.skip_tls_verify | 是否跳过远程仓库的tls认证 | bool | 否 | - |
| push | 从本地目录中推送镜像到远程仓库 | map | 否 | - |
| push.images_dir | 镜像存放的本地目录 | 字符串 | 否 | - |
| push.username | 用于认证远程仓库的用户 | 字符串 | 否 | - |
| push.password | 用于认证远程仓库的密码 | 字符串 | 否 | - |
| push.skip_tls_verify | 是否跳过远程仓库的tls认证 | bool | 否 | - |
| push.src_pattern | 正则表达式，过滤本地目录中存放的镜像 | map | 否 | - |
| push.dest | 模版语法，从本地目录镜像推送到的远程仓库镜像 | map | 否 | - |

每个本地目录存放的镜像对应一个dest镜像。
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
对于每个src镜像对应有一个dest。dest中有如下默认的模板变量
{{ .module.image.src.reference.registry }}: 本地目录中单个镜像的registry
{{ .module.image.src.reference.repository }}: 本地目录中单个镜像的repository
{{ .module.image.src.reference.reference }}: 本地目录中单个镜像的reference

## 使用示例

1. 拉取镜像
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

2. 推送镜像到远程仓库
```yaml
- name: push images
  push:
    images_dir: /tmp/images/
    dest: hub.kubekey/{{ .module.image.src.reference.repository }}:{{ .module.image.src.reference.reference }}
```
即：
docker.io/kubesphere/ks-apiserver:v4.1.3 => hub.kubekey/kubesphere/ks-apiserver:v4.1.3
docker.io/kubesphere/ks-controller-manager:v4.1.3 => hub.kubekey/kubesphere/ks-controller-manager:v4.1.3
docker.io/kubesphere/ks-console:3.19 => hub.kubekey/kubesphere/ks-console:v4.1.3