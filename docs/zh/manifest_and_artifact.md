#  KubeKey 清单和制品
KubeKey v2.0.0 版本（以下简称 kk ）新增了清单（manifest）和制品（artifact）的概念，为用户离线部署 Kubernetes 集群提供了一种解决方案。在过去，用户需要准备部署工具，镜像 tar 包和其他相关的二进制文件，每位用户需要部署的 Kubernetes 版本和需要部署的镜像都是不同的。现在使用 kk 的话，用户只需使用清单`manifest`文件来定义将要离线部署的集群环境需要的内容，再通过该`manifest`来导出制品`artifact`文件即可完成准备工作。离线部署时只需要 kk 和 `artifact`就可快速、简单的在环境中部署镜像仓库和 Kubernetes 集群。

## 什么是 KubeKey 清单 （manifest）？
`manifest`就是一个描述当前 Kubernetes 集群信息和定义`artifact`制品中需要包含哪些内容的文本文件。目前有两种方式来生成该文件：
* 根据模版手动创建并编写该文件。
* 使用 kk 命令根据已存在的集群生成该文件。

第一种方式需要了解该配置文件中不同字段的更多信息，请参考 [manifest-example.yaml](../manifest-example.md)。

以下内容仅针对于第二种由 kk 实现的生成文件的方式。

### 使用方式
> 注意：
> 该方式需提前准备好一个已经搭建好的集群环境，并为 kk 提供集群的`kubeconfig`文件。

命令如下：
```
./kk create manifest
```
默认情况下 kk 会使用`$HOME/.kube/config`文件，也可指定一个`kubeconfig`文件。
```
./kk create manifest --kubeconfig config
```
执行完毕后将在当前目录下生成`manifest-sample.yaml`文件。然后可根据实际情况修改`manifest-sample.yaml`文件的内容，用以之后导出期望的`artifact`文件。

### 原理
kk 通过 `kubeconfig` 文件连接对应的 Kubernetes 集群，然后检查出集群环境中以下信息：
* 节点架构
* 节点操作系统
* 节点上的镜像
* Kubernetes 版本
* CRI 信息

之后，这些当前集群的描述信息将最终写入`manifest`文件中。同时其他无法检测到的相关文件信息（如：ETCD 集群信息、镜像仓库等）也将按照 kk 推荐的默认值写入到`manifest`文件中。

## 什么是 KubeKey 制品（artifact）？
制品就是一个根据指定的`manifest`文件内容导出的包含镜像 tar 包和相关二进制文件的 tgz 包。在 kk 初始化镜像仓库、创建集群、添加节点和升级集群的命令中均可指定一个`artifact`，kk 将自动解包该`artifact`并将在执行命令时直接使用解包出来的文件。

### 使用方式
#### 导出 artifact
> 注意：
> 1. 导出命令会从互联网中下载相应的二进制文件，请确保网络连接正常。
> 2. 导出命令会根据`manifest`文件中的镜像列表逐个拉取镜像，请确保 kk 的工作节点已安装containerd 或最低版本为 18.09 的 docker。
> 3. kk 会解析镜像列表中的镜像名，若镜像名中的镜像仓库需要鉴权信息，可在`manifest`文件中的`.registry.auths`字段中进行配置。
> 4. 若需要导出的`artifact`文件中包含操作系统依赖文件（如：conntarck、chrony等），可在`operationSystem`元素中的`.repostiory.iso.url`中配置相应的 ISO 依赖文件下载地址。
* 导出。
```
./kk artifact export -m manifest-sample.yaml
```
执行完毕后将在当前目录下生成`kubekey-artifact.tar.gz`文件。

#### 使用 artifact
> 注意：
> 1. 在离线环境中，使用`artifact`前需要先使用 kk 生成`config-sample.yaml`文件并配置相应的信息。
> 2. 在离线环境中，使用创建集群和升级集群命令时默认情况下会将`artifact`中的镜像推送至私有镜像仓库，若私有镜像仓库需要鉴权信息，可在`config-sample.yaml`文件中的`.spec.registry.auths`字段中进行配置。

* 初始化镜像仓库，相关配置可参考 [容器镜像仓库](../registry.md)。
```
./kk init registry -f config-sample.yaml -a kubekey-artifact.tar.gz
```
* 推送镜像到私有镜像仓库
```
./kk artifact image push -f config-sample.yaml -a kubekey-artifact.tar.gz
```
* 创建集群。
> 注意：在离线环境中，需配置私有镜像仓库信息用于集群镜像的管理，相关配置可参考 [config-sample.yaml](../config-example.md) 和 [容器镜像仓库](../registry.md)。
```
./kk create cluster -f config-sample.yaml -a kubekey-artifact.tar.gz
```
* 创建集群，并且安装操作系统依赖（需要`artifact`中包含目标集群中节点的操作系统依赖文件）。
```
./kk create cluster -f config-sample.yaml -a kubekey-artifact.tar.gz --with-packages
```
* 添加节点。
```
./kk add nodes -f config-sample.yaml -a kubekey-artifact.tar.gz
```
* 升级集群。
```
./kk upgrade -f config-sample.yaml -a kubekey-artifact.tar.gz
```