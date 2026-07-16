#!/bin/bash

{{ if $.image_registry.auth.plain_http | default false -}}
scheme="http"
port="{{ $.image_registry.harbor.http_port }}"
curl_opts=""
{{ else -}}
scheme="https"
port="{{ $.image_registry.harbor.https_port }}"
curl_opts="-k"
{{ end -}}

function createRegistries() {
{{- range .groups.image_registry | default list }}
  {{- if ne . $.inventory_hostname }}
  curl $curl_opts -u '{{ printf "%s:%s" $.image_registry.auth.username $.image_registry.auth.password }}' -X POST -H "Content-Type: application/json" "${scheme}://{{ $.inventory_hostname }}:${port}/api/v2.0/registries" -d "{\"name\": \"{{ . }}\", \"type\": \"harbor\", \"url\":\"${scheme}://{{ . }}:${port}\", \"credential\": {\"access_key\": \"{{ $.image_registry.auth.username }}\", \"access_secret\": \"{{ $.image_registry.auth.password }}\"}, \"insecure\": true}"
  {{- end }}
{{- end }}
}

function createReplication() {
{{- range .groups.image_registry | default list }}
  {{- if ne . $.inventory_hostname }}
  curl $curl_opts -u '{{ printf "%s:%s" $.image_registry.auth.username $.image_registry.auth.password }}' -X POST -H "Content-Type: application/json" "${scheme}://{{ $.inventory_hostname }}:${port}/api/v2.0/replication/policies" -d "{\"name\": \"{{ printf "%s_%s" $.inventory_hostname . }}\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 1, \"name\": \"{{ . }}\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
  {{- end }}
{{- end }}
}

createRegistries
createReplication
