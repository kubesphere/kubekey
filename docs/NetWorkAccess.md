Network Access
------------
If your network configuration uses an firewall，you must ensure infrastructure components can communicate with each other through specific ports that act as communication endpoints for certain processes or services.

|services|protocol|action|start port|end port|comment
|---|---|---|---|---|---|
ssh|TCP|allow|22|
etcd|TCP|allow|2379|2380|
apiserver|TCP|allow|6443|
calico|TCP|allow|9099|9100|
bgp|TCP|allow|179||
nodeport|TCP|allow|30000|32767|
master|TCP|allow|10250|10258|
dns|TCP|allow|53|
dns|UDP|allow|53|
local-registry|TCP|allow|5000||offline environment|
local-apt|TCP|allow|5080||offline environment|
rpcbind|TCP|allow|111|| use NFS
ipip|IPENCAP / IPIP|allow| | |calico needs to allow the ipip protocol