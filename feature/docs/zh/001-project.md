# 项目
项目中存放要执行的任务模板. 由一系列的yaml文件构成  
为了便于使用者快速理解和上手，kk在对任务抽象时，参考借鉴了[ansible](https://github.com/ansible/ansible)任务编排规范
## 目录结构
```text
|-- project
|   |-- playbooks/  
|   |-- playbook1.yaml  
|   |-- playbook2.yaml  
|   |-- roles/
|   |   |-- roleName1/    
|   |   |-- roleName2/    
...
```
**[playbooks](002-playbook.md)**：执行入口, 存放一系列playbook. 一个playbook中, 可定义多个task或role. 每次执行流程模板时, 会按定义顺序执行对应的任务.   
**[roles](003-role.md)**：role集合. 一个role是一组task.
## 存放路径
项目可存放内建, 本地或git服务器上. 
### 内建
内建项目在`builtin`目录. 会集成到kubekey的命令中. 
执行示例：
```shell
kk precheck
```
执行`builtin`目录中的`playbooks/precheck.yaml`流程文件. 
### 本地
执行命令示例：
```shell
kk run demo.yaml
```
执行当前目录的`demo.yaml`流程文件. 
### git
执行命令示例：
```shell
kk  run playbooks/demo.yaml
  --project-addr=$(GIT_URL) \
  --project-branch=$(GIT_BRANCH)
```
执行git地址为`$(GIT_URL)`, 分支为`$(GIT_BRANCH)`上的`playbooks/demo.yaml`流程文件. 
