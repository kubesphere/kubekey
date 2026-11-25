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

# Registry creation is retried because a freshly started harbor-core can
# return 502 for a second or two while it is still binding :8080. 403/409 mean
# the entry already exists and count as success; other 4xx responses are real
# request errors and fail immediately. All diagnostics go to stderr so the
# final registry-list curl remains the only stdout content for register_type: json.
function createRegistries() {
{{- range .groups.image_registry | default list }}
  {{- if ne . $.inventory_hostname }}
  local http_code
  local created=0
  local attempt
  for attempt in 1 2 3; do
    http_code=$(curl $curl_opts -s -o /dev/null -w "%{http_code}" \
      -u "$auth" \
      -X POST \
      -H "Content-Type: application/json" \
      "${base_url}/registries" \
      -d "{\"name\": \"{{ . }}\", \"type\": \"harbor\", \"url\":\"${scheme}://{{ . }}:${port}\", \"credential\": {\"access_key\": \"{{ $.image_registry.auth.username }}\", \"access_secret\": \"{{ $.image_registry.auth.password }}\"}, \"insecure\": true}")

    case "$http_code" in
      2*)
        created=1
        break
        ;;
      403|409)
        echo "registry {{ . }} already exists (HTTP ${http_code})" >&2
        created=1
        break
        ;;
      4*)
        echo "ERROR: failed to create registry {{ . }}, HTTP ${http_code}" >&2
        exit 1
        ;;
      *)
        echo "WARN: failed to create registry {{ . }}, HTTP ${http_code:-empty} (attempt ${attempt}/3)" >&2
        if [[ "$attempt" -lt 3 ]]; then
          sleep 5
        fi
        ;;
    esac
  done
  if [[ "$created" -ne 1 ]]; then
    echo "ERROR: failed to create registry {{ . }} after 3 attempts, last HTTP ${http_code:-empty}" >&2
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
