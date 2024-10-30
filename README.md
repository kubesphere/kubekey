<div align=center><img src="docs/img/kubekey-logo.svg?raw=true"></div>

[![CI](https://github.com/kubesphere/kubekey/workflows/CI/badge.svg?branch=master&event=push)](https://github.com/kubesphere/kubekey/actions?query=event%3Apush+branch%3Amaster+workflow%3ACI+)

> English | [中文](README_zh-CN.md) | [日本語](README_ja-JP.md)

### 👋 Welcome to KubeKey!

KubeKey is an open-source lightweight tool for deploying Kubernetes clusters. It provides a flexible, rapid, and convenient way to install Kubernetes/K3s only, both Kubernetes/K3s and KubeSphere, and related cloud-native add-ons. It is also an efficient tool to scale and upgrade your cluster.

In addition, KubeKey also supports customized Air-Gap package, which is convenient for users to quickly deploy clusters in offline environments.

> KubeKey has passed [CNCF kubernetes conformance verification](https://www.cncf.io/certification/software-conformance/).

Use KubeKey in the following three scenarios.

* Install Kubernetes/K3s only
* Install Kubernetes/K3s and KubeSphere together in one command
* Install Kubernetes/K3s first, then deploy KubeSphere on it using [ks-installer](https://github.com/kubesphere/ks-installer)

> **Important:** If you have existing Kubernetes clusters, please refer to [ks-installer (Install KubeSphere on existing Kubernetes cluster)](https://github.com/kubesphere/ks-installer).

## Supported Environment

### Linux Distributions

* **Ubuntu**  *16.04, 18.04, 20.04, 22.04*
* **Debian**  *Bullseye, Buster, Stretch*
* **CentOS/RHEL**  *7*
* **AlmaLinux**  *9.0*
* **SUSE Linux Enterprise Server** *15*

> Recommended Linux Kernel Version: `4.15 or later` 
> You can run the `uname -srm` command to check the Linux Kernel Version.

### <span id = "KubernetesVersions">Kubernetes Versions</span>

* **v1.19**: &ensp; *v1.19.15*
* **v1.20**: &ensp; *v1.20.10*
* **v1.21**: &ensp; *v1.21.14*
* **v1.22**: &ensp; *v1.22.15*
* **v1.23**: &ensp; *v1.23.10*   (default)
* **v1.24**: &ensp; *v1.24.7*
* **v1.25**: &ensp; *v1.25.3*

> Looking for more supported versions: \
> [Kubernetes Versions](./docs/kubernetes-versions.md) \
> [K3s Versions](./docs/k3s-versions.md)

### Container Manager

* **Docker** / **containerd** / **CRI-O** / **iSula**

> `Kata Containers` can be set to automatically install and configure runtime class for it when the container manager is containerd or CRI-O.

### Network Plugins

* **Calico** / **Flannel** / **Cilium** / **Kube-OVN** / **Multus-CNI**

> Kubekey also supports users to set the network plugin to `none` if there is a requirement for custom network plugin.

## Requirements and Recommendations

* Minimum resource requirements (For Minimal Installation of KubeSphere only)：
  * 2 vCPUs
  * 4 GB RAM
  * 20 GB Storage

> /var/lib/docker is mainly used to store the container data, and will gradually increase in size during use and operation. In the case of a production environment, it is recommended that /var/lib/docker mounts a drive separately.

* OS requirements:
  * `SSH` can access to all nodes.
  * Time synchronization for all nodes.
  * `sudo`/`curl`/`openssl` should be used in all nodes.
  * `docker` can be installed by yourself or by KubeKey.
  * `Red Hat` includes `SELinux` in its `Linux release`. It is recommended to close SELinux or [switch the mode of SELinux](./docs/turn-off-SELinux.md) to `Permissive`

> * It's recommended that Your OS is clean (without any other software installed), otherwise there may be conflicts.
> * A container image mirror (accelerator) is recommended to be prepared if you have trouble downloading images from dockerhub.io. [Configure registry-mirrors for the Docker daemon](https://docs.docker.com/registry/recipes/mirror/#configure-the-docker-daemon).
> * KubeKey will install [OpenEBS](https://openebs.io/) to provision LocalPV for development and testing environment by default, this is convenient for new users. For production, please use NFS / Ceph / GlusterFS  or commercial products as persistent storage, and install the [relevant client](docs/storage-client.md) in all nodes.
> * If you encounter `Permission denied` when copying, it is recommended to check [SELinux and turn off it](./docs/turn-off-SELinux.md) first

* Dependency requirements:

KubeKey can install Kubernetes and KubeSphere together. Some dependencies need to be installed before installing kubernetes after version 1.18. You can refer to the list below to check and install the relevant dependencies on your node in advance.

|               | Kubernetes Version ≥ 1.18 |
| ------------- | -------------------------- |
| `socat`     | Required                   |
| `conntrack` | Required                   |
| `ebtables`  | Optional but recommended   |
| `ipset`     | Optional but recommended   |
| `ipvsadm`   | Optional but recommended   |

* Networking and DNS requirements:
  * Make sure the DNS address in `/etc/resolv.conf` is available. Otherwise, it may cause some issues of DNS in cluster.
  * If your network configuration uses Firewall or Security Group，you must ensure infrastructure components can communicate with each other through specific ports. It's recommended that you turn off the firewall or follow the link configuriation: [NetworkAccess](docs/network-access.md).

## Usage

### Get the KubeKey Executable File

* The fastest way to get KubeKey is to use the script:

  ```
  curl -sfL https://get-kk.kubesphere.io | sh -
  ```
* Binary downloads of the KubeKey also can be found on the [Releases page](https://github.com/kubesphere/kubekey/releases).
  Unpack the binary and you are good to go!
* Build Binary from Source Code

  ```shell
  git clone https://github.com/kubesphere/kubekey.git
  cd kubekey
  make kk
  ```

### Create a Cluster

#### Quick Start

Quick Start is for `all-in-one` installation which is a good start to get familiar with Kubernetes and KubeSphere.

> Note: Since Kubernetes temporarily does not support uppercase NodeName, contains uppercase letters in the hostname will lead to subsequent installation error

##### Command

> If you have problem to access `https://storage.googleapis.com`, execute first `export KKZONE=cn`.

```shell
./kk create cluster [--with-kubernetes version] [--with-kubesphere version]
```

##### Examples

* Create a pure Kubernetes cluster with default version (Kubernetes v1.23.10).

  ```shell
  ./kk create cluster
  ```
* Create a Kubernetes cluster with a specified version.

  ```shell
  ./kk create cluster --with-kubernetes v1.24.1 --container-manager containerd
  ```
* Create a Kubernetes cluster with KubeSphere installed.

  ```shell
  ./kk create cluster --with-kubesphere v3.2.1
  ```

#### Advanced

You have more control to customize parameters or create a multi-node cluster using the advanced installation. Specifically, create a cluster by specifying a configuration file.

> If you have problem to access `https://storage.googleapis.com`, execute first `export KKZONE=cn`.

1. First, create an example configuration file

   ```shell
   ./kk create config [--with-kubernetes version] [--with-kubesphere version] [(-f | --filename) path]
   ```

   **examples:**

   * create an example config file with default configurations. You also can specify the file that could be a different filename, or in different folder.

   ```shell
   ./kk create config [-f ~/myfolder/abc.yaml]
   ```

   * with KubeSphere

   ```shell
   ./kk create config --with-kubesphere v3.2.1
   ```
2. Modify the file config-sample.yaml according to your environment

> Note:  Since Kubernetes temporarily does not support uppercase NodeName, contains uppercase letters in workerNode`s name will lead to subsequent installation error
>
> A persistent storage is required in the cluster, when kubesphere will be installed. The local volume is used default. If you want to use other persistent storage, please refer to [addons](./docs/addons.md).

3. Create a cluster using the configuration file

   ```shell
   ./kk create cluster -f config-sample.yaml
   ```

### Enable Multi-cluster Management

By default, KubeKey will only install a **solo** cluster without Kubernetes federation. If you want to set up a multi-cluster control plane to centrally manage multiple clusters using KubeSphere, you need to set the `ClusterRole` in [config-example.yaml](docs/config-example.md). For multi-cluster user guide, please refer to [How to Enable the Multi-cluster Feature](https://github.com/kubesphere/community/tree/master/sig-multicluster/how-to-setup-multicluster-on-kubesphere).

### Enable Pluggable Components

KubeSphere has decoupled some core feature components since v2.1.0. These components are designed to be pluggable which means you can enable them either before or after installation. By default, KubeSphere will be started with a minimal installation if you do not enable them.

You can enable any of them according to your demands. It is highly recommended that you install these pluggable components to discover the full-stack features and capabilities provided by KubeSphere. Please ensure your machines have sufficient CPU and memory before enabling them. See [Enable Pluggable Components](https://github.com/kubesphere/ks-installer#enable-pluggable-components) for the details.

### Add Nodes

Add new node's information to the cluster config file, then apply the changes.

```shell
./kk add nodes -f config-sample.yaml
```

### Delete Nodes

You can delete the node by the following command，the nodeName that needs to be removed.

```shell
./kk delete node <nodeName> -f config-sample.yaml
```

### Delete Cluster

You can delete the cluster by the following command:

* If you started with the quick start (all-in-one):

```shell
./kk delete cluster
```

* If you started with the advanced (created with a configuration file):

```shell
./kk delete cluster [-f config-sample.yaml]
```

### Upgrade Cluster

#### Allinone

Upgrading cluster with a specified version.

```shell
./kk upgrade [--with-kubernetes version] [--with-kubesphere version] 
```

* Support upgrading Kubernetes only.
* Support upgrading KubeSphere only.
* Support upgrading Kubernetes and KubeSphere.

#### Multi-nodes

Upgrading cluster with a specified configuration file.

```shell
./kk upgrade [--with-kubernetes version] [--with-kubesphere version] [(-f | --filename) path]
```

* If `--with-kubernetes` or `--with-kubesphere` is specified, the configuration file will be also updated.
* Use `-f` to specify the configuration file which was generated for cluster creation.

> Note: Upgrading multi-nodes cluster need a specified configuration file. If the cluster was installed without kubekey or the configuration file for installation was not found, the configuration file needs to be created by yourself or following command.

Getting cluster info and generating kubekey's configuration file (optional).

```shell
./kk create config [--from-cluster] [(-f | --filename) path] [--kubeconfig path]
```

* `--from-cluster` means fetching cluster's information from an existing cluster.
* `-f` refers to the path where the configuration file is generated.
* `--kubeconfig` refers to the path where the kubeconfig.
* After generating the configuration file, some parameters need to be filled in, such as the ssh information of the nodes.

## Documents

* [Features List](docs/features.md)
* [Commands](docs/commands/kk.md)
* [Configuration example](docs/config-example.md)
* [Air-Gapped Installation](docs/manifest_and_artifact.md)
* [Highly Available clusters](docs/ha-mode.md)
* [Addons](docs/addons.md)
* [Network access](docs/network-access.md)
* [Storage clients](docs/storage-client.md)
* [kubectl auto-completion](docs/kubectl-autocompletion.md)
* [kubekey auto-completion](docs/kubekey-autocompletion.md)
* [Roadmap](docs/roadmap.md)
* [Check-Renew-Certificate](docs/check-renew-certificate.md)
* [Developer-Guide](docs/developer-guide.md)

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/pixiake"><img src="https://avatars0.githubusercontent.com/u/22290449?v=4?s=100" width="100px;" alt="pixiake"/><br /><sub><b>pixiake</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=pixiake" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=pixiake" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Forest-L"><img src="https://avatars2.githubusercontent.com/u/50984129?v=4?s=100" width="100px;" alt="Forest"/><br /><sub><b>Forest</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=Forest-L" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=Forest-L" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://kubesphere.io/"><img src="https://avatars2.githubusercontent.com/u/28859385?v=4?s=100" width="100px;" alt="rayzhou2017"/><br /><sub><b>rayzhou2017</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=rayzhou2017" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=rayzhou2017" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://www.chenshaowen.com/"><img src="https://avatars2.githubusercontent.com/u/43693241?v=4?s=100" width="100px;" alt="shaowenchen"/><br /><sub><b>shaowenchen</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=shaowenchen" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=shaowenchen" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://surenpi.com/"><img src="https://avatars1.githubusercontent.com/u/1450685?v=4?s=100" width="100px;" alt="Zhao Xiaojie"/><br /><sub><b>Zhao Xiaojie</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=LinuxSuRen" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=LinuxSuRen" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/zackzhangkai"><img src="https://avatars1.githubusercontent.com/u/20178386?v=4?s=100" width="100px;" alt="Zack Zhang"/><br /><sub><b>Zack Zhang</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=zackzhangkai" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://akhilerm.com/"><img src="https://avatars1.githubusercontent.com/u/7610845?v=4?s=100" width="100px;" alt="Akhil Mohan"/><br /><sub><b>Akhil Mohan</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=akhilerm" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/FeynmanZhou"><img src="https://avatars3.githubusercontent.com/u/40452856?v=4?s=100" width="100px;" alt="pengfei"/><br /><sub><b>pengfei</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=FeynmanZhou" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/min-zh"><img src="https://avatars1.githubusercontent.com/u/35321102?v=4?s=100" width="100px;" alt="min zhang"/><br /><sub><b>min zhang</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=min-zh" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=min-zh" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/zgldh"><img src="https://avatars1.githubusercontent.com/u/312404?v=4?s=100" width="100px;" alt="zgldh"/><br /><sub><b>zgldh</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=zgldh" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/xrjk"><img src="https://avatars0.githubusercontent.com/u/16330256?v=4?s=100" width="100px;" alt="xrjk"/><br /><sub><b>xrjk</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=xrjk" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/stoneshi-yunify"><img src="https://avatars2.githubusercontent.com/u/70880165?v=4?s=100" width="100px;" alt="yonghongshi"/><br /><sub><b>yonghongshi</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=stoneshi-yunify" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/shenhonglei"><img src="https://avatars2.githubusercontent.com/u/20896372?v=4?s=100" width="100px;" alt="Honglei"/><br /><sub><b>Honglei</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=shenhonglei" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/liucy1983"><img src="https://avatars2.githubusercontent.com/u/2360302?v=4?s=100" width="100px;" alt="liucy1983"/><br /><sub><b>liucy1983</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=liucy1983" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/lilien1010"><img src="https://avatars1.githubusercontent.com/u/3814966?v=4?s=100" width="100px;" alt="Lien"/><br /><sub><b>Lien</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=lilien1010" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/klj890"><img src="https://avatars3.githubusercontent.com/u/19380605?v=4?s=100" width="100px;" alt="Tony Wang"/><br /><sub><b>Tony Wang</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=klj890" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/hlwanghl"><img src="https://avatars3.githubusercontent.com/u/4861515?v=4?s=100" width="100px;" alt="Hongliang Wang"/><br /><sub><b>Hongliang Wang</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=hlwanghl" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://fafucoder.github.io/"><img src="https://avatars0.githubusercontent.com/u/16442491?v=4?s=100" width="100px;" alt="dawn"/><br /><sub><b>dawn</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=fafucoder" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/duanjiong"><img src="https://avatars1.githubusercontent.com/u/3678855?v=4?s=100" width="100px;" alt="Duan Jiong"/><br /><sub><b>Duan Jiong</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=duanjiong" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/calvinyv"><img src="https://avatars3.githubusercontent.com/u/28883416?v=4?s=100" width="100px;" alt="calvinyv"/><br /><sub><b>calvinyv</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=calvinyv" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/benjaminhuo"><img src="https://avatars2.githubusercontent.com/u/18525465?v=4?s=100" width="100px;" alt="Benjamin Huo"/><br /><sub><b>Benjamin Huo</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=benjaminhuo" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Sherlock113"><img src="https://avatars2.githubusercontent.com/u/65327072?v=4?s=100" width="100px;" alt="Sherlock113"/><br /><sub><b>Sherlock113</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=Sherlock113" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Fuchange"><img src="https://avatars1.githubusercontent.com/u/31716848?v=4?s=100" width="100px;" alt="fu_changjie"/><br /><sub><b>fu_changjie</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=Fuchange" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/yuswift"><img src="https://avatars1.githubusercontent.com/u/37265389?v=4?s=100" width="100px;" alt="yuswift"/><br /><sub><b>yuswift</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=yuswift" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ruiyaoOps"><img src="https://avatars.githubusercontent.com/u/35256376?v=4?s=100" width="100px;" alt="ruiyaoOps"/><br /><sub><b>ruiyaoOps</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=ruiyaoOps" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://www.luxingmin.com"><img src="https://avatars.githubusercontent.com/u/1918195?v=4?s=100" width="100px;" alt="LXM"/><br /><sub><b>LXM</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=lxm" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/sbhnet"><img src="https://avatars.githubusercontent.com/u/2368131?v=4?s=100" width="100px;" alt="sbhnet"/><br /><sub><b>sbhnet</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=sbhnet" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/misteruly"><img src="https://avatars.githubusercontent.com/u/31399968?v=4?s=100" width="100px;" alt="misteruly"/><br /><sub><b>misteruly</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=misteruly" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://johnniang.me"><img src="https://avatars.githubusercontent.com/u/16865714?v=4?s=100" width="100px;" alt="John Niang"/><br /><sub><b>John Niang</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=JohnNiang" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://alimy.me"><img src="https://avatars.githubusercontent.com/u/10525842?v=4?s=100" width="100px;" alt="Michael Li"/><br /><sub><b>Michael Li</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=alimy" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/duguhaotian"><img src="https://avatars.githubusercontent.com/u/3174621?v=4?s=100" width="100px;" alt="独孤昊天"/><br /><sub><b>独孤昊天</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=duguhaotian" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/lshmouse"><img src="https://avatars.githubusercontent.com/u/118687?v=4?s=100" width="100px;" alt="Liu Shaohui"/><br /><sub><b>Liu Shaohui</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=lshmouse" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/24sama"><img src="https://avatars.githubusercontent.com/u/43993589?v=4?s=100" width="100px;" alt="Leo Li"/><br /><sub><b>Leo Li</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=24sama" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/RolandMa1986"><img src="https://avatars.githubusercontent.com/u/1720333?v=4?s=100" width="100px;" alt="Roland"/><br /><sub><b>Roland</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=RolandMa1986" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://ops.m114.org"><img src="https://avatars.githubusercontent.com/u/2347587?v=4?s=100" width="100px;" alt="Vinson Zou"/><br /><sub><b>Vinson Zou</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=vinsonzou" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/tagGeeY"><img src="https://avatars.githubusercontent.com/u/35259969?v=4?s=100" width="100px;" alt="tag_gee_y"/><br /><sub><b>tag_gee_y</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=tagGeeY" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/liulangwa"><img src="https://avatars.githubusercontent.com/u/25916792?v=4?s=100" width="100px;" alt="codebee"/><br /><sub><b>codebee</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=liulangwa" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/TheApeMachine"><img src="https://avatars.githubusercontent.com/u/9572060?v=4?s=100" width="100px;" alt="Daniel Owen van Dommelen"/><br /><sub><b>Daniel Owen van Dommelen</b></sub></a><br /><a href="#ideas-TheApeMachine" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Naidile-P-N"><img src="https://avatars.githubusercontent.com/u/29476402?v=4?s=100" width="100px;" alt="Naidile P N"/><br /><sub><b>Naidile P N</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=Naidile-P-N" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/haiker2011"><img src="https://avatars.githubusercontent.com/u/8073429?v=4?s=100" width="100px;" alt="Haiker Sun"/><br /><sub><b>Haiker Sun</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=haiker2011" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/yj-cloud"><img src="https://avatars.githubusercontent.com/u/19648473?v=4?s=100" width="100px;" alt="Jing Yu"/><br /><sub><b>Jing Yu</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=yj-cloud" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/chaunceyjiang"><img src="https://avatars.githubusercontent.com/u/17962021?v=4?s=100" width="100px;" alt="Chauncey"/><br /><sub><b>Chauncey</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=chaunceyjiang" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/tanguofu"><img src="https://avatars.githubusercontent.com/u/87045830?v=4?s=100" width="100px;" alt="Tan Guofu"/><br /><sub><b>Tan Guofu</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=tanguofu" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/lvillis"><img src="https://avatars.githubusercontent.com/u/56720445?v=4?s=100" width="100px;" alt="lvillis"/><br /><sub><b>lvillis</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=lvillis" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/vincenthe11"><img src="https://avatars.githubusercontent.com/u/8400716?v=4?s=100" width="100px;" alt="Vincent He"/><br /><sub><b>Vincent He</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=vincenthe11" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://laminar.fun/"><img src="https://avatars.githubusercontent.com/u/2360535?v=4?s=100" width="100px;" alt="laminar"/><br /><sub><b>laminar</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=tpiperatgod" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/cumirror"><img src="https://avatars.githubusercontent.com/u/2455429?v=4?s=100" width="100px;" alt="tongjin"/><br /><sub><b>tongjin</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=cumirror" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://k8s.li"><img src="https://avatars.githubusercontent.com/u/42566386?v=4?s=100" width="100px;" alt="Reimu"/><br /><sub><b>Reimu</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=muzi502" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://bandism.net/"><img src="https://avatars.githubusercontent.com/u/22633385?v=4?s=100" width="100px;" alt="Ikko Ashimine"/><br /><sub><b>Ikko Ashimine</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=eltociear" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://yeya24.github.io/"><img src="https://avatars.githubusercontent.com/u/25150124?v=4?s=100" width="100px;" alt="Ben Ye"/><br /><sub><b>Ben Ye</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=yeya24" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/yinheli"><img src="https://avatars.githubusercontent.com/u/235094?v=4?s=100" width="100px;" alt="yinheli"/><br /><sub><b>yinheli</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=yinheli" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/hellocn9"><img src="https://avatars.githubusercontent.com/u/102210430?v=4?s=100" width="100px;" alt="hellocn9"/><br /><sub><b>hellocn9</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=hellocn9" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/brandan-schmitz"><img src="https://avatars.githubusercontent.com/u/6267549?v=4?s=100" width="100px;" alt="Brandan Schmitz"/><br /><sub><b>Brandan Schmitz</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=brandan-schmitz" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/yjqg6666"><img src="https://avatars.githubusercontent.com/u/1879641?v=4?s=100" width="100px;" alt="yjqg6666"/><br /><sub><b>yjqg6666</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=yjqg6666" title="Documentation">📖</a> <a href="https://github.com/kubesphere/kubekey/commits?author=yjqg6666" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/zaunist"><img src="https://avatars.githubusercontent.com/u/38528079?v=4?s=100" width="100px;" alt="失眠是真滴难受"/><br /><sub><b>失眠是真滴难受</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=zaunist" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/mangoGoForward"><img src="https://avatars.githubusercontent.com/u/35127166?v=4?s=100" width="100px;" alt="mango"/><br /><sub><b>mango</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/pulls?q=is%3Apr+reviewed-by%3AmangoGoForward" title="Reviewed Pull Requests">👀</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/wenwutang1"><img src="https://avatars.githubusercontent.com/u/45817987?v=4?s=100" width="100px;" alt="wenwutang"/><br /><sub><b>wenwutang</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=wenwutang1" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://kuops.com"><img src="https://avatars.githubusercontent.com/u/18283256?v=4?s=100" width="100px;" alt="Shiny Hou"/><br /><sub><b>Shiny Hou</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=kuops" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/zhouqiu0103"><img src="https://avatars.githubusercontent.com/u/108912268?v=4?s=100" width="100px;" alt="zhouqiu0103"/><br /><sub><b>zhouqiu0103</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=zhouqiu0103" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/77yu77"><img src="https://avatars.githubusercontent.com/u/73932296?v=4?s=100" width="100px;" alt="77yu77"/><br /><sub><b>77yu77</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=77yu77" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/hzhhong"><img src="https://avatars.githubusercontent.com/u/83079531?v=4?s=100" width="100px;" alt="hzhhong"/><br /><sub><b>hzhhong</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=hzhhong" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/arugal"><img src="https://avatars.githubusercontent.com/u/26432832?v=4?s=100" width="100px;" alt="zhang-wei"/><br /><sub><b>zhang-wei</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=arugal" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://twitter.com/xds2000"><img src="https://avatars.githubusercontent.com/u/37678?v=4?s=100" width="100px;" alt="Deshi Xiao"/><br /><sub><b>Deshi Xiao</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=xiaods" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=xiaods" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://besscroft.com"><img src="https://avatars.githubusercontent.com/u/33775809?v=4?s=100" width="100px;" alt="besscroft"/><br /><sub><b>besscroft</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=besscroft" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/zhangzhiqiangcs"><img src="https://avatars.githubusercontent.com/u/8319897?v=4?s=100" width="100px;" alt="张志强"/><br /><sub><b>张志强</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=zhangzhiqiangcs" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/lwabish"><img src="https://avatars.githubusercontent.com/u/7044019?v=4?s=100" width="100px;" alt="lwabish"/><br /><sub><b>lwabish</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=lwabish" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=lwabish" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/qyz87"><img src="https://avatars.githubusercontent.com/u/36068894?v=4?s=100" width="100px;" alt="qyz87"/><br /><sub><b>qyz87</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=qyz87" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/fangzhengjin"><img src="https://avatars.githubusercontent.com/u/12680972?v=4?s=100" width="100px;" alt="ZhengJin Fang"/><br /><sub><b>ZhengJin Fang</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=fangzhengjin" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://lhr.wiki"><img src="https://avatars.githubusercontent.com/u/6327311?v=4?s=100" width="100px;" alt="Eric_Lian"/><br /><sub><b>Eric_Lian</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=ExerciseBook" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/nicognaW"><img src="https://avatars.githubusercontent.com/u/66731869?v=4?s=100" width="100px;" alt="nicognaw"/><br /><sub><b>nicognaw</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=nicognaW" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/deqingLv"><img src="https://avatars.githubusercontent.com/u/6064297?v=4?s=100" width="100px;" alt="吕德庆"/><br /><sub><b>吕德庆</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=deqingLv" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/littleplus"><img src="https://avatars.githubusercontent.com/u/11694750?v=4?s=100" width="100px;" alt="littleplus"/><br /><sub><b>littleplus</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=littleplus" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://www.linkedin.com/in/%D0%BA%D0%BE%D0%BD%D1%81%D1%82%D0%B0%D0%BD%D1%82%D0%B8%D0%BD-%D0%B0%D0%BA%D0%B0%D0%BA%D0%B8%D0%B5%D0%B2-13130b1b4/"><img src="https://avatars.githubusercontent.com/u/82488489?v=4?s=100" width="100px;" alt="Konstantin"/><br /><sub><b>Konstantin</b></sub></a><br /><a href="#ideas-Nello-Angelo" title="Ideas, Planning, & Feedback">🤔</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://kiragoo.github.io"><img src="https://avatars.githubusercontent.com/u/7400711?v=4?s=100" width="100px;" alt="kiragoo"/><br /><sub><b>kiragoo</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=kiragoo" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/jojotong"><img src="https://avatars.githubusercontent.com/u/100849526?v=4?s=100" width="100px;" alt="jojotong"/><br /><sub><b>jojotong</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=jojotong" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/littleBlackHouse"><img src="https://avatars.githubusercontent.com/u/54946465?v=4?s=100" width="100px;" alt="littleBlackHouse"/><br /><sub><b>littleBlackHouse</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=littleBlackHouse" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=littleBlackHouse" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/testwill"><img src="https://avatars.githubusercontent.com/u/8717479?v=4?s=100" width="100px;" alt="guangwu"/><br /><sub><b>guangwu</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=testwill" title="Code">💻</a> <a href="https://github.com/kubesphere/kubekey/commits?author=testwill" title="Documentation">📖</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/wongearl"><img src="https://avatars.githubusercontent.com/u/36498442?v=4?s=100" width="100px;" alt="wongearl"/><br /><sub><b>wongearl</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=wongearl" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/wenwenxiong"><img src="https://avatars.githubusercontent.com/u/10548812?v=4?s=100" width="100px;" alt="wenwenxiong"/><br /><sub><b>wenwenxiong</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=wenwenxiong" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://baimeow.cn/"><img src="https://avatars.githubusercontent.com/u/38121125?v=4?s=100" width="100px;" alt="柏喵Sakura"/><br /><sub><b>柏喵Sakura</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=BaiMeow" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://dashen.tech"><img src="https://avatars.githubusercontent.com/u/15921519?v=4?s=100" width="100px;" alt="cui fliter"/><br /><sub><b>cui fliter</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=cuishuang" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/liuxu623"><img src="https://avatars.githubusercontent.com/u/9653438?v=4?s=100" width="100px;" alt="刘旭"/><br /><sub><b>刘旭</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=liuxu623" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/yzxiu"><img src="https://avatars.githubusercontent.com/u/13790023?v=4?s=100" width="100px;" alt="yuyu"/><br /><sub><b>yuyu</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=yzxiu" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/chilianyi"><img src="https://avatars.githubusercontent.com/u/5917832?v=4?s=100" width="100px;" alt="chilianyi"/><br /><sub><b>chilianyi</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=chilianyi" title="Code">💻</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/xrwang8"><img src="https://avatars.githubusercontent.com/u/68765051?v=4?s=100" width="100px;" alt="Ronald Fletcher"/><br /><sub><b>Ronald Fletcher</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=xrwang8" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/baikjy0215"><img src="https://avatars.githubusercontent.com/u/110450904?v=4?s=100" width="100px;" alt="baikjy0215"/><br /><sub><b>baikjy0215</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=baikjy0215" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/knowmost"><img src="https://avatars.githubusercontent.com/u/167442703?v=4?s=100" width="100px;" alt="knowmost"/><br /><sub><b>knowmost</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=knowmost" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Hiiirad"><img src="https://avatars.githubusercontent.com/u/26739670?v=4?s=100" width="100px;" alt="Hirad Rasoolinejad"/><br /><sub><b>Hirad Rasoolinejad</b></sub></a><br /><a href="https://github.com/kubesphere/kubekey/commits?author=Hiiirad" title="Code">💻</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
