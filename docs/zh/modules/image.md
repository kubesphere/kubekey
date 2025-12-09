# image 模块

image模块允许用户下载镜像到本地目录或上传镜像到远程目录。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| pull | 把镜像从远程仓库中拉取到本地目录 | map | 否 | - |
| pull.images_dir | 镜像存放的本地目录 | 字符串 | 否 | - |
| pull.manifests | 需要拉取的镜像列表 | 字符串数组 | 是 | - |
| pull.auths | 远程仓库的认证信息 | Object数组 | 否 | - |
| pull.auths.repo | 用于认证远程仓库的地址 | 字符串 | 否 | - |
| pull.auths.username | 用于认证远程仓库的用户名 | 字符串 | 否 | - |
| pull.auths.password | 用于认证远程仓库的密码 | 字符串 | 否 | - |
| pull.auths.insecure | 是否跳过当前远程仓库的tls认证 | bool | 否 | - |
| pull.auths.plain_http | 是否使用http访问远程仓库 | bool | 否 | - |
| pull.platform | 镜像的架构信息 | 字符串数组 | 否 | - |
| pull.skip_tls_verify | 默认的是否跳过远程仓库的tls认证 | bool | 否 | - |
| push | 从本地目录中推送镜像到远程仓库 | map | 否 | - |
| push.images_dir | 镜像存放的本地目录 | 字符串 | 否 | - |
| push.auths | 远程仓库的认证信息 | Object数组 | 否 | - |
| push.auths.repo | 用于认证远程仓库的地址 | 字符串 | 否 | - |
| push.auths.username | 用于认证远程仓库的用户名 | 字符串 | 否 | - |
| push.auths.password | 用于认证远程仓库的密码 | 字符串 | 否 | - |
| push.auths.insecure | 是否跳过当前远程仓库的tls认证 | bool | 否 | - |
| push.auths.plain_http | 是否使用http访问远程仓库 | bool | 否 | - |
| push.skip_tls_verify | 默认的是否跳过远程仓库的tls认证 | bool | 否 | - |
| push.src_pattern | 正则表达式，过滤本地目录中存放的镜像 | map | 否 | - |
| push.dest | 模版语法，从本地目录镜像推送到的远程仓库镜像 | map | 否 | - |
| copy | 模版语法，将镜像在文件系统和镜像仓库内相互复制 | map | 否 | - |
| copy.platform | 镜像的架构信息 | 字符串数组 | 否 | - |
| copy.from | 模版语法，源镜像信息 | map | 否 | - |
| copy.from.path | 镜像源文件路径 | 字符串 | 否 | - |
| copy.from.manifests | 源镜像列表 | 字符串数组 | 否 | - |
| copy.to | 模版语法，从本地目录镜像推送到的远程仓库镜像 | map | 否 | - |
| copy.to.path | 镜像目标文件路径 | 字符串 | 否 | - |
| copy.to.pattern | 正则表达式，过滤复制至目标的镜像 | 字符串 | 否 | - |


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
      platform:
        - amd64
        - arm64
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

3. 将镜像从文件系统复制至其他文件系统
```yaml
- name: file to file
  image:
    copy:
      from:
        path: "/tmp/images/"
        manifests:
          - docker.io/calico/apiserver:v3.28.2
      to:
        path: "/tmp/others/images/"
```