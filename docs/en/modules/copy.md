# copy Module

The copy module allows users to copy files or directories to connected target hosts.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| src       | Source file or directory path | string | No (required if content is empty) | - |
| content   | Content of the source file or directory | string | No (required if src is empty) | - |
| dest      | Destination path on the target host | string | Yes | - |

## Usage Examples

1. Copy a relative path file to the target host  
The relative path is under the `files` directory corresponding to the current task. The current task path is specified by the task annotation `kubesphere.io/rel-path`.
```yaml
- name: copy relative path
  copy:
    src: a.yaml
    dest: /tmp/b.yaml
```

2. Copy an absolute path file to the target host  
Local absolute path file:
```yaml
- name: copy absolute path
  copy:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

3. Copy a directory to the target host  
Copies all files and directories under the directory to the target host:
```yaml
- name: copy dir
  copy:
    src: /tmp
    dest: /tmp
```

4. Copy content to a file on the target host
```yaml
- name: copy content
  copy:
    content: hello
    dest: /tmp/b.txt
```