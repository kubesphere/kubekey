# 任务执行模块 
module定义了一个任务实际要执行的操作
## assert
用于断言host上的variable是否满足某个条件
```yaml
assert:
  that: I'm assertion statement
  success_msg: I'm success message
  fail_msg: I'm failed message
  msg: I'm failed message
```
**that**: 断言语句, 必填.值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**success_msg**: 成功时输出, 非必填, 默认值为"True".  
**fail_msg**: 失败时输出, 非必填, 默认值为"True".  
**msg**: 失败时输出, 非必填, 默认值为"True". 优先输出`fail_msg`.
## command/shell
执行命令, command和shell的用法相同
```yaml
command: I'm command statement
```
值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.

## copy
复制本地文件到host.
```yaml
copy:
  src: srcpath
  content: srcpath
  dest: destpath
  mode: 0755
```
**src**: 来源地址, 可以为绝对路径或相对路径, 可以是目录或者文件, 非必填(`content`未定义时, 必填). 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.
- 绝对路径: 从执行命令的机器上的绝对路径上获取.
- 相对路径: 从`project_dir`中获取, 获取顺序: $(project_dir)/roles/roleName/files/$(srcpath) > $(project_dir)/playbooks/.../$(current_playbook)/roles/$(roleName)/files/$(srcpath) > $(project_dir)/files/$(srcpath).  
**content**: 来源文件内容, 非必填(`src`未定义时, 必填). 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**dest**: 目标地址, host上的绝对路径, 可以是目录或者文件(与`src`对应, 如果为文件,需要在末尾添加"/"), 必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**mode**: 复制到host上的文件权限, 非必填, 默认源文件权限.
## fetch
从host上获取文件到本地.
```yaml
fetch:
  src: srcpath
  dest: destpath
```
**src**: 来源文件地址, host上的绝对路径, 必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.
**dest**: 目标文件地址, 本地绝对路径, 可以是目录或者文件(与`src`对应, 如果为文件,需要在末尾添加"/"), 必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.
## debug
打印信息
```yaml
debug:
  var: I'm variable statement
  msg: I'm message statement
```
**var**: 打印变量, 非必填, 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**msg**: 打印信息, 非必填, 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.
## template
templates中的文件内容采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
将文件内容转换为实际文件后,复制本地文件到host.  
```yaml
template:
  src: srcpath
  dest: destpath
  mode: 0755
```
**src**: 来源地址, 可以为绝对路径或相对路径, 可以是目录或者文件, 必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.
- 绝对路径: 从执行命令的机器上的绝对路径上获取.
- 相对路径: 从`project_dir`中获取, 获取顺序: $(project_dir)/roles/roleName/templates/$(srcpath) > $(project_dir)/playbooks/.../$(current_playbook)/roles/$(roleName)/templates/$(srcpath) > $(project_dir)/templates/$(srcpath).
**dest**: 目标地址, host上的绝对路径, 可以是目录或者文件(与`src`对应, 如果为文件,需要在末尾添加"/"), 必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.
**mode**: 复制到host上的文件权限, 非必填, 默认源文件权限.
## set_fact
给所有host设置variable. 层级结构保持不变  
```yaml
set_fact:
  key: value
```
**key**: 必填, 可以为可以为多级结构(比如{k1:{k2:value}}).
**value**: 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.
## gen_cert
在工作目录生成证书, $(work_dir)/kubekey/pki/
```yaml
gen_cert:
  root_key: keypath
  root_cert: certpath
  date: 87600h
  sans: ["ip1","dns1"]
  cn: common name
  out_key: keypath
  out_cert: certpath
```
**root_key**: 父证书的key文件绝对路径, 用于生成子证书, 非必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**root_cert**: 父证书的cert文件绝对路径, 用于生成子证书, 非必填,  值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值. 
**date**: 证书失效时间, 时间间隔格式(单位: s,m,h), 非必填, 默认10年.  
**sans**: Subject Alternate Names, 支持数组或数组类型的json字符串格式, 非必填.  
**cn**: Common Name. 必填.
**out_key**: 输出的证书key文件绝对路径, 必填, 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**out_cert**: 输出的证书cert文件绝对路径, 必填, 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
当root_key或root_cert未定义时, 生成自签名证书.  
## image
拉取镜像到本地目录, 或推送镜像到远程服务器
```yaml
image:
  skip_tls_verify: true
  pull: ["image1", "image2"]
  push:
    registry: local.kubekey
    username: username
    password: password
    namespace_override: new_namespace
```
**skip_tls_verify**: 跳过证书认证. 默认true.
**pull**: 拉取镜像到本地工作目录, 非必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**push**: 推送工作目录中的镜像到远程仓库, 非必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**registry**: 远程仓库地址, 必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**username**: 远程仓库认证用户, 非必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**password**: 远程仓库认证密码, 非必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
**namespace_override**: 是否用新的路径, 覆盖镜像原来的路径, 非必填. 值采用[模板语法](101-syntax.md)编写, 对每个的host单独计算值.  
