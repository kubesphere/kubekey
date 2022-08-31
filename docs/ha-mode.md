# HA mode (internal loadbalancing)
K8s components require a loadbalancer to access the apiservers via a reverse proxy. Kubekey uses **kube-vip** and **haproxy** to provide internal ha mode. 
## haproxy
The way kubekey uses is referred to as localhost loadbalancing. The kubelet of each master node connects the local kube-apiserver, and the kubelet of each worker node connects the kube-apiserver via a local reverse proxy. Based on this, kubekey will deploy a haproxy-based proxy that resides on each worker node as the local reverse proxy.

![Image](img/haproxy.png?raw=true)

## kube-vip
The load balancing is provided through IPVS (IP Virtual Server) and provides a Layer 4 (TCP-based) round-robin across all of the control plane nodes. By default, the load balancer will listen on the default port of 6443 as the Kubernetes API server. The IPVS virtual server lives in kernel space and doesn't create an "actual" service that listens on port 6443. This allows the kernel to parse packets before they're sent to an actual TCP port. Based on this, kubekey will deploy a static pod that resides on each control-plane node as the internal loadbalancing.

![Image](img/kube-vip.png?raw=true)

## Usage
Modify your configuration file and uncomment the item `internalLoadbalancer`:
```yaml
controlPlaneEndpoint:
    internalLoadbalancer: haproxy #Internal loadbalancer for apiservers. Support: haproxy, kube-vip [Default: ""]
    
    domain: lb.kubesphere.local 
    address: "" # The IP address of your load balancer. If you use internalLoadblancer in "kube-vip" mode, a VIP is required here.
    port: 6443
```

Then whether you exec the command `create cluster`, `add nodes` or `upgrade`, kubekey will enable HA mode and deploy the interanl load balancer. 