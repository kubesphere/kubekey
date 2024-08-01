# KubeKey

[![CI](https://github.com/kubesphere/kubekey/workflows/CI/badge.svg?branch=master&event=push)](https://github.com/kubesphere/kubekey/actions?query=event%3Apush+branch%3Amaster+workflow%3ACI+)

> [English](README.md) | [中文](README_zh-CN.md) | 日本語

KubeKeyは、Kubernetesクラスターをデプロイするためのオープンソースの軽量ツールです。Kubernetes/K3s、Kubernetes/K3sとKubeSphere、および関連するクラウドネイティブアドオンをインストールするための柔軟で迅速かつ便利な方法を提供します。また、クラスターのスケールとアップグレードを効率的に行うツールでもあります。

さらに、KubeKeyはカスタマイズされたエアギャップパッケージもサポートしており、オフライン環境でクラスターを迅速にデプロイすることができます。

> KubeKeyは[CNCF kubernetes 一致性認証](https://www.cncf.io/certification/software-conformance/)に合格しています。

KubeKeyを使用する3つのシナリオがあります。

* Kubernetes/K3sのみをインストールする
* Kubernetes/K3sとKubeSphereを1つのコマンドでインストールする
* まずKubernetes/K3sをインストールし、その上に[ks-installer](https://github.com/kubesphere/ks-installer)を使用してKubeSphereをデプロイする

> **重要:** 既存のKubernetesクラスターがある場合は、[ks-installer (既存のKubernetesクラスターにKubeSphereをインストールする)](https://github.com/kubesphere/ks-installer)を参照してください。

## サポートされている環境

### Linuxディストリビューション

* **Ubuntu**  *16.04, 18.04, 20.04, 22.04*
* **Debian**  *Bullseye, Buster, Stretch*
* **CentOS/RHEL**  *7*
* **AlmaLinux**  *9.0*
* **SUSE Linux Enterprise Server** *15*

> 推奨Linuxカーネルバージョン: `4.15以降` 
> Linuxカーネルバージョンを確認するには、`uname -srm`コマンドを実行します。

### <span id = "KubernetesVersions">Kubernetesバージョン</span>

* **v1.19**: &ensp; *v1.19.15*
* **v1.20**: &ensp; *v1.20.10*
* **v1.21**: &ensp; *v1.21.14*
* **v1.22**: &ensp; *v1.22.15*
* **v1.23**: &ensp; *v1.23.10*   (デフォルト)
* **v1.24**: &ensp; *v1.24.7*
* **v1.25**: &ensp; *v1.25.3*

> サポートされているバージョンの詳細については、以下を参照してください: \
> [Kubernetesバージョン](./docs/kubernetes-versions.md) \
> [K3sバージョン](./docs/k3s-versions.md)

### コンテナマネージャー

* **Docker** / **containerd** / **CRI-O** / **iSula**

> `Kata Containers`は、コンテナマネージャーがcontainerdまたはCRI-Oの場合に、ランタイムクラスを自動的にインストールおよび構成するように設定できます。

### ネットワークプラグイン

* **Calico** / **Flannel** / **Cilium** / **Kube-OVN** / **Multus-CNI**

> カスタムネットワークプラグインの要件がある場合、Kubekeyはネットワークプラグインを`none`に設定することもサポートしています。

## 要件と推奨事項

* 最小リソース要件（KubeSphereの最小インストールの場合）：
  * 2 vCPU
  * 4 GB RAM
  * 20 GBストレージ

> /var/lib/dockerは主にコンテナデータを保存するために使用され、使用および操作中に徐々にサイズが大きくなります。プロダクション環境の場合、/var/lib/dockerが別のドライブにマウントされることをお勧めします。

* OS要件:
  * `SSH`がすべてのノードにアクセスできること。
  * すべてのノードの時間が同期されていること。
  * すべてのノードで`sudo`/`curl`/`openssl`が使用できること。
  * `docker`は自分でインストールするか、KubeKeyを使用してインストールできます。
  * `Red Hat`の`Linuxリリース`には`SELinux`が含まれています。SELinuxを無効にするか、[SELinuxのモードを切り替える](./docs/turn-off-SELinux.md)ことをお勧めします。

> * OSがクリーンであることをお勧めします（他のソフトウェアがインストールされていない）。そうしないと、競合が発生する可能性があります。
> * dockerhub.ioからのイメージのダウンロードに問題がある場合は、コンテナイメージミラー（アクセラレータ）を準備することをお勧めします。[Dockerデーモンのレジストリミラーを構成する](https://docs.docker.com/registry/recipes/mirror/#configure-the-docker-daemon)。
> * KubeKeyはデフォルトで[OpenEBS](https://openebs.io/)をインストールして、開発およびテスト環境のLocalPVをプロビジョニングします。これは新しいユーザーにとって非常に便利です。プロダクション環境では、NFS / Ceph / GlusterFSまたは商用製品を永続ストレージとして使用し、すべてのノードに[関連クライアント](docs/storage-client.md)をインストールしてください。
> * コピー時に`Permission denied`が発生した場合は、まず[SELinuxを無効にする](./docs/turn-off-SELinux.md)ことをお勧めします。

* 依存関係の要件:

KubeKeyはKubernetesとKubeSphereを一緒にインストールできます。バージョン1.18以降、kubernetesのインストール前にいくつかの依存関係をインストールする必要があります。以下のリストを参照して、事前にノードに関連する依存関係をチェックしてインストールしてください。

|               | Kubernetesバージョン ≥ 1.18 |
| ------------- | -------------------------- |
| `socat`     | 必須                   |
| `conntrack` | 必須                   |
| `ebtables`  | オプションですが推奨   |
| `ipset`     | オプションですが推奨   |
| `ipvsadm`   | オプションですが推奨   |

* ネットワーキングとDNSの要件:
  * `/etc/resolv.conf`のDNSアドレスが利用可能であることを確認してください。そうしないと、クラスター内のDNSに関する問題が発生する可能性があります。
  * ネットワーク構成でファイアウォールまたはセキュリティグループを使用している場合、インフラストラクチャコンポーネントが特定のポートを介して相互に通信できることを確認する必要があります。ファイアウォールをオフにするか、リンク構成に従うことをお勧めします: [ネットワークアクセス](docs/network-access.md)。

## 使用方法

### KubeKey実行ファイルの取得

* 最速の方法はスクリプトを使用することです:

  ```
  curl -sfL https://get-kk.kubesphere.io | sh -
  ```
* KubeKeyのバイナリダウンロードは[リリースページ](https://github.com/kubesphere/kubekey/releases)にもあります。
  バイナリを解凍して、すぐに使用できます！
* ソースコードからバイナリをビルドする

  ```shell
  git clone https://github.com/kubesphere/kubekey.git
  cd kubekey
  make kk
  ```

### クラスターの作成

#### クイックスタート

クイックスタートは`all-in-one`インストール用で、KubernetesとKubeSphereに慣れるための良いスタートです。

> 注意: Kubernetesは一時的に大文字のNodeNameをサポートしていないため、ホスト名に大文字が含まれていると、後続のインストールエラーが発生します

##### コマンド

> `https://storage.googleapis.com`にアクセスする際に問題がある場合は、最初に`export KKZONE=cn`を実行してください。

```shell
./kk create cluster [--with-kubernetes version] [--with-kubesphere version]
```

##### 例

* デフォルトバージョン（Kubernetes v1.23.10）で純粋なKubernetesクラスターを作成する。

  ```shell
  ./kk create cluster
  ```
* 指定されたバージョンのKubernetesクラスターを作成する。

  ```shell
  ./kk create cluster --with-kubernetes v1.24.1 --container-manager containerd
  ```
* KubeSphereがインストールされたKubernetesクラスターを作成する。

  ```shell
  ./kk create cluster --with-kubesphere v3.2.1
  ```

#### 高度なインストール

高度なインストールを使用すると、カスタムパラメーターを制御したり、マルチノードクラスターを作成したりすることができます。具体的には、構成ファイルを指定してクラスターを作成します。

> `https://storage.googleapis.com`にアクセスする際に問題がある場合は、最初に`export KKZONE=cn`を実行してください。

1. まず、サンプル構成ファイルを作成します

   ```shell
   ./kk create config [--with-kubernetes version] [--with-kubesphere version] [(-f | --filename) path]
   ```

   **例:**

   * デフォルトの構成でサンプル構成ファイルを作成します。ファイル名やフォルダーを指定することもできます。

   ```shell
   ./kk create config [-f ~/myfolder/abc.yaml]
   ```

   * KubeSphereを含む

   ```shell
   ./kk create config --with-kubesphere v3.2.1
   ```
2. 環境に応じてconfig-sample.yamlファイルを変更します

> 注意: Kubernetesは一時的に大文字のNodeNameをサポートしていないため、workerNodeの名前に大文字が含まれていると、後続のインストールエラーが発生します
>
> KubeSphereをインストールする場合、クラスターには永続ストレージが必要です。デフォルトではローカルボリュームが使用されます。他の永続ストレージを使用する場合は、[addons](./docs/addons.md)を参照してください。

3. 構成ファイルを使用してクラスターを作成します

   ```shell
   ./kk create cluster -f config-sample.yaml
   ```

### マルチクラスター管理の有効化

デフォルトでは、KubeKeyはKubernetesフェデレーションなしで**ソロ**クラスターのみをインストールします。KubeSphereを使用して複数のクラスターを集中管理するためのマルチクラスターコントロールプレーンを設定する場合は、[config-example.yaml](docs/config-example.md)で`ClusterRole`を設定する必要があります。マルチクラスターのユーザーガイドについては、[マルチクラスター機能の有効化方法](https://github.com/kubesphere/community/tree/master/sig-multicluster/how-to-setup-multicluster-on-kubesphere)を参照してください。

### プラグイン可能なコンポーネントの有効化

KubeSphereはv2.1.0以降、いくつかのコア機能コンポーネントをデカップリングしました。これらのコンポーネントはプラグイン可能であり、インストール前またはインストール後に有効にすることができます。デフォルトでは、これらのコンポーネントを有効にしない場合、KubeSphereは最小インストールで開始されます。

ニーズに応じて、これらのコンポーネントを有効にすることができます。KubeSphereが提供するフルスタックの機能と機能を発見するために、これらのプラグイン可能なコンポーネントをインストールすることを強くお勧めします。これらを有効にする前に、マシンに十分なCPUとメモリがあることを確認してください。詳細については、[プラグイン可能なコンポーネントの有効化](https://github.com/kubesphere/ks-installer#enable-pluggable-components)を参照してください。

### ノードの追加

新しいノードの情報をクラスター構成ファイルに追加し、変更を適用します。

```shell
./kk add nodes -f config-sample.yaml
```

### ノードの削除

次のコマンドでノードを削除できます。削除するノード名を指定します。

```shell
./kk delete node <nodeName> -f config-sample.yaml
```

### クラスターの削除

次のコマンドでクラスターを削除できます:

* クイックスタート（all-in-one）を開始した場合:

```shell
./kk delete cluster
```

* 高度なインストール（構成ファイルを使用して作成されたクラスター）を開始した場合:

```shell
./kk delete cluster [-f config-sample.yaml]
```

### クラスターのアップグレード

#### Allinone

指定されたバージョンでクラスターをアップグレードします。

```shell
./kk upgrade [--with-kubernetes version] [--with-kubesphere version] 
```

* Kubernetesのみのアップグレードをサポートします。
* KubeSphereのみのアップグレードをサポートします。
* KubernetesとKubeSphereの両方のアップグレードをサポートします。

#### マルチノード

指定された構成ファイルでクラスターをアップグレードします。

```shell
./kk upgrade [--with-kubernetes version] [--with-kubesphere version] [(-f | --filename) path]
```

* `--with-kubernetes`または`--with-kubesphere`が指定されている場合、構成ファイルも更新されます。
* クラスター作成に使用された構成ファイルを指定するには、`-f`を使用します。

> 注意: マルチノードクラスターのアップグレードには、指定された構成ファイルが必要です。クラスターがkubekeyなしでインストールされた場合、またはインストールに使用された構成ファイルが見つからない場合は、構成ファイルを自分で作成するか、次のコマンドを使用して作成する必要があります。

クラスター情報を取得し、kubekeyの構成ファイルを生成します（オプション）。

```shell
./kk create config [--from-cluster] [(-f | --filename) path] [--kubeconfig path]
```

* `--from-cluster`は、既存のクラスターからクラスター情報を取得することを意味します。
* `-f`は、構成ファイルが生成されるパスを指します。
* `--kubeconfig`は、kubeconfigのパスを指します。
* 構成ファイルを生成した後、ノードのssh情報などのいくつかのパラメーターを入力する必要があります。

## ドキュメント

* [機能リスト](docs/features.md)
* [コマンド](docs/commands/kk.md)
* [構成例](docs/config-example.md)
* [エアギャップインストール](docs/manifest_and_artifact.md)
* [高可用性クラスター](docs/ha-mode.md)
* [アドオン](docs/addons.md)
* [ネットワークアクセス](docs/network-access.md)
* [ストレージクライアント](docs/storage-client.md)
* [kubectl自動補完](docs/kubectl-autocompletion.md)
* [kubekey自動補完](docs/kubekey-autocompletion.md)
* [ロードマップ](docs/roadmap.md)
* [証明書の確認と更新](docs/check-renew-certificate.md)
* [開発者ガイド](docs/developer-guide.md)

## 貢献者 ✨

これらの素晴らしい人々に感謝します（[絵文字キー](https://allcontributors.org/docs/en/emoji-key)）：

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
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

このプロジェクトは [all-contributors](https://github.com/all-contributors/all-contributors) の仕様に従っています。どのような種類の貢献でも歓迎します！
