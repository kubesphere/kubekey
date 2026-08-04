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

base_url="${scheme}://{{ $.inventory_hostname }}:${port}/api/v2.0"
auth='{{ printf "%s:%s" $.image_registry.auth.username $.image_registry.auth.password }}'

function createRegistries() {
{{- range .groups.image_registry | default list }}
  {{- if ne . $.inventory_hostname }}
  local http_code
  http_code=$(curl $curl_opts -s -o /dev/null -w "%{http_code}" \
    -u "$auth" \
    -X POST \
    -H "Content-Type: application/json" \
    "${base_url}/registries" \
    -d "{\"name\": \"{{ . }}\", \"type\": \"harbor\", \"url\":\"${scheme}://{{ . }}:${port}\", \"credential\": {\"access_key\": \"{{ $.image_registry.auth.username }}\", \"access_secret\": \"{{ $.image_registry.auth.password }}\"}, \"insecure\": true}")

  if [ -z "$http_code" ] || { [[ ! "$http_code" =~ ^2 ]] && [ "$http_code" != "409" ]; }; then
    echo "ERROR: failed to create registry {{ . }}, HTTP ${http_code}" >&2
    exit 1
  fi
  {{- end }}
{{- end }}
}

createRegistries

# Output the registry list as the sole stdout content so register_type: json
# can parse it.
curl $curl_opts -s -S -f \
  -u "$auth" \
  -H "Content-Type: application/json" \
  "${base_url}/registries"
