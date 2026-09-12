{{/*
Name of the fiftyone-mq redis Deployment/Service.
*/}}
{{- define "fiftyone-mq.redis.name" -}}
{{- printf "%s-fiftyone-mq-redis" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Resolves the URL for the fiftyone-mq queue Redis backend.

If `fiftyoneMq.redis.external.url` is set, returns it (chart skips the bundled
Redis Deployment/Service and producers and workers are wired at the external URL
instead). Otherwise returns the in-cluster Service URL of the bundled Redis as a
fully-qualified `<svc>.<ns>.svc.cluster.local` hostname so cross-namespace
consumers (e.g. delegated-operator Jobs scheduled into a different namespace)
still resolve it.

External instances must set `maxmemory-policy noeviction`. An evicting policy
drops queued jobs.

Always returns a non-empty URL when fiftyone-mq is enabled.
*/}}
{{- define "fiftyone-mq.redis.url" -}}
{{- if .Values.fiftyoneMq.redis.external.url -}}
{{- .Values.fiftyoneMq.redis.external.url -}}
{{- else -}}
{{- printf "redis://%s.%s.svc.cluster.local:6379/0" (include "fiftyone-mq.redis.name" .) .Values.namespace.name -}}
{{- end -}}
{{- end }}

{{/*
Selector labels for the fiftyone-mq redis Deployment/Service.
*/}}
{{- define "fiftyone-mq.redis.selectorLabels" -}}
app.kubernetes.io/name: fiftyone-mq-redis
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Combined labels for the fiftyone-mq redis Deployment/Service.
*/}}
{{- define "fiftyone-mq.redis.labels" -}}
{{- include "fiftyone-teams-app.commonLabels" . }}
{{ include "fiftyone-mq.redis.selectorLabels" . }}
app.voxel51.com/component: fiftyone-mq-redis
{{- end }}

{{/*
Emit a `FIFTYONE_MQ_REDIS_URL` env entry for a main workload container.
Renders empty when fiftyone-mq is disabled, allowing safe inclusion from
env-vars-list helpers.
*/}}
{{- define "fiftyone-mq.redis-url-env" -}}
{{- if .Values.fiftyoneMq.enabled }}
- name: FIFTYONE_MQ_REDIS_URL
  value: {{ include "fiftyone-mq.redis.url" . | quote }}
{{- end }}
{{- end }}
