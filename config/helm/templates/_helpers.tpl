{{/*
Common labels
*/}}
{{- define "common.labels" -}}
helm.sh/chart: {{ include "common.chart" . }}
{{ include "common.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "common.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}


{{- define "common.image" -}}
{{- $registryName := .Values.operator.image.registry -}}
{{- $repositoryName := .Values.operator.image.repository -}}
{{- $separator := ":" -}}
{{- $termination := .Values.operator.image.tag | toString -}}
{{- if .Values.operator.image.digest }}
    {{- $separator = "@" -}}
    {{- $termination = .Values.operator.image.digest | toString -}}
{{- end -}}
{{- if $registryName }}
{{- printf "%s/%s%s%s" $registryName $repositoryName $separator $termination -}}
{{- else }}
{{- printf "%s%s%s" $repositoryName $separator $termination -}}
{{- end -}}
{{- end -}}
