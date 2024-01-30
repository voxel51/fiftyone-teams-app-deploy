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
Create a default name for the fiftyone app service
*/}}
{{- define "fiftyone-app.name" -}}
{{- if .Values.appSettings.service.name }}
{{- .Values.appSettings.service.name | trunc 63 | trimSuffix "-" }}
{{- else }}
"fiftyone-app"
{{- end }}
{{- end }}

{{/*
Create a default name for the teams cas service
*/}}
{{- define "teams-cas.name" -}}
{{- if .Values.casSettings.service.name }}
{{- .Values.casSettings.service.name | trunc 63 | trimSuffix "-" }}
{{- else }}
"fiftyone-teams-cas"
{{- end }}
{{- end }}

{{/*
Create a default name for the teams api service
*/}}
{{- define "teams-api.name" -}}
{{- if .Values.apiSettings.service.name }}
{{- .Values.apiSettings.service.name | trunc 63 | trimSuffix "-" }}
{{- else }}
"fiftyone-teams-api"
{{- end }}
{{- end }}

{{/*
Create a default name for the teams app service
*/}}
{{- define "teams-app.name" -}}
{{- if .Values.teamsAppSettings.service.name }}
{{- .Values.teamsAppSettings.service.name | trunc 63 | trimSuffix "-" }}
{{- else }}
"fiftyone-teams-app"
{{- end }}
{{- end }}

{{/*
Create a default name for the teams plugins service
*/}}
{{- define "teams-plugins.name" -}}
{{- if .Values.pluginsSettings.service.name }}
{{- .Values.pluginsSettings.service.name | trunc 63 | trimSuffix "-" }}
{{- else }}
"fiftyone-teams-plugins"
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
{{- define "fiftyone-teams-app.commonLabels" -}}
helm.sh/chart: {{ include "fiftyone-teams-app.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
API Selector labels
*/}}
{{- define "fiftyone-teams-api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "teams-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
API Combined labels
*/}}
{{- define "fiftyone-teams-api.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "fiftyone-teams-api.selectorLabels" . }}
{{- end }}


{{/*
APP Selector labels
*/}}
{{- define "fiftyone-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fiftyone-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
APP Combined labels
*/}}
{{- define "fiftyone-app.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "fiftyone-app.selectorLabels" . }}
{{- end }}

{{/*
CAS Selector labels
*/}}
{{- define "fiftyone-teams-cas.selectorLabels" -}}
app.kubernetes.io/name: {{ include "teams-cas.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
CAS Combined labels
*/}}
{{- define "fiftyone-teams-cas.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "fiftyone-teams-cas.selectorLabels" . }}
{{- end }}

{{/*
Plugins Selector labels
*/}}
{{- define "teams-plugins.selectorLabels" -}}
app.kubernetes.io/name: {{ include "teams-plugins.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Plugins Combined labels
*/}}
{{- define "teams-plugins.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "teams-plugins.selectorLabels" . }}
{{- end }}

{{/*
Teams APP Selector labels
*/}}
{{- define "fiftyone-teams-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fiftyone-teams-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Teams APP Combined labels
*/}}
{{- define "fiftyone-teams-app.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "fiftyone-teams-app.selectorLabels" . }}
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
Create a merged list of environment variables for fiftyone-teams-api
*/}}
{{- define "fiftyone-teams-api.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: AUTH0_API_CLIENT_ID
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: apiClientId
- name: AUTH0_API_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: apiClientSecret
- name: AUTH0_DOMAIN
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: auth0Domain
- name: AUTH0_CLIENT_ID
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: clientId
- name: FIFTYONE_DATABASE_NAME
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneDatabaseName
- name: FIFTYONE_DATABASE_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: mongodbConnectionString
- name: FIFTYONE_ENCRYPTION_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: encryptionKey
- name: MONGO_DEFAULT_DB
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneDatabaseName
- name: CAS_BASE_URL
  value: {{ printf "http://%s:%.0f/cas/api" .Values.casSettings.service.name .Values.casSettings.service.port | quote }}
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
{{- range $key, $val := .Values.apiSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- end -}}

{{/*
Create a merged list of environment variables for fiftyone-app
*/}}
{{- define "fiftyone-app.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .Values.apiSettings.service.name .Values.apiSettings.service.port | quote }}
- name: FIFTYONE_DATABASE_NAME
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneDatabaseName
- name: FIFTYONE_DATABASE_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: mongodbConnectionString
- name: FIFTYONE_ENCRYPTION_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: encryptionKey
- name: FIFTYONE_TEAMS_DOMAIN
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: auth0Domain
- name: FIFTYONE_TEAMS_AUDIENCE
  value: "https://$(FIFTYONE_TEAMS_DOMAIN)/api/v2/"
- name: FIFTYONE_TEAMS_CLIENT_ID
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: clientId
- name: FIFTYONE_TEAMS_ORGANIZATION
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: organizationId
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
{{- range $key, $val := .Values.appSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- end -}}

{{/*
Create a merged list of environment variables for fiftyone-teams-cas
*/}}
{{- define "teams-cas.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: CAS_MONGODB_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: mongodbConnectionString
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
- name: NEXTAUTH_URL
  value: {{ printf "https://%s/cas/api/auth" .Values.teamsAppSettings.dnsName | quote }}
- name: HUB_MONGODB_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: mongodbConnectionString
{{- range $key, $val := .Values.casSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- end -}}

{{/*
Create a merged list of environment variables for fiftyone-teams-plugins
*/}}
{{- define "teams-plugins.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .Values.apiSettings.service.name .Values.apiSettings.service.port | quote }}
- name: FIFTYONE_DATABASE_NAME
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneDatabaseName
- name: FIFTYONE_DATABASE_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: mongodbConnectionString
- name: FIFTYONE_DATABASE_ADMIN
  value: "false"
- name: FIFTYONE_ENCRYPTION_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: encryptionKey
- name: FIFTYONE_TEAMS_DOMAIN
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: auth0Domain
- name: FIFTYONE_TEAMS_AUDIENCE
  value: "https://$(FIFTYONE_TEAMS_DOMAIN)/api/v2/"
- name: FIFTYONE_TEAMS_CLIENT_ID
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: clientId
- name: FIFTYONE_TEAMS_ORGANIZATION
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: organizationId
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
{{- range $key, $val := .Values.pluginsSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- end -}}


{{/*
Create a merged list of environment variables for fiftyone-teams-app
*/}}
{{- define "fiftyone-teams-app.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .Values.apiSettings.service.name .Values.apiSettings.service.port | quote }}
- name: AUTH0_DOMAIN
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: auth0Domain
- name: AUTH0_AUDIENCE
  value: "https://$(AUTH0_DOMAIN)/api/v2/"
- name: AUTH0_BASE_URL
  value: {{ printf "https://%s" .Values.teamsAppSettings.dnsName | quote }}
- name: AUTH0_CLIENT_ID
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: clientId
- name: AUTH0_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: clientSecret
- name: AUTH0_ISSUER_BASE_URL
  value: "https://$(AUTH0_DOMAIN)"
- name: AUTH0_ORGANIZATION
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: organizationId
- name: AUTH0_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: cookieSecret
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
- name: FIFTYONE_SERVER_ADDRESS
  value: ""
- name: FIFTYONE_SERVER_PATH_PREFIX
  value: "/api/proxy/fiftyone-teams"
- name: FIFTYONE_TEAMS_PROXY_URL
  value: {{ printf "http://%s:%.0f" .Values.appSettings.service.name .Values.appSettings.service.port | quote }}
- name: FIFTYONE_TEAMS_PLUGIN_URL
{{- if .Values.pluginsSettings.enabled }}
  value: {{ printf "http://%s:%.0f" .Values.pluginsSettings.service.name .Values.pluginsSettings.service.port | quote }}
{{- else }}
  value: {{ printf "http://%s:%.0f" .Values.appSettings.service.name .Values.appSettings.service.port | quote }}
{{- end }}
- name: NEXTAUTH_BASEPATH
  value: "/cas/api/auth"

{{- range $key, $val := .Values.teamsAppSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- end -}}
