{{/*
Create the name of the ConfigMap to use
*/}}
{{- define "delegated-operator-templates.config-map-name" }}
{{- if .Values.delegatedOperatorJobTemplates.configMap.name }}
{{- .Values.delegatedOperatorJobTemplates.configMap.name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := "do-templates" }}
{{- printf "%s-%s" (include "fiftyone-teams-app.fullname" .) $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "delegated-operator-templates.config-map-labels" }}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "delegated-operator-templates.config-map-name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: do-templates
{{- with .Values.delegatedOperatorJobTemplates.configMap.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create a merged list of environment variables for delegated-operator templates
*/}}
{{- define "delegated-operator-templates.env-vars-list" }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .apiServiceName .apiServicePort | quote }}
- name: FIFTYONE_DATABASE_ADMIN
  value: "false"
- name: FIFTYONE_INTERNAL_SERVICE
  value: "true"
- name: FIFTYONE_DATABASE_NAME
  valueFrom:
    secretKeyRef:
      name: {{ .secretName }}
      key: fiftyoneDatabaseName
- name: FIFTYONE_DATABASE_URI
  valueFrom:
    secretKeyRef:
      name: {{ .secretName }}
      key: mongodbConnectionString
- name: FIFTYONE_ENCRYPTION_KEY
  valueFrom:
    secretKeyRef:
      name: {{ .secretName }}
      key: encryptionKey
{{- range $key, $val := .env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- range $key, $val := .secretEnv }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $val.secretName }}
      key: {{ $val.secretKey }}
{{- end }}
{{- end }}

{{/*
Delegated Operator Executor Selector labels
*/}}
{{- define "delegated-operator-templates.templateLabels" -}}
app.voxel51.com/delegate-operator-task-id: {{ `{{ _id }}` }}
app.voxel51.com/delegate-operator-task-type: delegated_operation
{{- end }}

{{- define "delegated-operator-templates.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "delegated-operator-templates.templateLabels" . }}
app.voxel51.com/delegate-operator-template-name: {{ .jobTemplateName }}
{{- end }}
