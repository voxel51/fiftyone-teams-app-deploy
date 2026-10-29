{{/*
Name of every object this release creates.
*/}}
{{- define "model-server.name" -}}
{{- .Values.nameOverride | default .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Selector labels, which the Service and NetworkPolicy match on.
*/}}
{{- define "model-server.selectorLabels" -}}
app.kubernetes.io/name: fiftyone-model-server
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "model-server.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ include "model-server.selectorLabels" . }}
{{- end }}
