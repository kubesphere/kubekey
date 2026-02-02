# 任务（Task）

Task 分为**单层**与**多层**两种：

- **单层 task**：包含 module 相关字段，不包含 `block`。一个 task 仅能使用一个 module。
- **多层 task**：不包含 module 字段，包含 `block`（及可选的 `rescue`、`always`）。

Task 会在 play 的每个 host 上分别执行（除非 `run_once: true`）。

## 文件定义

```yaml
- include_tasks: other/task.yaml
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: { a: b }

- name: Block Name
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: { a: b }
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

| 字段 | 说明 |
|------|------|
| **include_tasks** | 引用其他 task 文件。 |
| **name** | task 名称，可选。 |
| **tags** | 标签，可选。仅作用于该 task，不继承 play / role 的 tags。 |
| **when** | 执行条件，可选。可为字符串或数组，使用 [模板语法](101-syntax.md)，对每个 host 分别求值。 |
| **failed_when** | 失败条件，可选。满足时视为失败，支持 [模板语法](101-syntax.md)。 |
| **run_once** | 是否只执行一次，可选，默认 `false`。在第一个 host 上执行。 |
| **ignore_errors** | 是否忽略失败，可选，默认 `false`。 |
| **vars** | 该 task 的变量，可选，YAML 格式。 |
| **loop** | 循环执行 module，每次迭代以 `item` 传递当前值。可为字符串或数组，使用 [模板语法](101-syntax.md)。 |
| **retries** | 失败时重试次数，可选。 |
| **register** | 将执行结果写入 [变量](201-variable.md)，供后续 task 使用。含 `stderr`、`stdout` 等子字段。 |
| **register_type** | `register` 的解析格式：`string`（默认）、`json`、`yaml`。 |
| **block** | task 列表。未定义 module 时必填，正常流程执行。 |
| **rescue** | task 列表。`block` 中任一同级 task 失败时执行。 |
| **always** | task 列表。`block`（及若有 `rescue`）执行完后无论成败都会执行。 |
| **module** | 具体操作，与 [已注册模块](README.md#模块) 对应。未使用 `block` 时必填。 |

## 已注册模块

| 模块 | 说明 |
|------|------|
| [add_hostvars](modules/add_hostvars.md) | 向指定主机注入变量 |
| [assert](modules/assert.md) | 条件断言 |
| [command](modules/command.md) | 执行命令 |
| [copy](modules/copy.md) | 复制文件/目录到目标主机 |
| [debug](modules/debug.md) | 打印变量 |
| [fetch](modules/fetch.md) | 从远程主机拉取文件到本地 |
| [gen_cert](modules/gen_cert.md) | 校验或生成证书 |
| [image](modules/image.md) | 拉取/推送/复制镜像 |
| [include_vars](modules/include_vars.md) | 从 YAML 文件加载变量 |
| [prometheus](modules/prometheus.md) | 查询 Prometheus 指标 |
| [result](modules/result.md) | 写入 playbook status detail |
| [set_fact](modules/set_fact.md) | 在当前主机设置变量 |
| [setup](modules/setup.md) | 采集主机信息 |
| [template](modules/template.md) | 渲染模板并复制到目标主机 |
