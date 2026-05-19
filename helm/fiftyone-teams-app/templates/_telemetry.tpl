{{/*
Name of the telemetry redis Deployment/Service/PVC.
*/}}
{{- define "telemetry.redis.name" -}}
{{- printf "%s-telemetry-redis" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Selector labels for the telemetry redis Deployment/Service.
*/}}
{{- define "telemetry.redis.selectorLabels" -}}
app.kubernetes.io/name: telemetry-redis
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Combined labels for the telemetry redis Deployment/Service/PVC.
*/}}
{{- define "telemetry.redis.labels" -}}
{{- include "fiftyone-teams-app.commonLabels" . }}
{{ include "telemetry.redis.selectorLabels" . }}
app.voxel51.com/component: telemetry-redis
{{- end }}

{{/*
Name of the telemetry Role and RoleBinding for the sidecar's pods/log access.
*/}}
{{- define "telemetry.role.name" -}}
{{- printf "%s-telemetry-pod-logs" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Combined labels for the telemetry Role and RoleBinding.
*/}}
{{- define "telemetry.role.labels" -}}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: telemetry
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: telemetry
{{- end }}
