{{- define "images.platform.list" }}
{{- $platform := list }}
{{- range . }}
  {{- $platform = append $platform (printf "linux/%s" .) }}
{{- end }}
{{- $platform | toJson }}
{{- end -}}

{{- define "versions.archs.list" }}
{{- $result := list -}}
{{- range $arch := $.archs }}
  {{- range $version := $.versions }}
    {{- $result = append $result (dict "arch" $arch "version" $version) -}}
  {{- end }}
{{- end }}
{{- toJson $result }}
{{- end -}}

{{- define "ip.host.dns.list" }}
{{- $hostDnsDomains := .host_dns_domains|default list }}
{{- range .hostnames | default list }}
{{- dict "hostvars" $.hostvars "hostname" . "host_dns_domains" $hostDnsDomains| include "ip.host.dns" }}
{{- end }}
{{- end -}}

{{- define "ip.host.dns" }}
{{- $hostname := index .hostvars .hostname "hostname" }}
{{- $ipv4 := index .hostvars .hostname "internal_ipv4" }}
{{- $ipv6 := index .hostvars .hostname "internal_ipv6" }}
{{- $customDns := "" -}}
{{- range .host_dns_domains | default list}}
{{- $customDns = printf "%s %s.%s" $customDns $hostname .}}
{{- end }}
{{- if $ipv4 | empty | not }}
{{ $ipv4 }} {{ $hostname }}{{ $customDns }}
{{- end }}
{{- if $ipv6 | empty | not }}
{{ $ipv6 }} {{ $hostname }}{{ $customDns }}
{{- end }}
{{- end -}}