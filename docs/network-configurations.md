### Network Configurations

#### IPVS

If your cluster's kubeProxy mode is `ipvs` which is default value in `kk`, kubernetes will add some ipvs records on each node. You can use `ipvsadm` command to get more information.

#### Iptables

If your cluster's kubeProxy mode is `iptables`, kubernetes will add some iptables records on each node. You can use `iptables` command to get more information.

#### Virtual Device

Most of CNI Plugins will create some virtual devices on each node, such as `cni0`. You can use `ip link` command to inspect them in details.

As for `flannel`, virtual devices named with `flannel` prefix will be created.
As for `calico`, virtual devices named in `cali[a-f0-9]*` regexp format will be created.
As for `cilium`, virtual devices named with `cilium_` prefix will be created.

If your cluster's kubeProxy mode is `ipvs`, additional virtual device `kube-ipvs0` will be created.

If your cluster enables `nodelocaldns` feature for DNS caching purpose, additional virtual device `nodelocaldns` will be created.

#### Network Namespace

CNI plugins may create some network namespaces named with `cni-` prefix depends on which CNI plugin you choose to use. You can use `ip netns show 2>/dev/null | grep cni-` command to get CNI network namespace list.
