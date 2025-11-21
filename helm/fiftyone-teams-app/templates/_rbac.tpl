{{/*
Create the name of the K8s Job Delegation Role
*/}}
{{- define "rbac-do-templates.role-name" }}
{{- if .Values.apiSettings.rbac.role.name }}
{{- .Values.apiSettings.rbac.role.name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := "do-management" }}
{{- printf "%s-%s" (include "fiftyone-teams-app.fullname" .) $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create the name of the K8s Job Delegation Role Binding
*/}}
{{- define "rbac-do-templates.role-binding-name" }}
{{- if .Values.apiSettings.rbac.roleBinding.name }}
{{- .Values.apiSettings.rbac.roleBinding.name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := "do-management" }}
{{- printf "%s-%s" (include "fiftyone-teams-app.fullname" .) $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create the name of the K8s Job Service Account
*/}}
{{- define "rbac-do-templates.service-account-name" }}
{{- if .Values.apiSettings.rbac.serviceAccount.name }}
{{- .Values.apiSettings.rbac.serviceAccount.name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := "teams-api" }}
{{- printf "%s-%s" (include "fiftyone-teams-app.fullname" .) $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create the labels of the K8s Job Delegation Role
*/}}
{{- define "rbac-do-templates.role-labels" }}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "rbac-do-templates.role-name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: on-demand-delegated-operators
{{- with .Values.apiSettings.rbac.role.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create the labels of the K8s Job Delegation Role Binding
*/}}
{{- define "rbac-do-templates.role-binding-labels" }}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "rbac-do-templates.role-binding-name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: on-demand-delegated-operators
{{- with .Values.apiSettings.rbac.roleBinding.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create the labels of the K8s Job Delegation Service Account
*/}}
{{- define "rbac-do-templates.service-account-labels" }}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "rbac-do-templates.service-account-name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: teams-api
{{- with .Values.apiSettings.rbac.serviceAccount.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}
