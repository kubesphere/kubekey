# 流程
## 文件定义
一个playbook文件中, 按定义顺序执行多个playbook, 每个playbook指定在哪些host上执行哪些任务. 
```yaml
- import_playbook: others/playbook.yaml

- name: Playbook Name
  tags: ["always"]
  hosts: ["host1", "host2"]
  serial: 1
  run_once: false
  ignore_errors: false
  gather_facts: false
  vars: {a: b}
  vars_files: ["vars/variables.yaml"]
  pre_tasks:
    - name: Task Name
      debug:
        msg: "I'm Task"
  roles:
    - role: role1
      when: true
  tasks:
    - name: Task Name
      debug:
        msg: "I'm Task"
  post_tasks:
    - name: Task Name
      debug:
        msg: "I'm Task"
```
**import_playbooks**: 定义引用的playbook文件名称, 通常为相对路径, 文件查找顺序为：`项目路径/playbooks/`,  `当前路径/playbooks/`,  `当前路径/`  
**name**: playbook名称, 非必填.   
**tags**: playbook的标签, 非必填.仅作用于playbook, playbook下的role, task不会继承该标签.  
在执行playbook命令时, 通过参数筛选需要执行的playbook. 示例如下:  
- `kk run [playbook] --tags tag1 --tags tag2`: 执行带有tag1标签或带有tag2标签的playbook  
- `kk run [playbook] --skip-tags tag1 --skip-tags tag2`: 执行时跳过带有tag1标签或带有tag2标签的playbook  
其中, 带有`always`标签的playbook始终执行, 带有`never`标签的playbook始终不执行  
传入参数为`all`时, 表示选择所有playbook, 参数参数为`tagged`时, 表示选择打了标签的playbook
**hosts**: 定义在哪些机器上执行, 必填. 所有hosts需要在`inventory`中定义(localhost除外). 可以填host名称, 也可以填group名称.   
**serial**: 分批次执行playbook, 可以定义单个值(字符串或数字)或一组值(数组), 非必填. 默认一批执行。
- serial值为一组数字时, 按固定的数量来给`hosts`分组, 超出`serial`定义范围时, 按最后一个`serial`值扩展. 
  比如serial的值为[1, 2], hosts的值为[a, b, c, d]时. 会分3批来执行playbook, 第一批在[a]上执行, 第二批在[b, c]上执行, 第三批 在[d]上执行. 
- serial值为百分比时, 按百分比计算出每批次实际的`hosts`数量(下行整数), 然后给`hosts`分组, 超出`serial`定义范围时, 按最后一个`serial`值扩展. 
  比如serial的值为[30%, 60%], hosts的值为[a, b, c, d]时. 先计算出serial为[1.2,  2.4], 即为[1, 2]. 
百分比和数字可以混合设置.  
**run_once**: 是否只执行一次, 非必填, 默认false, 会在第一个hosts上执行.   
**ignore_errors**: 该playbook下所关联的task执行失败时, 是否忽略失败, 非必填, 默认false.   
**gather_facts**: 是否获取服务器信息, 非必填, 默认false. 针对不同的host获取不同的数据.   
- localConnector: 获取release(/etc/os-release),  kernel_version(uname -r),  hostname(hostname),  architecture(arch). 目前仅支持linux系统  
- sshConnector: 获取release(/etc/os-release),  kernel_version(uname -r),  hostname(hostname),  architecture(arch). 目前仅支持linux系统  
- kubernetesConnector：暂无  
**vars**: 配置默认参数, 非必填, yaml格式.  
**vars_files**: 配置默认参数, 非必填, yaml文件格式. vars和vars_files定义的字段不能重复.  
**pre_tasks**: 定义需要执行的[tasks](004-task.md), 非必填.  
**roles**: 定义需要执行的[roles](003-role.md), 非必填.  
**tasks**: 定义需要执行的[tasks](004-task.md), 非必填.  
**post_tasks**: 定义需要执行的[tasks](004-task.md), 非必填.  
## playbook执行顺序
不同的playbook: 按定义的先后顺序执行. 如果包含了import_playbook, 会将引用的playbook文件, 转成playbook.   
同一个playbook中: 任务执行顺序pre_tasks->roles->tasks->post_tasks  
当其中一个task失败时(不包含ignore状态), playbook执行失败.  
