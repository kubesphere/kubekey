# template Module

Render template files using [template syntax](../101-syntax.md) and copy to the connected target host.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| src | Template file or directory path | string | Yes (or with content) | - |
| dest | Path on target host | string | Yes | - |

- Relative paths are relative to the `templates` directory for the current task; task path is specified by the task's annotation `kubesphere.io/rel-path`.
- Absolute paths can also be used to point to local template files.

## Examples

**1. Copy relative path template**

```yaml
- name: template relative path
  template:
    src: a.yaml
    dest: /tmp/b.yaml
```

**2. Copy absolute path template**

```yaml
- name: template absolute path
  template:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

**3. Copy template directory**

Renders all templates in the directory and copies to the target host.

```yaml
- name: template dir
  template:
    src: /tmp/tpl
    dest: /tmp
```
