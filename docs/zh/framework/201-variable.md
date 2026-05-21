# 变量

变量分为**静态变量**（运行前定义）与**动态变量**（运行中生成）。  
优先级：**动态变量 > 静态变量**。

## 静态变量

包含：节点清单（inventory）、全局配置（config）、模板中定义的参数。  
优先级：**全局配置 > 节点清单 > 模板中定义的参数**。

### 节点清单

YAML 格式，**不支持**模板语法。通过 `-i` 传入（`kk -i inventory.yaml ...`），在每个 host 上生效。

**规范示例**：

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    hostname1:
      k1: v1
    hostname2:
      k2: v2
    hostname3: {}
  groups:
    groupname1:
      groups:
        - groupname2
      hosts:
        - hostname1
      vars:
        k1: v1
    groupname2: {}
  vars:
    k1: v1
```

| 字段 | 说明 |
|------|------|
| **hosts** | key 为 host 名，value 为该 host 的变量。 |
| **groups** | 对 host 分组。key 为组名，value 含 `groups`、`hosts`、`vars`。`groups` 为嵌套组，`hosts` 为该组包含的 host，`vars` 为该组变量，对组内所有 host 生效。组内 host 集合 = `groups` 中的 host ∪ `hosts` 中的 host。 |
| **vars** | 全局变量，对所有 host 生效。 |

变量优先级：**host 变量 > 组变量 > 全局变量**。

### 全局配置

YAML 格式，**不支持**模板语法。通过 `-c` 传入（`kk -c config.yaml ...`），在每个 host 上生效。

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
metadata:
  name: default
spec:
  k: v
```

`spec` 下可任意定义键值，与 [节点清单](#节点清单) 等一起参与变量解析。

### 模板中定义的参数

包括：

- playbook 的 `vars`、`vars_files`
- role 的 `defaults/main.yml`、`vars`
- task 的 `vars`

## 动态变量

运行时产生，包含：

- **gather_facts** 采集的主机信息
- **register** 注册的 task 输出
- **set_fact** 设置的变量

同一名字时，**后定义的覆盖先定义的**。
