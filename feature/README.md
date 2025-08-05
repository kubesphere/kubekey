# 背景
当前kubekey中，如果要添加命令，或修改命令，都需要提交代码并重新发版。扩展性较差。
1. 任务与框架分离（优势，目的，更方便扩展，借鉴ansible的playbook设计）
2. 支持gitops（可通过git方式，管理自动化任务）
3. 支持connector扩展
4. 支持云原生方式自动化批量任务管理

# 安装kubekey
## kubernetes中安装
```shell
helm upgrade --install --create-namespace -n kubekey-system kubekey kubekey-1.0.0.tgz
```
然后通过创建 `Inventory` 和 `Playbook` 资源来执行命令  
**Inventory**: 任务执行的host清单. 用于定义与host相关, 与任务模板无关的变量. 详见[参数定义](docs/zh/201-variable.md)  
**Playbook**: playbook的配置信息，在哪些host中执行，执行哪个playbook文件， 执行时参数等等。

## 二进制执行
可直接用二进制在命令行中执行命令
```shell
kk run -i inventory.yaml -c config.yaml playbook.yaml
```
运行命令后, 会在工作目录的runtime下生成对应的 `Inventory` 和 `Playbook` 资源

# 文档
**[项目模版编写规范](docs/zh/001-project.md)**  
**[模板语法](docs/zh/101-syntax.md)**  
**[参数定义](docs/zh/201-variable.md)**    
**[集群管理](docs/zh/core/README.md)**    

