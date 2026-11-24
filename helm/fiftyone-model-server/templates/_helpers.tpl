{{/*
Name of every object this release creates. 52 leaves room for the pod
index and the StatefulSet's revision hash label.
*/}}
{{- define "model-server.name" -}}
{{- .Values.nameOverride | default .Release.Name | trunc 52 | trimSuffix "-" }}
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

{{/*
The Secret holding the bearer token, under the key "token", or nothing.
*/}}
{{- define "model-server.tokenSecret" -}}
{{- if .Values.token.secretName }}
{{- .Values.token.secretName }}
{{- else if .Values.token.value }}
{{- printf "%s-token" (include "model-server.name" .) }}
{{- end }}
{{- end }}

{{/*
How many pods may exist: each gets its own Service.
*/}}
{{- define "model-server.maxPods" -}}
{{- if .Values.autoscaling.enabled }}
{{- if eq .Values.service.type "LoadBalancer" }}
{{- fail "autoscaling with a LoadBalancer would make a load balancer per pod it may start: autoscaled pods are found by DNS inside the cluster" }}
{{- end }}
{{- if lt (int .Values.autoscaling.maxReplicas) (int .Values.autoscaling.minReplicas) }}
{{- fail "autoscaling.maxReplicas must be at least autoscaling.minReplicas" }}
{{- end }}
{{- .Values.autoscaling.maxReplicas | int }}
{{- else }}
{{- .Values.replicas | int }}
{{- end }}
{{- end }}
