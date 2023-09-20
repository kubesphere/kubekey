## 一、Harbor 简介

Harbor 是由 VMware 公司使用 Go 语言开发，主要就是用于存放镜像使用，同时我们还可以通过 Web 界面来对存放的镜像进行管理。并且 Harbor 提供的功能有：基于角色的访问控制，镜像远程复制同步，以及审计日志等功能。官方文档

### 1.Harbor 功能介绍

1）基于角色的访问控制： 我们可以通过项目来对用户进行权限划分，项目中可以包含多个镜像。

2）审计管理： 我们可以在用户审计管理中，找到我们所有对镜像仓库做的操作。

3）镜像复制： 我们可以通过配置，使在两台 Harbor 服务器间实现镜像同步。

4）漏洞扫描： Harbor 会定期对仓库中的镜像进行扫描，并进行策略检查，以防止部署出易受到攻击的镜像。

### 2.Harbor 高可用方式

目前 Harbor 最常见的高可用方式有两种，分别是：

1）安装两台 Harbor 仓库，他们共同使用一个存储（一般常见的便是 NFS 共享存储）

![Image](img/harbor_ha01.png?raw=true)

2）安装两台 Harbor 仓库，并互相配置同步关系。

![image](img/harbor_ha02.png?raw=true)

因为第一种方式的话，需要额外配置 Redis 和 PostgreSQL 以及 NFS 服务，所以我们下面使用第二种方式进行 Harbor 高可用配置。

# Harbor镜像仓库高可用方案设计

采用2台harbor仓库互为主备的方案，如下图所示

![image](img/harbor_keepalived.png?raw=true)

注意：主备方案

由于VIP的浮动，主备节点其实是互为主备；在部署harbor时，需要注意主备节点上的harbor.yml中hostname不要配置为浮动IP（reg.harbor.4a或192.168.10.200），应配置为各自IP或者hostname；
早先，将VIP的域名reg.harbor.4a配置到190和191上的harbor.yml中（hostname: reg.harbor.4a）导致一个问题：只有主节点可做为target被添加，用作镜像同步（也就是无法在主节点的仓库管理中创建备节点的target，即便添加了也无法连通）。

准备文件(在源码的script文件夹下,keepalived镜像可以从互联网阿里云镜像仓库下载)如下

```
# tree .
.
├── harborCreateRegistriesAndReplications.sh
├── keepalived21.tar
├── kk
└── harbor_keepalived
    ├── check_harbor.sh
    ├── docker-compose-keepalived-backup.yaml
    ├── docker-compose-keepalived-master.yaml
    ├── keepalived-backup.conf
    └── keepalived-master.conf

1 directory, 8 files
```

kk: kubekey支持多节点harbor仓库代码(包含本pr)编译生成二进制文件

harborCreateRegistriesAndReplications.sh：配置harbor互为主备的脚本

keepalived21.tar：keepalived的docker镜像

harbor_keepalived：keepalived master和slave的docker-compose部署文件

## kubekey部署多节点harbor仓库

通过二次开发kubekey源码，实现了kubekey部署harbor仓库支持多节点，并且配置同一套harbor证书。证书中包含所有部署harbor节点的主机名和IP认证设置。

后续集成到一键部署脚本中，通过配置registry角色的多个节点来部署多harbor仓库。推荐2个harbor仓库，部署过多占用资源。

## harbor仓库互为主备设置

harbor仓库部署后，通过调用harbor仓库api建立备份仓库，建立备份规则。

例如master1节点上仓库和master2节点仓库配置如下

```
#!/bin/bash

Harbor_master1_Address=master1:7443
master1_Address=192.168.122.61
Harbor_master2_Address=master2:7443
master2_Address=192.168.122.62

Harbor_User=admin                                  #登录Harbor的用户
Harbor_Passwd="Harbor12345"                  #登录Harbor的用户密码
Harbor_UserPwd="$Harbor_User:$Harbor_Passwd"

# create registry
curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/registries" -d "{\"name\": \"master1_2_master2\", \"type\": \"harbor\", \"url\":\"https://${master2_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"
# create registry
curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/registries" -d "{\"name\": \"master2_2_master1\", \"type\": \"harbor\", \"url\":\"https://${master1_Address}:7443\", \"credential\": {\"access_key\": \"${Harbor_User}\", \"access_secret\": \"${Harbor_Passwd}\"}, \"insecure\": true}"

#createReplication
curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master1_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master1_2_master2\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 1, \"name\": \"master1_2_master2\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"

#createReplication
curl -k -u $Harbor_UserPwd  -X POST -H "Content-Type: application/json" "https://${Harbor_master2_Address}/api/v2.0/replication/policies" -d "{\"name\": \"master2_2_master1\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 1, \"name\": \"master2_2_master1\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
```

## keepalived管理harbor服务VIP

使用docker-compose管理keepalived服务

keepalived master服务器配置如下

```
# cat docker-compose-keepalived-master.yaml
version: '3.8'

# Docker-Compose 单容器使用参考 YAML 配置文件
# 更多配置参数请参考镜像 README.md 文档中说明
services:
  keepalived:
    image: 'dockerhub.kubekey.local/kubesphere/keepalived:2.1'
    privileged: true
    network_mode: host
    volumes:
      - ./keepalived-master.conf:/srv/conf/keepalived/keepalived.conf
      - ./check_harbor.sh:/srv/conf/keepalived/check_harbor.sh
    container_name: keepalived
    restart: on-failure
    
# cat keepalived-master.conf
vrrp_script check_harbor {
        script "/srv/conf/keepalived/check_harbor.sh"
        interval 10   # 间隔时间，单位为秒，默认1秒
        fall 2        # 脚本几次失败转换为失败
        rise 2        # 脚本连续监测成功后，把服务器从失败标记为成功的次数
        timeout 5
        init_fail
}
global_defs {
        script_user root
        router_id harbor-ha
        enable_script_security
        lvs_sync_daemon ens3 VI_1
}
vrrp_instance VI_1 {
        state  MASTER
        interface ens3
        virtual_router_id 31    # 如果同一个局域网中有多套keepalive，那么要保证该id唯一
        priority 100
        advert_int 1
        authentication {
                auth_type PASS
                auth_pass k8s-test
        }
        virtual_ipaddress {
                192.168.122.59
        }
        track_script {
                check_harbor
        }
}
# cat check_harbor.sh
#!/bin/bash
#count=$(docker-compose -f /opt/harbor/docker-compose.yml ps -a|grep healthy|wc -l)
# 不能频繁调用docker-compose 否则会有非常多的临时目录被创建：/tmp/_MEI*
count=$(docker ps |grep goharbor|grep healthy|wc -l)
status=$(ss -tlnp|grep -w 443|wc -l)
if [ $count -ne 11 -a  ];then
   exit 8
elif [ $status -lt 2 ];then
   exit 9
else
   exit 0
fi
```

keepalived slave服务器跟master区别配置如下

1、state  BACKUP 与 MASTER

2、priority master配置为100，slave设置为50

```
# cat docker-compose-keepalived-backup.yaml
version: '3.8'

# Docker-Compose 单容器使用参考 YAML 配置文件
# 更多配置参数请参考镜像 README.md 文档中说明
services:
  keepalived:
    image: 'dockerhub.kubekey.local/kubesphere/keepalived:2.1'
    privileged: true
    network_mode: host
    volumes:
      - ./keepalived-backup.conf:/srv/conf/keepalived/keepalived.conf
      - ./check_harbor.sh:/srv/conf/keepalived/check_harbor.sh
    container_name: keepalived
    restart: on-failure
    
# cat keepalived-backup.conf
vrrp_script check_harbor {
        script "/srv/conf/keepalived/check_harbor.sh"
        interval 10   # 间隔时间，单位为秒，默认1秒
        fall 2        # 脚本几次失败转换为失败
        rise 2        # 脚本连续监测成功后，把服务器从失败标记为成功的次数
        timeout 5
        init_fail
}
global_defs {
        script_user root
        router_id harbor-ha
        enable_script_security
        lvs_sync_daemon ens3 VI_1
}
vrrp_instance VI_1 {
        state  BACKUP
        interface ens3
        virtual_router_id 31    # 如果同一个局域网中有多套keepalive，那么要保证该id唯一
        priority 50
        advert_int 1
        authentication {
                auth_type PASS
                auth_pass k8s-test
        }
        virtual_ipaddress {
                192.168.122.59
        }
        track_script {
                check_harbor
        }
}
```

经常需要变动的参数是设置keepalived的interface和vip地址值，实际环境下可以参数化keepalived这2个值。