# 注入 Playbook（inject_playbooks.yaml）

![architecture](../../images/architecture.png)

`inject_playbooks.yaml` 是 KubeKey 内置的一个**注入用示例 playbook**。它本身默认是空操作，
用于在不改动内置 playbook 与 role 代码的前提下，在默认参数加载完成后覆盖 / 修改参数。

该文件随内置包发布，位于 `builtin/core/playbooks/hook/inject_playbooks.yaml`。可直接通过
[`playbooks` 注入机制](../framework/002-playbook.md#注入自定义-playbookinject-playbooks) 引用：

```yaml
spec:
  playbooks:
    - order: 1.5                        # 插在第 1、2 个原始 play 之间
      path: hook/inject_playbooks.yaml # 本文件，已随内置包提供
```

## 你能做什么

- 用 `set_fact` 覆盖当前主机的运行时变量（针对单个 host play）。
- 用 `add_hostvars` 批量给多台主机写入 / 覆盖变量（类似 `set_fact` 但作用于多主机）。
- 用 `debug` 打印变量，方便排查注入是否正确生效。

## 完整示例

文件内容默认全部以注释形式给出，取消对应注释即可生效：

```yaml
- name: Inject | Inject and override playbook variables
  hosts:
    - all
  tasks:
    # 示例 1：覆盖单一变量（针对当前 play 的所有主机生效）
    # - name: Inject | Override a single variable via set_fact
    #   set_fact:
    #     .kubernetes.version: "v1.31.0"

    # 示例 2：按分组 / 主机条件覆盖变量
    # - name: Inject | Override variables for a specific group
    #   set_fact:
    #     .kubernetes.container_manager: "containerd"
    #   when:
    #     - .groups.k8s_cluster | default list | has .inventory_hostname

    # 示例 3：批量给多台主机注入变量（add_hostvars）
    # - name: Inject | Add host variables to multiple hosts
    #   add_hostvars:
    #     hosts: ["all"]
    #     vars:
    #       custom_label: "injected-by-hook"

    # 示例 4：调试 —— 确认注入前的变量取值
    # - name: Inject | Debug current variables
    #   debug:
    #     msg: "kubernetes version is {{ .kubernetes.version }}"
```

## 注意事项

- 注入的变量仅在「当前 playbook 运行期间」有效（运行时变量），不会写回你的 config 文件。
  如需持久化定制，应当优先在 config 的 spec 中声明。
- 仅当确有「默认参数加载后仍需二次计算 / 条件覆盖」的需求时才启用示例。
- **内置 playbook** 可直接引用随包提供的 `hook/inject_playbooks.yaml`；
  **本地项目**（非内置 playbook）请在自己的项目目录下创建该文件，并用相对路径引用。
- 关于 `order` 定位、排序规则、空路径跳过等行为，详见
  [注入自定义 Playbook](../framework/002-playbook.md#注入自定义-playbookinject-playbooks)。
