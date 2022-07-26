{{/*
Expand the name of the chart.
*/}}
{{- define "fiftyone-teams-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fiftyone-teams-app.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "fiftyone-teams-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "fiftyone-teams-app.labels" -}}
helm.sh/chart: {{ include "fiftyone-teams-app.chart" . }}
{{ include "fiftyone-teams-app.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "fiftyone-teams-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fiftyone-teams-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "fiftyone-teams-app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "fiftyone-teams-app.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create a merged list of environment variables
*/}}
{{- define "fiftyone-teams-app.env-vars-list" -}}
{{- $secretName := .Values.secret.name -}}
{{- range $key, $val := .Values.env.sensitive }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ $val }}
{{- end }}
{{- range $key, $val := .Values.env.nonsensitive }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- end -}}

{{/*
Create a port for the FiftyOne Teams app to run on and for Services to point to
*/}}
{{- define "fiftyone-teams.port" -}}
{{- if .Values.env.nonsensitive.FIFTYONE_DEFAULT_APP_PORT }}
{{- .Values.env.nonsensitive.FIFTYONE_DEFAULT_APP_PORT }}
{{- else }}
5151
{{- end }}
{{- end }}
