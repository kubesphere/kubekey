# template Module

The template module allows users to parse a template file and copy it to the connected target host.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| src       | Path to the original file or directory | string | No (required if content is empty) | - |
| dest      | Destination path on the target host | string | Yes | - |

## Usage Examples

1. Copy a relative path file to the target host
Relative paths are under the `templates` directory of the current task. The current task path is specified by the task annotation `kubesphere.io/rel-path`.
```yaml
- name: copy relative path
  template:
    src: a.yaml
    dest: /tmp/b.yaml
```

2. Copy an absolute path file to the target host
Local template file with an absolute path
```yaml
- name: copy absolute path
  template:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

3. Copy a directory to the target host
Parse all template files in the directory and copy them to the target host
```yaml
- name: copy dir
  template:
    src: /tmp
    dest: /tmp
```

4. Copy content to the target host
```yaml
- name: copy content
  template:
    content: hello
    dest: /tmp/b.txt
```