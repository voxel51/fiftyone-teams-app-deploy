{{/*
Create the name of the builtin-services ConfigMap to use
*/}}
{{- define "service-orchestrator.builtin-services-config-map-name" }}
{{- if .Values.serviceOrchestrator.builtinServices.configMap.name }}
{{- .Values.serviceOrchestrator.builtinServices.configMap.name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-builtin-services" (include "fiftyone-teams-app.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "service-orchestrator.builtin-services-config-map-labels" }}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "service-orchestrator.builtin-services-config-map-name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: service-orchestrator
{{- with .Values.serviceOrchestrator.builtinServices.configMap.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}
