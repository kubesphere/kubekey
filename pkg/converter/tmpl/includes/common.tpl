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