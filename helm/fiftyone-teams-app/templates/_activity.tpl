{{/*
Fail the render when activity is enabled without a queue to reach.

Turning activity on while `fiftyoneMq.enabled` is false renders the
workers and the producer env with nothing to connect to, so every event is
dropped. That looks like a healthy install recording nothing, which is why
it is a hard failure rather than a note in values.yaml.

Deliberately one-directional. The reverse — queue on, activity off — is a
supported state: the queue being reachable is not consent to emit, which
is the whole reason the gate is a flag rather than an inference from
`FIFTYONE_MQ_REDIS_URL`.
*/}}
{{- define "activity.validate" -}}
{{- if and .Values.activitySettings.enabled (not .Values.fiftyoneMq.enabled) }}
{{- fail "activitySettings.enabled is true but fiftyoneMq.enabled is false: the activity workers and producers have no queue to reach, so every event is dropped. Enable fiftyoneMq, or disable activitySettings." }}
{{- end }}
{{- end }}

{{/*
Name of an Activity Core worker Deployment.
Inputs (dict):
  ctx     — root context (.)
  worker  — the key from `activitySettings.workers` (e.g. "ingest")
*/}}
{{- define "activity.worker.name" -}}
{{- printf "%s-activity-%s-worker" .ctx.Release.Name .worker | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Selector labels for an Activity Core worker Deployment.
Inputs: same dict as activity.worker.name.
*/}}
{{- define "activity.worker.selectorLabels" -}}
app.kubernetes.io/name: {{ printf "activity-%s-worker" .worker }}
app.kubernetes.io/instance: {{ .ctx.Release.Name }}
{{- end }}

{{/*
Combined labels for an Activity Core worker Deployment.
Inputs: same dict as activity.worker.name.
*/}}
{{- define "activity.worker.labels" -}}
{{- include "fiftyone-teams-app.commonLabels" .ctx }}
{{ include "activity.worker.selectorLabels" . }}
app.voxel51.com/component: {{ printf "activity-%s-worker" .worker }}
{{- end }}

{{/*
Emit a `FIFTYONE_ACTIVITY_ENABLED` env entry for a main workload container.
Renders empty when Activity Core is disabled, allowing safe inclusion
from env-vars-list helpers.

This is the single gate the producer side reads: `emit()`, `flush()`, and the
operator mutation capture all no-op when it is unset or false, before any
queue client is constructed. It deliberately carries no default in the chart
— the var is either rendered as "true" or not rendered at all, so an absent
var and an explicit "false" mean the same thing to the workload.

Prior to this the producers inferred enablement from `FIFTYONE_MQ_REDIS_URL`
being set. That signal cannot distinguish "unset" from "explicitly pointed at
localhost", because the URL carries a default of its own downstream; and it
conflated *whether* to emit with *where* to emit. The flag is the gate; the
URL is only how the queue is reached once enabled.
*/}}
{{- define "activity.enabled-env" -}}
{{- if .Values.activitySettings.enabled }}
- name: FIFTYONE_ACTIVITY_ENABLED
  value: "true"
{{- end }}
{{- end }}

{{/*
Emit the `VFF_WF_ACTIVITY` env entry for the `teams-app` container.
Renders empty when Activity Core is disabled.

`activity.enabled-env` gates *capture*; this gates the surfaces that display
what was captured. They are separate layers, but they are not independently
useful: a deployment that set `activitySettings.enabled=true` collected and
rolled up events with no way to view them, because every viewing surface —
the workflow Activity tab and the label History panel — sits behind this
frontend flag. Nothing distinguished "no activity yet" from "no UI wired
up". So enabling capture now graduates the surfaces that read it.

`VFF_WF_METRIC` is deliberately NOT rendered here. It gates the pre-release
Metrics tab, which stays a per-environment opt-in — set it through
`teamsAppSettings.env` if you want it. Coupling it to this flag would ship an
unreleased surface to every deployment that turns on capture.

Rendered before the `teamsAppSettings.env` passthrough so that an explicit
entry there still wins: later duplicates take precedence in a container's
env list, which leaves a deployment able to force the flag off.
*/}}
{{- define "activity.ui-flags-env" -}}
{{- if .Values.activitySettings.enabled }}
- name: VFF_WF_ACTIVITY
  value: "true"
{{- end }}
{{- end }}

{{/*
Emit a `FIFTYONE_ACTIVITY_MONGO_DB` env entry for a main workload container.
Renders empty when Activity Core is disabled OR when no dedicated
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
Renders empty when Activity Core is disabled or no organization is set,
allowing safe inclusion from env-vars-list helpers.

Unset is the normal case, and is NOT a misconfiguration: the producers stamp
the authenticated organization carried on the request, and the workers — which
have no request to read — discover a single-org deployment's organization from
CAS. The value is an override for a deployment holding several organizations,
where the deployment-wide state counts cannot be attributed to one of them.
*/}}
{{- define "activity.org-id-env" -}}
{{- if and .Values.activitySettings.enabled .Values.activitySettings.orgId }}
- name: FIFTYONE_ACTIVITY_ORG_ID
  value: {{ .Values.activitySettings.orgId | quote }}
{{- end }}
{{- end }}
