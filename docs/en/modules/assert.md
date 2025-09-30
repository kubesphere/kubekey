# assert Module

The assert module allows users to perform assertions on parameter conditions.

## Parameters

| Parameter   | Description | Type | Required | Default |
|-------------|-------------|------|----------|---------|
| that        | Assertion condition. Must use [template syntax](../101-syntax.md). | Array or string | Yes | - |
| success_msg | Message output to task result stdout when the assertion evaluates to true. | String | No | True |
| fail_msg    | Message output to task result stderr when the assertion evaluates to false. | String | No | False |
| msg         | Same as fail_msg. Lower priority than fail_msg. | String | No | False |

## Usage Examples

1. Assertion condition as a string
```yaml
- name: assert single condition
  assert:
    that: eq 1 1
```
Task execution result:  
stdout: "True"  
stderr: ""

2. Assertion condition as an array
```yaml
- name: assert multi-condition
  assert:
    that: 
     - eq 1 1
     - eq 1 2
```
Task execution result:  
stdout: "False"  
stderr: "False"

3. Set custom success output
```yaml
- name: assert is succeed
  assert:
    that: eq 1 1
    success_msg: "It's succeed"
```
Task execution result:  
stdout: "It's succeed"  
stderr: ""

4. Set custom failure output
```yaml
- name: assert is failed
  assert:
    that: eq 1 2
    fail_msg: "It's failed"
    msg: "It's failed!"
```
Task execution result:  
stdout: "False"  
stderr: "It's failed"