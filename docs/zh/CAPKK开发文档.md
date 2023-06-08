# CAPKK 开发者文档

Cluster-API-provider-kubekey（简称：CAPKK）是遵循 cluster-api 的开发者文档 [Provider contracts](https://cluster-api.sigs.k8s.io/developer/providers/contracts.html#provider-contract) 实现的基于 SSH 的基础设施供应商（ infrastructure provider ）。

在整个 CAPKK 项目中主要包含了三种资源及其对应的 controller：KKCluster，KKMachine，KKInstance。

## 项目结构

```bash
├── api
│   └── v1beta1              ## CAPKK 定义的 CR
├── bootstrap                ## k3s 版本的 bootstrap provider
├── config                   ## CAPKK 项目的 Kubernetes yaml 文件
├── controllers              ## CAPKK 包含的 controller
│   ├── kkcluster
│   ├── kkinstance
│   └── kkmachine
├── controlplane             ## k3s 版本的 control-plane provider
├── docs                     ## 文档
├── exp                      ## 实验性功能
├── hack
├── pkg
│   ├── clients              ## CAPKK 用到的客户端
│   │   └── ssh
│   ├── rootfs               ## CAPKK 的数据存储目录
│   ├── scope                ## CAPKK 中按照功能划分定义的不同范围的接口，主要用于拆分 KKCluster、KKMachine、KKInstance 的属性字段。
│   ├── service              ## CAPKK 中各种操作的具体实现
│   │   ├── binary           ## CAPKK 对二进制文件的操作
│   │   ├── bootstrap        ## CAPKK 对机器环境初始化的操作
│   │   ├── containermanager ## CAPKK 对容器运行时的操作
│   │   ├── operation        ## CAPKK 对 Linux 机器的基础操作
│   │   ├── provisioning     ## CAPKK 对 cloud-init 或 ignition 文件进行解析并映射为 SSH 命令
│   │   ├── repository       ## CAPKK 对 rpm 包的操作
│   │   └── util
│   └── util
│       ├── filesystem
│       └── hash
├── scripts
├── templates                ## 用于自动化脚本生成 GitHub release 中的 sample yaml
├── test                     ## CAPKK 的 e2e 测试
│   └── e2e
├── util                     ## KubeKey 的通用工具类
└── version                  ## 包含 KubeKey 的 version 和二进制组件的 SHA256

```

## KKCluster

KKCluster 用于定义集群的全局的基本信息。

```go
type KKClusterSpec struct {
	// 定义 Kubernetes 的发行版，目前支持的参数有 kubernetes，k3s
	Distribution string `json:"distribution,omitempty"`

	// 定义节点的 SSH 认证信息
	Nodes Nodes `json:"nodes"`

	// 定义控制平面的端点
	// 可选
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`

	// 定义控制平面 LoadBalancer 的地址
	ControlPlaneLoadBalancer *KKLoadBalancerSpec `json:"controlPlaneLoadBalancer,omitempty"`

	// 定义集群组件二进制文件的下载地址，通常用于覆写默认地址，离线部署场景
	// 可选
	Component *Component `json:"component,omitempty"`

	// 定义私有镜像仓库信息
	// 可选
	Registry Registry `json:"registry,omitempty"`
}
```

在上述内容中，需注意如下几点：
* nodes 字段填写集群所包含的所有节点的 SSH 信息
* CAPKK  默认使用 kube-vip 实现集群控制平面高可用，因此 controlPlaneLoadBalancer 可填写为网段内的任意为使用的 IP 。
* CAPKK 默认使用二进制组件的官方下载地址，如需使用 KubeSphere 团队维护的国内资源地址可按如下方式配置：
    ```yaml
    spec:
      component:
        zone: cn
    ```
* component 字段还可用于指定自定义的文件服务器地址，并可配置 overrides 数组来修改 CAPKK 在文件服务器中对不同二进制的寻址 path 路径和 SHA256 checksum。

以下为一个常见 kkcluster 的 sample yaml 文件：
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: KKCluster
metadata:
  name: quick-start
spec:
  component:
    zone: ""
  nodes:
    auth:
      user: ubuntu
      password: P@88w0rd
    instances:
      - name: test1
        address: 192.168.0.3
      - name: test2
        address: 192.168.0.4
      - name: test3
        address: 192.168.0.5
  controlPlaneLoadBalancer:
    host: 192.168.0.100
```

在 apply 该 yaml 文件后，KKCluster 的 webhook 会添加一些默认值，如 `roles: [controle-plane, worker]`。

## KKMachine

KKMachine 主要用于表示 Kubernetes node 节点和基础设施主机之间的映射。

```go
type KKMachineSpec struct {
	// 基础设施供应商的独有 ID
	ProviderID *string `json:"providerID,omitempty"`

	// 对应的 KKInstance 资源的名称
	InstanceID *string `json:"instanceID,omitempty"`

	// 该资源的角色
	// 可选
	Roles []Role `json:"roles"`

	// 该资源对应节点的容器运行时
	// 可选
	ContainerManager ContainerManager `json:"containerManager"`

	// 该资源对应节点的 Linux 软件包
	// 可选
	Repository *Repository `json:"repository,omitempty"`
}
```

通常我们不直接使用 KKMachine 资源，而是声明 KKMachineTemplate 资源供 cluster-api 的资源进行引用。

### KKMachineTemplate

所谓的 xxxTemplate 资源就是该类资源的模版类型，其包含了对应资源的 spec 中的所有内容。在 cluster-api 的使用过程中，我们会声明 KubeadmControlPlane（KCP）、MachineDeployment（MD）这样的带有 replica 字段资源，即说明了该类资源可进行动态伸缩。而这些资源中又包含了 infrastructureRef 字段，用于引用具体的基础设施模版资源（xxxMachineTemplate）。当发生扩容时，KCP、MD 这些资源就会根据模版资源创建出对应的，真正的，单个该资源实例。

以下为一个常见的控制平面 KKMachineTemplate 的 sample yaml 例子：
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: KKMachineTemplate
metadata:
  name: quick-start
  namespace: default
spec:
  template:
    spec:
      roles:
        - control-plane
      containerManager:
        type: containerd
      repository:
        iso: "none"
        update: false
```

以下为注意事项：
* roles 角色指定 control-plane，即引用该 KKMachineTemplate 的资源只会选择 KKCluster 中对应 role 的机器进行引导。
* containerManager 指定该模版创建的 KKMachine 配置容器运行时为 containerd。
* repository 指定该模版创建的 KKMachine 操作 rpm 软件源的策略，该事例中的策略表示 CAPKK 会使用对应 Linux 机器上的默认软件源安装默认需要软件依赖（ conntrack，socat 等）。

## KKInstance

在 cluster-api 的开发者文档中对于 infra 的定义仅包含 xxxCluster 和 xxxMachine 资源，而 KKInstance 是 CAPKK 独有的资源。CAPKK 额外定义 KKInstance 的目的是解耦 operator 和 controller 的逻辑，即 KKMachine 专注于维护于 cluster-api CR 交互，而 KKInstance 专注于维护 Linux 机器（通过 SSH 对机器进行命令式的操作）。

KKInstance 的属性字段主要就是 KKCluster 中 nodes 字段和 KKMachine 字段的集合，这里不做赘述。

KKInstance 的 controller 对于机器的操作归纳起来主要有以下几部分：
* Bootstrop：CAPKK 中定义的一些对机器的环境初始化或清理操作。如创建相应目录、添加 Linux 用户、运行初始化 shell 脚本、重置网络、删除文件、Kubeadm重置等操作。
* Repository：CAPKK 中定义的处理 rpm 软件源的操作。如挂载 ISO 软件包，更新源，安装依赖软件包等操作。
* Binary：CAPKK 中定义的处理集群组件二进制文件的操作。如下载二进制文件等。
* ContainerManager：CAPKK 中定义的处理机器容器运行时的操作。如检查容器运行时，安装容器运行时等操作。
* Provisioning：CAPKK 解析 cluster-api 提供的对应机器的 cloud-init 文件，并转换为 SSH 命令的操作。该 cloud-init 文件中就会包含诸如 “kubeadm init“、”kubeadm join“ 等操作。

对于这些操作的接口定义，见 [接口](https://github.com/kubesphere/kubekey/blob/master/pkg/service/interface.go)。
