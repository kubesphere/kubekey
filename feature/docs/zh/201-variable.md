# 变量
变量分为静态变量(运行前定义的变量)和动态变量(运行时生成的变量)  
参数优先级为: 动态变量 > 静态变量
## 静态变量
静态变量包含节点清单, 全局配置, 模板中定义的参数.
参数优先级为: 全局配置 > 节点清单 > 模版中定义的参数.
### 节点清单
yaml格式文件, 不包含模板语法, 通过`-i`参数传入(`kk -i inventory.yaml ...`), 在每个host上生效  
**定义规范**:
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    - hostname1: 
        k1: v1
        #...
    - hostname2: 
        k2: v2
        #...
    - hostname3:
      #...
  groups:
    groupname1:
      groups:
        - groupname2
        # ...
      hosts:
        - hostname1
        #...
      vars:
        k1: v1
        #...
    groupname2:
    #...
  vars:
    k1: v1
    #...
```
**hosts**: key为host名称, value为给该host设置的变量.  
**groups**: 给host进行分组. key为组名称, value有groups, hosts和vars
- groups: 该组包含哪些其他组.
- hosts:  该组包含哪些hosts.
- vars: 组级别的变量, 针对组中所有host生效.  
groups包含的总hosts为`groups`包含的host + `hosts`中包含的host.  
**vars**: 全局变量, 针对所有host生效.  
变量优先级为: $(host_variable) > $(group_variable) > $(global_variable)
### 全局配置
yaml格式文件, 不包含模板语法, 通过`-c`参数传入(`kk -c config.yaml ...`), 在每个host上生效
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
metadata:
  name: default
spec:
  k: v
  #...
```
任意类型的参数
### 模板中定义的参数
模板中定义的var参数包含: 
- playbook中`vars`字段和`vars_files`字段定义的参数
- role中的defaults/main.yaml定义的参数
- role中`vars`字段定义的参数
- task中`vars`字段定义的参数
## 动态变量
动态变量是节点执行是生成的变量数据, 包含:  
- `gather_facts`定义的参数  
- `register`定义的参数  
- `set_fact`定义的参数  
优先级为参数定义的顺序, 后定义参数高于先定义的参数.  
