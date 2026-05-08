# 依赖包管理

KubeKey 为以下发行版预编译了系统依赖包，可在 [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest) 获取：

- AlmaLinux 9.0
- CentOS 8
- Debian 10 / 11
- Kylin V10SP1 / V10SP2 / V10SP3 / V10SP3-2403
- Ubuntu 18.04 / 20.04 / 22.04 / 24.04

> 注：部分发行版（如 CentOS 7、Ubuntu 16.04）在仓库中存在 Dockerfile，但因版本过老、官方已停止维护等原因，未纳入 [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest) 的自动构建。具体请以 `.github/workflows/gen-repository-iso.yaml` 中的构建矩阵为准。

## 依赖包制作流程

各发行版的依赖包制作脚本位于 [`hack/gen-repository-iso`](../../hack/gen-repository-iso) 目录下，其构建产物由 `.github/workflows/gen-repository-iso.yaml` 自动发布到 [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest)。

### 文件说明

- `dockerfile.<os>`：各发行版对应的 Dockerfile，用于在容器内打包依赖。
- `download-pkgs.sh`：通用下载脚本。
- `packages.yaml`：声明所有需要预编译的依赖包列表。

### 手动构建

如需自定义依赖包，可进入 `hack/gen-repository-iso` 目录后执行：

```shell
docker build -f dockerfile.<os> -t kk-iso .
```

或使用 Docker Buildx 构建多架构镜像并导出产物：

```shell
docker buildx build --platform linux/amd64,linux/arm64 -f dockerfile.<os> -o ./output .
```

构建完成后，可在输出目录中找到对应的 `.iso` 文件。