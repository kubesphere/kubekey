# copy Module

Copy files, directories, or inline content to the connected target host.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| src | Source file or directory path | string | No (required when `content` is empty) | - |
| content | Inline content, written directly to target | string | No (required when `src` is empty) | - |
| dest | Path on target host | string | Yes | - |

- Relative paths are relative to the `files` directory for the current task; task path is specified by the task's annotation `kubesphere.io/rel-path`.
- Absolute paths can also be used to point to local files or directories.

## Examples

**1. Copy relative path file**

```yaml
- name: copy relative path
  copy:
    src: a.yaml
    dest: /tmp/b.yaml
```

**2. Copy absolute path file**

```yaml
- name: copy absolute path
  copy:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

**3. Copy directory**

Copies all files and subdirectories in the directory to the target host.

```yaml
- name: copy dir
  copy:
    src: /tmp
    dest: /tmp
```

**4. Copy inline content**

```yaml
- name: copy content
  copy:
    content: hello
    dest: /tmp/b.txt
```
