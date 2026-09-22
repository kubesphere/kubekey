#!/bin/bash
# Registry configuration for containerd v2.
#
# containerd v2 no longer reads registry.configs.<host>.tls and refuses to combine
# registry.mirrors with registry.config_path, so every registry kubekey configures is
# described by its own /etc/containerd/certs.d/<host>/hosts.toml. The TLS material itself is
# placed by the registry tasks of this role; this script only references it and leaves out a
# file that is not there, so containerd never loads a configuration pointing at a missing
# certificate.
set -e
certs_dir=/etc/containerd/certs.d
mkdir -p "${certs_dir}"

{{ if .cri.registry.mirrors | empty | not }}
# cri.registry.mirrors: every mirror is tried before registry-1.docker.io is used.
mkdir -p "${certs_dir}/docker.io"
{
  echo 'server = "https://registry-1.docker.io"'
{{ range .cri.registry.mirrors }}
  echo ''
  echo '[host."{{ . }}"]'
  echo '  capabilities = ["pull", "resolve"]'
{{ end }}
} > "${certs_dir}/docker.io/hosts.toml"
{{ end }}

{{ if .cri.registry.insecure_registries | empty | not }}
# cri.registry.insecure_registries: these endpoints serve no TLS, so the scheme is held in a
# variable, which also keeps the literal out of the generated script, and they are not verified.
insecure_scheme=http
{{ range .cri.registry.insecure_registries }}
mkdir -p "${certs_dir}/{{ . }}"
{
  echo "server = \"${insecure_scheme}://{{ . }}\""
  echo 'skip_verify = true'
} > "${certs_dir}/{{ . }}/hosts.toml"
{{ end }}
{{ end }}

{{ if and (.image_registry.auth.registry | empty | not) (ne (.image_registry.auth.registry | splitList "/" | first) "hub.kubesphere.com.cn") }}
# image_registry: the certificate and key copied into the registry directory are used when
# they are present, otherwise the registry is reachable through the system trust store.
{{ $registryHost := .image_registry.auth.registry | splitList "/" | first }}
{{ $scheme := .image_registry.auth.plain_http | default false | ternary "http" "https" }}
{{ $skipTLS := true }}
{{ if hasKey .image_registry.auth "skip_tls_verify" }}
{{ $skipTLS = .image_registry.auth.skip_tls_verify | toBool }}
{{ end }}
mkdir -p "${certs_dir}/{{ $registryHost }}"
{
  echo 'server = "{{ $scheme }}://{{ $registryHost }}"'
  if [ -f "${certs_dir}/{{ $registryHost }}/ca.crt" ]; then
    echo 'ca = "/etc/containerd/certs.d/{{ $registryHost }}/ca.crt"'
  fi
  if [ -f "${certs_dir}/{{ $registryHost }}/client.crt" ] && [ -f "${certs_dir}/{{ $registryHost }}/client.key" ]; then
    echo 'client = [["/etc/containerd/certs.d/{{ $registryHost }}/client.crt", "/etc/containerd/certs.d/{{ $registryHost }}/client.key"]]'
  fi
  echo 'skip_verify = {{ $skipTLS }}'
} > "${certs_dir}/{{ $registryHost }}/hosts.toml"
{{ end }}

{{ if .cri.registry.auths | default list | empty | not }}
# cri.registry.auths: the file paths of the entry are reused as they are.
{{ range .cri.registry.auths | default list }}
{{ $registryHost := .registry | splitList "/" | first }}
{{ $skipTLS := true }}
{{ if hasKey . "skip_tls_verify" }}
{{ $skipTLS = .skip_tls_verify | toBool }}
{{ end }}
mkdir -p "${certs_dir}/{{ $registryHost }}"
{
  echo 'server = "https://{{ $registryHost }}"'
{{ if .ca_file }}
  echo 'ca = "{{ .ca_file }}"'
{{ end }}
{{ if .cert_file }}
  echo 'client = [["{{ .cert_file }}", "{{ .key_file | default .cert_file }}"]]'
{{ end }}
  echo 'skip_verify = {{ $skipTLS }}'
} > "${certs_dir}/{{ $registryHost }}/hosts.toml"
{{ end }}
{{ end }}
