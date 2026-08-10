{{/*
Name of an Activity Analytics worker Deployment.
Inputs (dict):
  ctx     — root context (.)
  worker  — the key from `activitySettings.workers` (e.g. "ingest")
*/}}
{{- define "activity.worker.name" -}}
{{- printf "%s-activity-%s-worker" .ctx.Release.Name .worker | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Selector labels for an Activity Analytics worker Deployment.
Inputs: same dict as activity.worker.name.
*/}}
{{- define "activity.worker.selectorLabels" -}}
app.kubernetes.io/name: {{ printf "activity-%s-worker" .worker }}
app.kubernetes.io/instance: {{ .ctx.Release.Name }}
{{- end }}

{{/*
Combined labels for an Activity Analytics worker Deployment.
Inputs: same dict as activity.worker.name.
*/}}
{{- define "activity.worker.labels" -}}
{{- include "fiftyone-teams-app.commonLabels" .ctx }}
{{ include "activity.worker.selectorLabels" . }}
app.voxel51.com/component: {{ printf "activity-%s-worker" .worker }}
{{- end }}

{{/*
Emit a `FIFTYONE_ACTIVITY_MONGO_DB` env entry for a main workload container.
Renders empty when Activity Analytics is disabled OR when no dedicated
database is configured, allowing safe inclusion from env-vars-list helpers.

Only the dedicated-database override is ever emitted: when
`activitySettings.mongo.database` is unset the activity collections are
co-located in the per-deployment FiftyOne database, and the readers fall
back to the standard `FIFTYONE_DATABASE_NAME` on their own. Emitting the
var unconditionally would pin readers to a value that can drift from what
the workers template resolves (it honors `mongo.database`) — a dedicated-DB
deployment would then write database X and read database Y with no error.
*/}}
{{- define "activity.mongo-db-env" -}}
{{- if and .Values.activitySettings.enabled .Values.activitySettings.mongo.database }}
- name: FIFTYONE_ACTIVITY_MONGO_DB
  value: {{ .Values.activitySettings.mongo.database | quote }}
{{- end }}
{{- end }}

{{/*
Emit a `FIFTYONE_ACTIVITY_ORG_ID` env entry for a main workload container.
Renders empty when Activity Analytics is disabled or no organization is set,
allowing safe inclusion from env-vars-list helpers.

Rollups are scoped per organization, so events emitted without one are not
returned by the read path.
*/}}
{{- define "activity.org-id-env" -}}
{{- if and .Values.activitySettings.enabled .Values.activitySettings.orgId }}
- name: FIFTYONE_ACTIVITY_ORG_ID
  value: {{ .Values.activitySettings.orgId | quote }}
{{- end }}
{{- end }}
