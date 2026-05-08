# 角色（Role）

Role 是一组 [task](004-task.md) 的集合。

## 在 Playbook 中引用

```yaml
- name: Playbook Name
  # ...
  roles:
    - name: Role Name
      tags: ["always"]
      when: true
      run_once: false
      ignore_errors: false
      vars: { a: b }
      role: Role-ref Name
```

| 字段 | 说明 |
|------|------|
| **name** | role 显示名称，可选。与 `role` 引用名可不同。 |
| **tags** | 标签，可选。仅作用于该 role 引用。 |
| **when** | 执行条件，可选。可为字符串或数组，对每个 host 分别求值。 |
| **run_once** | 是否只执行一次，可选，默认 `false`。在第一个 host 上执行。 |
| **ignore_errors** | 该 role 下 task 失败时是否忽略，可选，默认 `false`。 |
| **role** | 引用名，必填。对应 `roles/` 下子目录名。 |
| **vars** | 默认变量，可选，YAML 格式。 |

## Role 目录结构

```text
project/roles/roleName/
├── defaults/
│   └── main.yml    # 默认变量，对该 role 下所有 task 生效
├── tasks/
│   └── main.yml    # [task](004-task.md) 定义
├── templates/      # 模板文件，供 template 类 task 使用
│   └── template1
└── files/          # 静态文件，供 copy 类 task 使用
    └── file1
```

- **roleName**：即 playbook 中 `role` 的引用名，可多级目录（如 `a/b`）。
- **defaults**：在 `main.yml` 中定义该 role 的默认参数。
- **tasks**：在 `main.yml` 中定义该 role 的 [task](004-task.md) 列表。
- **templates**：模板文件，通常含 [模板语法](101-syntax.md) 变量引用。
- **files**：原始文件，在 [copy](modules/copy.md) task 中通过相对路径引用。
