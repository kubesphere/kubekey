# 流程（Playbook）

## 文件定义

一个 playbook 文件中可按定义顺序执行多个 play；每个 play 指定在哪些 host 上执行哪些任务。

```yaml
- import_playbook: others/playbook.yaml

- name: Playbook Name
  tags: ["always"]
  hosts: ["host1", "host2"]
  serial: 1
  run_once: false
  ignore_errors: false
  gather_facts: false
  vars: { a: b }
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

| 字段 | 说明 |
|------|------|
| **import_playbook** | 引用的 playbook 路径（通常为相对路径）。查找顺序：`项目路径/playbooks/` → `当前路径/playbooks/` → `当前路径/`。 |
| **name** | play 名称，可选。 |
| **tags** | play 的标签，可选。仅作用于该 play，不会继承到其下 role / task。执行时可通过 `--tags` / `--skip-tags` 筛选。`always` 始终执行，`never` 始终不执行；`all` 表示所有 play，`tagged` 表示带标签的 play。 |
| **hosts** | 执行目标，必填。可为 host 名或 group 名，均需在 [inventory](201-variable.md#节点清单) 中定义（localhost 除外）。 |
| **serial** | 分批执行。可为单个值（数字或字符串）或数组。默认一批执行。若为数组，按固定数量对 `hosts` 分组；超出时按最后一个值扩展。如 `[1, 2]`、`hosts: [a,b,c,d]` → 第一批 `[a]`，第二批 `[b,c]`，第三批 `[d]`。支持百分比（如 `[30%, 60%]`），可与数字混用。 |
| **run_once** | 是否只执行一次，可选，默认 `false`。为 `true` 时在第一个 host 上执行。 |
| **ignore_errors** | 该 play 下 task 失败时是否忽略，可选，默认 `false`。 |
| **gather_facts** | 是否采集主机信息，可选，默认 `false`。按 connector 类型采集不同数据（如 `local` / `ssh`：`release`、`kernel_version`、`hostname`、`architecture`，仅 Linux）。 |
| **vars** | 默认变量，可选，YAML 格式。 |
| **vars_files** | 从 YAML 文件加载默认变量，可选。与 `vars` 的 key 不可重复。 |
| **pre_tasks** | 前置 [tasks](004-task.md)，可选。 |
| **roles** | 要执行的 [roles](003-role.md)，可选。 |
| **tasks** | 主 [tasks](004-task.md)，可选。 |
| **post_tasks** | 后置 [tasks](004-task.md)，可选。 |

## 执行顺序

- **多个 play**：按定义顺序执行；`import_playbook` 会先展开为对应 play。
- **同一 play 内**：`pre_tasks` → `roles` → `tasks` → `post_tasks`。
- 任一 task 失败（且未 `ignore_errors`）则 play 失败。

## 注入自定义 Playbook（Inject Playbooks）

除在 playbook 文件内写死 `import_playbook` 外，还可以通过 playbook 的 config spec 声明
`playbooks` 列表，把**自定义 playbook 注入到顶层（第一个）源 playbook 的任意位置**。
这样无需改动内置 playbook 与 role 代码，就能在默认参数加载完成后覆盖 / 修改参数。

### 配置

```yaml
spec:
  playbooks:
    - order: 1.5                        # 插在第 1、2 个原始 play 之间
      path: hook/inject_playbooks.yaml # 要注入的 playbook（相对项目根，支持模板渲染）
    - order: 0                          # 0 表示插在最前面（第 1 个 play 之前）
      path: hook/pre_custom.yaml
```

### 规则

- **原始 play 自动获得权重 `1, 2, 3, ...`**（按文档顺序，1-based）。
- **注入项给显式 `order`**（float，可小数、可负数），用于定位插入位置；例如
  `order: 1.5` 表示插在第 1 个与第 2 个原始 play 之间，`order: 0` 表示插在最前。
- **排序比较器（确定性裁决，order 相同不报错）**：
  1. `order` 升序；
  2. `playbooks` 配置项优先于文件内 play；
  3. 同来源内按定义顺序（配置列表顺序 / 文件内顺序）。
- **`path` 走模板渲染**：
  - 渲染为空串 / 未设置变量 → 该项被**跳过**（不报错）；
  - 设置了路径但文件不存在 → 按原逻辑**报错**，便于发现路径拼写错误。
- 该机制**仅作用于顶层（第一个）源 playbook.yaml**；被导入的子 playbook 不再二次注入。

### 示例

内置包已提供一个示例 playbook `hook/inject_playbooks.yaml`（随内置包发布，可直接用
`path: hook/inject_playbooks.yaml` 引用）。它默认是空操作，取消注释即可覆盖变量。
完整示例与字段说明见参考页：[注入 Playbook（inject_playbooks.yaml）](../reference/playbooks/inject_playbooks.md)。

