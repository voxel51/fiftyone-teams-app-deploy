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
Delegated Operator Executor Selector labels
*/}}
{{- define "delegated-operator-deployments.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Delegated Operator Executor Combined labels
*/}}
{{- define "delegated-operator-deployments.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "delegated-operator-deployments.selectorLabels" . }}
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

NOTE: Selector labels are immutable.
We will keep app.kubernetes.io/name
as fiftyone-teams-app.name and not teams-app.name.
*/}}
{{- define "teams-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fiftyone-teams-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Teams APP Combined labels
*/}}
{{- define "teams-app.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "teams-app.selectorLabels" . }}
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
Service Account labels
*/}}
{{- define "fiftyone-teams-app.serviceAccountLabels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ default (include "fiftyone-teams-app.fullname" .) .Values.serviceAccount.name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- with .Values.serviceAccount.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Ingress labels
*/}}
{{- define "fiftyone-teams-app.ingressLabels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "fiftyone-teams-app.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- with .Values.ingress.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Secret labels
*/}}
{{- define "fiftyone-teams-app.secretLabels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "fiftyone-teams-app.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- with .Values.secret.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Common Topology Constraints
*/}}
{{- define "fiftyone-teams-app.commonTopologySpreadConstraints" -}}
{{- range $constraint := .constraints -}}
- maxSkew: {{ $constraint.maxSkew }}
  {{- if $constraint.minDomains }}
  minDomains: {{ $constraint.minDomains }}
  {{- end }}
  topologyKey: {{ $constraint.topologyKey }}
  whenUnsatisfiable: {{ $constraint.whenUnsatisfiable }}
  {{- if $constraint.labelSelector }}
  labelSelector:
    {{- $constraint.labelSelector | toYaml | nindent 4 }}
  {{- else }}
  labelSelector:
    matchLabels:
      {{- include $.selectorLabels $.context | nindent 6 }}
  {{- end }}
  {{- if $constraint.matchLabelKeys }}
  matchLabelKeys:
    {{- $constraint.matchLabelKeys | toYaml | nindent 4 }}
  {{- end }}
  {{- if $constraint.nodeAffinityPolicy }}
  nodeAffinityPolicy: {{ $constraint.nodeAffinityPolicy }}
  {{- end }}
  {{- if $constraint.nodeTaintsPolicy }}
  nodeTaintsPolicy: {{ $constraint.nodeTaintsPolicy }}
  {{- end }}
{{ end }}
{{- end }}

{{/*
Common Init Containers
*/}}
{{- define "fiftyone-teams-app.commonInitContainers" -}}
- name: init-cas
  image: {{ $.repository }}:{{ $.tag }}
  command:
    - 'sh'
    - '-c'
    - "until wget -qO /dev/null {{ $.casServiceName }}.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local/cas/api; do echo waiting for cas; sleep 2; done"
  {{- if hasKey $ "resources" }}
  resources:
    {{- toYaml $.resources | nindent 4 }}
  {{- end }}
  {{- if hasKey $ "containerSecurityContext" }}
  securityContext:
    {{- toYaml $.containerSecurityContext | nindent 4 }}
  {{- end }}
{{- end }}

{{/*
Create a merged list of environment variables for delegated-operator-executor
*/}}
{{- define "delegated-operator-deployments.env-vars-list" }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .apiServiceName .apiServicePort | quote }}
- name: FIFTYONE_DATABASE_ADMIN
  value: "false"
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
Create a merged list of environment variables for fiftyone-teams-api
*/}}
{{- define "fiftyone-teams-api.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: CAS_BASE_URL
  value: {{ printf "http://%s:%.0f/cas/api" .Values.casSettings.service.name (float64 .Values.casSettings.service.port) | quote }}
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
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
{{- range $key, $val := .Values.apiSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- range $key, $val := .Values.apiSettings.secretEnv }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $val.secretName }}
      key: {{ $val.secretKey }}
{{- end }}
{{- end -}}

{{/*
Create a merged list of environment variables for fiftyone-app
*/}}
{{- define "fiftyone-app.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .Values.apiSettings.service.name (float64 .Values.apiSettings.service.port) | quote }}
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
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
{{- range $key, $val := .Values.appSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- range $key, $val := .Values.appSettings.secretEnv }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $val.secretName }}
      key: {{ $val.secretKey }}
{{- end }}
{{- end -}}

{{/*
Create a string that contains all license files to be created in the
`teams-cas` deployment
*/}}
{{- define "teams-cas.license-key-file-paths" }}
{{- $licensePaths := "" }}
{{- range $i, $name := .Values.fiftyoneLicenseSecrets }}
{{- if $i }}
{{- $licensePaths = print $licensePaths "," }}
{{- end }}
{{- $licensePaths = print $licensePaths "/opt/fiftyone/licenses/" $name }}
{{- end }}
{{- print $licensePaths }}
{{- end }}

{{/*
Create a merged list of environment variables for fiftyone-teams-cas
*/}}
{{- define "teams-cas.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: CAS_MONGODB_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ .Values.casSettings.env.CAS_MONGODB_URI_KEY }}
- name: CAS_URL
  value: {{ printf "https://%s" .Values.teamsAppSettings.dnsName | quote }}
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
- name: FIFTYONE_ENCRYPTION_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: encryptionKey
- name: LICENSE_KEY_FILE_PATHS
  value: {{ include "teams-cas.license-key-file-paths" . | quote }}
- name: NEXTAUTH_URL
  value: {{ printf "https://%s/cas/api/auth" .Values.teamsAppSettings.dnsName | quote }}
- name: TEAMS_API_DATABASE_NAME
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneDatabaseName
- name: TEAMS_API_MONGODB_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: mongodbConnectionString
{{- range $key, $val := .Values.casSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- range $key, $val := .Values.casSettings.secretEnv }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $val.secretName }}
      key: {{ $val.secretKey }}
{{- end }}
{{- end -}}

{{/*
Create a merged list of environment variables for fiftyone-teams-plugins
*/}}
{{- define "teams-plugins.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .Values.apiSettings.service.name (float64 .Values.apiSettings.service.port) | quote }}
- name: FIFTYONE_AUTH_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneAuthSecret
- name: FIFTYONE_DATABASE_ADMIN
  value: "false"
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
{{- range $key, $val := .Values.pluginsSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- range $key, $val := .Values.pluginsSettings.secretEnv }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $val.secretName }}
      key: {{ $val.secretKey }}
{{- end }}
{{- end -}}


{{/*
Create a merged list of environment variables for fiftyone-teams-app
*/}}
{{- define "fiftyone-teams-app.env-vars-list" -}}
{{- $secretName := .Values.secret.name }}
- name: API_URL
  value: {{ printf "http://%s:%.0f" .Values.apiSettings.service.name (float64 .Values.apiSettings.service.port) | quote }}
- name: FIFTYONE_API_URI
{{- if .Values.teamsAppSettings.fiftyoneApiOverride }}
  value: {{ .Values.teamsAppSettings.fiftyoneApiOverride }}
{{- else if .Values.apiSettings.dnsName }}
  value: {{ printf "https://%s" .Values.apiSettings.dnsName }}
{{- else }}
  value: {{ printf "https://%s" .Values.teamsAppSettings.dnsName }}
{{- end }}
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
  value: {{ printf "http://%s:%.0f" .Values.appSettings.service.name (float64 .Values.appSettings.service.port) | quote }}
- name: FIFTYONE_TEAMS_PLUGIN_URL
{{- if .Values.pluginsSettings.enabled }}
  value: {{ printf "http://%s:%.0f" .Values.pluginsSettings.service.name (float64 .Values.pluginsSettings.service.port) | quote }}
{{- else }}
  value: {{ printf "http://%s:%.0f" .Values.appSettings.service.name (float64 .Values.appSettings.service.port) | quote }}
{{- end }}
{{- range $key, $val := .Values.teamsAppSettings.env }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- range $key, $val := .Values.teamsAppSettings.secretEnv }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ $val.secretName }}
      key: {{ $val.secretKey }}
{{- end }}
{{- end -}}
