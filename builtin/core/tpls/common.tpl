{{- define "ip.host.dns.list" }}
{{- $hostSuffixDomains := .host_suffix_domains|default list }}
{{- range .nodes | default list }}
  {{- dict "hostvars" $.hostvars "node" . "host_suffix_domains" $hostSuffixDomains| include "ip.host.dns" }}
{{- end }}
{{- end -}}

{{- define "ip.host.dns" }}
{{- $hostname := index .hostvars .node "hostname" }}
{{- $ipv4 := index .hostvars .node "internal_ipv4" }}
{{- $ipv6 := index .hostvars .node "internal_ipv6" }}
{{- $domains := .domains|default $hostname }}
{{- range .host_suffix_domains | default list}}
  {{- $domains = printf "%s %s.%s" $domains $hostname .}}
{{- end }}
{{- if $ipv4 | empty | not }}
{{ $ipv4 }} {{ $domains }}
{{- end }}
{{- if $ipv6 | empty | not }}
{{ $ipv6 }} {{ $domains }}
{{- end }}
{{- end -}}