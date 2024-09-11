# 角色
角色是一个任务组
## 在playbook文件定义role引用
```yaml
- name: Playbook Name
  #...
  roles:
    - name: Role Name
      tags: ["always"]
      when: true
      run_once: false
      ignore_errors: false
      vars: {a: b}
      role: Role-ref Name
```
**name**: role名称, 非必填. 该名称不同于playbook中role的引用名称.  
**tags**: playbook的标签, 非必填. 仅作用于playbook, playbook下的role, task不会继承该标签.  
**when**: 执行条件, 可以定义单个值(字符串)或多个值(数组), 非必填, 默认执行该role. 对每个的host单独计算值.  
**run_once**: 是否只执行一次, 非必填, 默认false, 会在第一个hosts上执行.  
**ignore_errors**: 该role下所关联的task执行失败时, 是否忽略失败, 非必填, 默认false.  
**role**: playbook中引用的名称, 对应roles目录下的子目录, 必填.  
**vars**: 配置默认参数, 非必填, yaml格式.  
## 在role目录结构
```text
|-- project
|   |-- roles/
|   |   |-- roleName/  
|   |   |   |-- defaults/  
|   |   |   |   |-- main.yml  
|   |   |   |-- tasks/  
|   |   |   |   |-- main.yml  
|   |   |   |-- templates/  
|   |   |   |   |-- template1  
|   |   |   |-- files/  
|   |   |   |   |-- file1    
```
**roleName**：role的引用名称, 一级或多级目录.   
**defaults**：对role下的所有task, 定义默认参数值. 在main.yaml文件中定义.   
**[tasks](004-task.md)**：role下所关联的task模板, 一个角色可以有多个task, 在main.yaml文件中定义.    
**templates**：模板文件, 文件中通常会引用变量, 在`templates`类型的task中使用  
**files**：原始文件, 在`copy`类型的task中使用  
