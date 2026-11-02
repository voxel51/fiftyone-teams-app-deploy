{{/*
Name of the inference objects. The labeler's add "-labeler", hence 55.
*/}}
{{- define "inference.name" -}}
{{- .Release.Name | trunc 55 | trimSuffix "-" }}
{{- end }}

{{- define "inference.selectorLabels" -}}
app.kubernetes.io/name: fiftyone-inference
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "inference.labeler.selectorLabels" -}}
app.kubernetes.io/name: fiftyone-agentic-labeler
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "inference.chartLabels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
The Secret holding the shared token, under the key "token".
*/}}
{{- define "inference.tokenSecret" -}}
{{- .Values.sharedToken.secretName | default (printf "%s-token" (include "inference.name" .)) }}
{{- end }}

{{/*
Non empty when a backend is an LLM endpoint, which turns the labeler on.
*/}}
{{- define "inference.labeler.enabled" -}}
{{- range .Values.backends }}
{{- if eq (.kind | default "") "openai" }}true{{ end }}
{{- end }}
{{- end }}

{{/*
Pod settings both pods share. Values override the security defaults key by key.
*/}}
{{- define "inference.podSettings" -}}
securityContext:
  {{- toYaml (merge (deepCopy (.Values.podSecurityContext | default dict)) (dict "runAsNonRoot" true "runAsUser" 1000 "runAsGroup" 1000)) | nindent 2 }}
{{- with .Values.nodeSelector }}
nodeSelector:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.affinity }}
affinity:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.tolerations }}
tolerations:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{- define "inference.securityContext" -}}
securityContext:
  {{- toYaml (merge (deepCopy (.Values.securityContext | default dict)) (dict "allowPrivilegeEscalation" false)) | nindent 2 }}
{{- end }}

{{/*
Fails unless the model the labeler calls is served on its route: the
inference service serves a generation model on the route of its wireFormat.
*/}}
{{- define "inference.labeler.checkModel" -}}
{{- if ne (.model.wireFormat | default "") .wireFormat }}
{{- fail (printf "model %s of backend %s needs wireFormat: %s, the route the labeler calls it on" .model.name .id .wireFormat) }}
{{- end }}
{{- $tasks := .model.tasks | default list }}
{{- if not (and (kindIs "slice" $tasks) (has "generation" $tasks)) }}
{{- fail (printf "model %s of backend %s needs generation in its tasks" .model.name .id) }}
{{- end }}
{{- end }}

{{/*
BACKENDS: the shipped model server backend, unless off, then yours.
*/}}
{{- define "inference.backends" -}}
{{- $backends := list }}
{{- if .Values.modelServer.enabled }}
{{- $urls := .Values.modelServer.urls | default list }}
{{- if not $urls }}
{{- /* The fiftyone-model-server chart's Service per pod */}}
{{- range $i := until (int .Values.modelServer.replicas) }}
{{- $urls = append $urls (printf "http://%s-%d:8080" $.Values.modelServer.name $i) }}
{{- end }}
{{- end }}
{{- if not $urls }}
{{- fail "modelServer: no model server URL. Set modelServer.replicas to at least 1, or modelServer.urls" }}
{{- end }}
{{- $endpoints := list }}
{{- range $urls }}
{{- $endpoints = append $endpoints (dict "url" .) }}
{{- end }}
{{- /* The service keeps the last of two backends with one id, so a second
model-server would take over the shipped models */}}
{{- range .Values.backends }}
{{- if and (kindIs "map" .) (eq (.id | default "") "model-server") }}
{{- fail "backends: the id model-server is the shipped model server backend's: name yours otherwise, or set modelServer.enabled to false" }}
{{- end }}
{{- end }}
{{- $shipped := .Files.Get "catalog/model-server.yaml" | fromYaml }}
{{- if .Values.modelServer.tokenSecret }}
{{- /* The model servers' token, from MODEL_SERVER_TOKEN */}}
{{- $_ := set $shipped "auth" (dict "mode" "bearer-token" "tokenEnv" "MODEL_SERVER_TOKEN") }}
{{- end }}
{{- $backends = append $backends (set $shipped "endpoints" $endpoints) }}
{{- end }}
{{- $backends = concat $backends (.Values.backends | default list) }}
{{- if $backends }}
{{- toYaml $backends }}
{{- else }}
{{- "[]" }}
{{- end }}
{{- end }}

{{/*
LICENSE_KEY_FILE_PATHS: one file per license Secret, as mounted.
*/}}
{{- define "inference.licensePaths" -}}
{{- $paths := list }}
{{- range .Values.licenseSecrets }}
{{- $paths = append $paths (printf "/opt/fiftyone/licenses/%s" .) }}
{{- end }}
{{- join "," $paths }}
{{- end }}
