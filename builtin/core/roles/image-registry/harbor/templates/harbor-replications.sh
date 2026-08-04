#!/bin/bash

{{ if $.image_registry.auth.plain_http | default false -}}
scheme="http"
port="{{ $.image_registry.http_port | default 80 }}"
curl_opts=""
{{ else -}}
scheme="https"
port="{{ $.image_registry.https_port | default 443 }}"
curl_opts="-k"
{{ end -}}

# Build a name -> registry id map from the API response registered earlier.
{{- $registryIdMap := dict }}
{{- range .install_harbor_registries.stdout }}
{{- $_ := set $registryIdMap .name .id }}
{{- end }}

function createReplication() {
{{- range .groups.image_registry | default list }}
  {{- if ne . $.inventory_hostname }}
  {{- $peer := . }}
  {{- $peerId := index $registryIdMap $peer | default -1 }}
  local peer_id="{{ $peerId }}"
  if [ "$peer_id" -lt 0 ]; then
    echo "ERROR: registry id for {{ $peer }} not found in /api/v2.0/registries response" >&2
    exit 1
  fi
  curl $curl_opts -u '{{ printf "%s:%s" $.image_registry.auth.username $.image_registry.auth.password }}' -X POST -H "Content-Type: application/json" "${scheme}://{{ $.inventory_hostname }}:${port}/api/v2.0/replication/policies" -d "{\"name\": \"{{ printf "%s_%s" $.inventory_hostname . }}\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": {{ $peerId }}, \"name\": \"{{ . }}\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
  {{- end }}
{{- end }}
}

createReplication
