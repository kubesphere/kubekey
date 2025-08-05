# 任务
task分为单层级task,和多层级task
单层级task: 包含 module 相关字段, 不包含block字段. 一个task只能包含一个module.   
多层级task: 不包含 module 相关字段, 包含block字段.  
task执行时, 会在定义的host分别上执行.  
## 文件定义
```yaml
- include_tasks: other/task.yaml
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: {a: b}
  
- name: Block Name
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: {a: b}
  block:
    - name: Task Name
      [module]
  rescue:
    - name: Task Name
      [module]
  always:
    - name: Task Name
      [module]
  
- name: Task Name
  tags: ["always"]
  when: true
  loop: [""]
  [module]
```
**include_tasks**: 该任务中引用其他任务模板文件.  
**name**: task名称, 非必填.   
**tags**: task的标签, 非必填. 仅作用于playbook, playbook下的role, task不会继承该标签.  
**when**: 执行条件, 可以定义单个值(字符串)或多个值(数组), 非必填, 默认执行该role. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**failed_when**: 失败条件, host满足该条件时,判定为执行失败, 可以定义单个值(字符串)或多个值(数组), 非必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**run_once**: 是否只执行一次, 非必填, 默认false, 会在第一个hosts上执行.   
**ignore_errors**: 是否忽略失败, 非必填, 默认false.   
**vars**: 配置默认参数, 非必填, yaml格式.  
**loop**: 循环执行module中定义的操作, 每次执行时,以`item: loop-value`的形式将值传递给module. 可以定义单个值(字符串)或多个值(数组), 非必填, 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**retries**: task执行失败时. 需要重新尝试几次.  
**register**: 值为字符串, 将任务执行结果注册到[variable](201-variable.md)中, 传递给后续的task. register包含两个子字段： 
- stderr: 失败输出
- stdout: 成功输出
**register_type**: 注册register中stderr和stdout的格式。
- string: 默认，以字符串的格式注册。
- json: 以json的格式注册。
- yaml: 以yaml的格式注册。
**block**: task集合, 非必填(当未定义module相关字段时, 必填), 一定会执行.  
**rescue**: task集合, 非必填, 当block执行失败(task集合有一个执行失败即为该block失败)时,执行该task集合.   
**always**: task集合, 非必填, 当block和rescue执行完毕后(无论成功失败)都会执行该task集合.  
**module**: task实际要执行的操作, 非必填(当未block字段时, 必填).map格式的数据, key为module_name, value为args. 可用的module需提前在项目中进行注册。已注册的module如下
- [add_hostvars](modules/add_hostvars.md)
- [assert](modules/assert.md)
- [command](modules/command.md)
- [copy](modules/copy.md)
- [debug](modules/debug.md)
- [fetch](modules/fetch.md)
- [gen_cert](modules/gen_cert.md)
- [image](modules/image.md)
- [prometheus](modules/prometheus.md)
- [result](modules/result.md)
- [set_fact](modules/set_fact.md)
- [setup](modules/setup.md)
- [template](modules/template.md)