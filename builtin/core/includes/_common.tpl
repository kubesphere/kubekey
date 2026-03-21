{{- define "images.platform.list" }}
{{- $platform := list }}
{{- range . }}
  {{- $platform = append $platform (printf "linux/%s" .) }}
{{- end }}
{{- $platform | toJson }}
{{- end -}}