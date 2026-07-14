#!/bin/bash

function createRegistries() {
{{- range .groups.image_registry | default list }}
  {{- if ne . $.inventory_hostname }}
  curl -k -u '{{ printf "%s:%s" $.image_registry.auth.username $.image_registry.auth.password }}' -X POST -H "Content-Type: application/json" "https://{{ $.inventory_hostname }}:{{ $.image_registry.harbor.https_port | default 443 }}/api/v2.0/registries" -d "{\"name\": \"{{ . }}\", \"type\": \"harbor\", \"url\":\"https://{{ . }}:{{ $.image_registry.harbor.https_port | default 443 }}\", \"credential\": {\"access_key\": \"{{ $.image_registry.auth.username }}\", \"access_secret\": \"{{ $.image_registry.auth.password }}\"}, \"insecure\": true}"
  {{- end }}
{{- end }}
}

function createReplication() {
{{- range .groups.image_registry | default list }}
  {{- if ne . $.inventory_hostname }}
  curl -k -u '{{ printf "%s:%s" $.image_registry.auth.username $.image_registry.auth.password }}' -X POST -H "Content-Type: application/json" "https://{{ $.inventory_hostname }}:{{ $.image_registry.harbor.https_port | default 443 }}/api/v2.0/replication/policies" -d "{\"name\": \"{{ printf "%s_%s" $.inventory_hostname . }}\", \"enabled\": true, \"deletion\":true, \"override\":true, \"replicate_deletion\":true, \"dest_registry\":{ \"id\": 1, \"name\": \"{{ . }}\"}, \"trigger\": {\"type\": \"event_based\"}, \"dest_namespace_replace_count\":1 }"
  {{- end }}
{{- end }}
}

createRegistries
createReplication
