# 项目

项目中存放要执行的任务模板，由一系列 YAML 文件构成。为便于理解与上手，KubeKey 在任务编排上参考了 [Ansible](https://github.com/ansible/ansible) 的规范。

## 目录结构

```text
project/
├── playbooks/          # 可选：存放 playbook 文件
│   ├── playbook1.yaml
│   └── playbook2.yaml
├── playbook1.yaml      # 或直接在项目根目录放置 playbook
├── playbook2.yaml
└── roles/
    ├── roleName1/
    └── roleName2/
```

- **[playbooks](002-playbook.md)**：执行入口，存放 playbook。一个 playbook 可定义多个 task 或 role，执行时按定义顺序依次运行。
- **[roles](003-role.md)**：角色集合。一个 role 是一组 [task](004-task.md)。

## 存放路径

项目可存放在**内建**、**本地**或 **Git** 中。

### 内建

内建项目位于 `builtin` 目录，会集成到 KubeKey 命令中。

```shell
kk precheck
```

执行 `builtin` 目录下的 `playbooks/precheck.yaml`。

### 本地

```shell
kk run demo.yaml
```

执行当前目录的 `demo.yaml`。

### Git

```shell
kk run playbooks/demo.yaml \
  --project-addr="$(GIT_URL)" \
  --project-branch="$(GIT_BRANCH)"
```

执行指定 Git 地址与分支上的 `playbooks/demo.yaml`。
