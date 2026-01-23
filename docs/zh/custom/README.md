# 自定义项目开发

本文档说明如何基于 KubeKey 编写与运行自定义 playbook 项目。KubeKey 在任务编排上参考了 [Ansible](https://github.com/ansible/ansible) 的规范，便于快速理解与上手。

## 文档导航

### 基础概念

| 文档 | 说明 |
|------|------|
| [项目 (001-project)](001-project.md) | 项目结构、playbooks / roles 目录、内建 / 本地 / Git 存放方式 |
| [流程 (002-playbook)](002-playbook.md) | Playbook 定义、hosts / tags / serial、pre_tasks / roles / tasks / post_tasks |
| [角色 (003-role)](003-role.md) | Role 结构、defaults / tasks / templates / files、在 playbook 中引用 |
| [任务 (004-task)](004-task.md) | Task 定义、单层/多层 task、block / rescue / always、loop / register |

### 语法与变量

| 文档 | 说明 |
|------|------|
| [模板语法 (101-syntax)](101-syntax.md) | Go template 与 Sprig、toYaml / fromYaml、ipInCIDR、自定义函数 |
| [变量 (201-variable)](201-variable.md) | 静态变量（inventory、全局配置、模板参数）与动态变量（gather_facts、register、set_fact） |

### 模块

任务中可用的 module 需在项目中注册，以下为已注册模块：

| 模块 | 说明 |
|------|------|
| [add_hostvars](modules/add_hostvars.md) | 向指定主机注入变量 |
| [assert](modules/assert.md) | 条件断言 |
| [command](modules/command.md) | 执行命令（shell / kubectl 等） |
| [copy](modules/copy.md) | 复制文件或目录到目标主机 |
| [debug](modules/debug.md) | 打印变量 |
| [fetch](modules/fetch.md) | 从远程主机拉取文件到本地 |
| [gen_cert](modules/gen_cert.md) | 校验或生成证书 |
| [image](modules/image.md) | 拉取 / 推送 / 复制镜像 |
| [include_vars](modules/include_vars.md) | 从 YAML 文件加载变量 |
| [prometheus](modules/prometheus.md) | 查询 Prometheus 指标 |
| [result](modules/result.md) | 写入 playbook status detail |
| [set_fact](modules/set_fact.md) | 在当前主机设置变量 |
| [setup](modules/setup.md) | 获取主机信息（gather_facts 底层） |
| [template](modules/template.md) | 渲染模板并复制到目标主机 |

## 快速开始

1. 阅读 [项目](001-project.md) 了解目录结构与存放方式。
2. 阅读 [流程](002-playbook.md) 与 [任务](004-task.md) 编写第一个 playbook 与 task。
3. 使用 [模板语法](101-syntax.md) 与 [变量](201-variable.md) 引用和传递参数。
4. 按需查阅各 [模块](modules/) 文档，选用合适的 module 实现具体逻辑。
