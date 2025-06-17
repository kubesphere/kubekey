# assert 模块

assert模块允许用户对参数条件进行断言。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| that | 断言的条件.需采用[模板语法](../101-syntax.md)编写  | 数组或字符串 | 是 | - |
| success_msg | 断言结果为true时，输出到任务结果的 stdout 信息 | 字符串 | 否 | True |
| fail_msg | 断言结果为false时，输出到任务结果的 stderr 信息 | 字符串 | 否 | False |
| msg | 通fail_msg. 优先级低于fail_msg | 字符串 | 否 | False |

## 使用示例

1. 断言条件为字符串
```yaml
- name: assert single condition
  assert:
    that: eq 1 1
```
任务执行结果:
stdout: "True"
stderr: ""

2. 断言条件为数组
```yaml
- name: assert multi-condition
  assert:
    that: 
     - eq 1 1
     - eq 1 2
```
任务执行结果:
stdout: "False"
stderr: "False"

3. 设置成功输出
```yaml
- name: assert is succeed
  assert:
    that: eq 1 1
    success_msg: "It's succeed"
```
任务执行结果:
stdout: "It's succeed"
stderr: ""

1. 设置失败输出
```yaml
- name: assert is failed
  assert:
    that: eq 1 2
    fail_msg: "It's failed"
    msg: "It's failed!"
```
任务执行结果:
stdout: "False"
stderr: "It's failed"