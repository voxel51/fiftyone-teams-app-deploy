{{/*
The in-cluster URL of the plugin server: the dedicated `teams-plugins`
service when it is enabled, otherwise `fiftyone-app`, which then serves the
plugins itself.
*/}}
{{- define "fiftyone-teams-app.plugin-server-url" -}}
{{- if .Values.pluginsSettings.enabled -}}
{{- printf "http://%s:%.0f" .Values.pluginsSettings.service.name (float64 .Values.pluginsSettings.service.port) -}}
{{- else -}}
{{- printf "http://%s:%.0f" .Values.appSettings.service.name (float64 .Values.appSettings.service.port) -}}
{{- end -}}
{{- end }}

{{/*
Emit a `FIFTYONE_WORKFLOW_SCHEDULE_API_URI` env entry for teams-api.
Renders empty when workflow schedules are disabled, allowing safe inclusion
from the env-vars-list helper.

The variable is both where and whether: teams-api runs the sweep that fires
the workflows *On Schedule* stage only when it is set, and posts each due
tick to the plugin server it names. No credential goes with it. teams-api
runs each tick as the user who started the workflow, with a short-lived
token it gets from CAS.
*/}}
{{- define "workflows.schedule-api-uri-env" -}}
{{- if .Values.workflowSettings.schedules.enabled }}
- name: FIFTYONE_WORKFLOW_SCHEDULE_API_URI
  value: {{ include "fiftyone-teams-app.plugin-server-url" . | quote }}
{{- end }}
{{- end }}

{{/*
Emit a `FIFTYONE_WORKFLOW_SCHEDULE_ENABLED` env entry for a plugin server
container (fiftyone-app, teams-plugins). Renders empty when workflow
schedules are disabled.

The plugin server offers the *On Schedule* stage, and lets a workflow with
one start, only when this is set. It is rendered from the same switch as the
sweep's URI so the two cannot disagree: a schedule that nothing fires would
leave its workflow standing forever without ever running.
*/}}
{{- define "workflows.schedule-enabled-env" -}}
{{- if .Values.workflowSettings.schedules.enabled }}
- name: FIFTYONE_WORKFLOW_SCHEDULE_ENABLED
  value: "true"
{{- end }}
{{- end }}

{{/*
The in-cluster URL the trigger worker posts dataset-event batches to: the
relay on `teams-api`. The worker appends `/operators/execute`.
*/}}
{{- define "workflows.trigger-relay-url" -}}
{{- printf "http://%s:%.0f/internal/workflow-triggers" .Values.apiSettings.service.name (float64 .Values.apiSettings.service.port) -}}
{{- end }}

{{/*
Fail the render when workflow dataset events are on without what feeds them.

The events come from Activity Core's database watch. With either off, the
stages would be offered and their workflows would start, and then never
receive anything, which looks like a healthy install doing nothing.
*/}}
{{- define "workflows.validate" -}}
{{- if .Values.workflowSettings.datasetEvents.enabled }}
{{- if not .Values.activitySettings.enabled }}
{{- fail "workflowSettings.datasetEvents.enabled is true but activitySettings.enabled is false: dataset events come from Activity Core, so no workflow would ever receive one. Enable activitySettings, or disable workflowSettings.datasetEvents." }}
{{- end }}
{{- if not .Values.activitySettings.watch.enabled }}
{{- fail "workflowSettings.datasetEvents.enabled is true but activitySettings.watch.enabled is false: the database watch is what sees samples being added and deleted, so no workflow would ever receive one. Enable activitySettings.watch, or disable workflowSettings.datasetEvents." }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Emit a `FIFTYONE_WORKFLOW_TRIGGER_API_URI` env entry for teams-api. Renders
empty when workflow dataset events are disabled.

The variable is both where and whether: teams-api answers the trigger
worker's relay route only when it is set, and runs each batch on the plugin
server it names, as the user who started the workflow.
*/}}
{{- define "workflows.trigger-api-uri-env" -}}
{{- if .Values.workflowSettings.datasetEvents.enabled }}
- name: FIFTYONE_WORKFLOW_TRIGGER_API_URI
  value: {{ include "fiftyone-teams-app.plugin-server-url" . | quote }}
{{- end }}
{{- end }}

{{/*
Emit a `FIFTYONE_WORKFLOW_DATASET_EVENTS_ENABLED` env entry for a plugin
server container (fiftyone-app, teams-plugins). Renders empty when workflow
dataset events are disabled.

The plugin server offers the dataset-event stages, and lets a workflow with
one start, only when this is set. It is rendered from the same switch as the
trigger worker so the two cannot disagree.
*/}}
{{- define "workflows.dataset-events-enabled-env" -}}
{{- if .Values.workflowSettings.datasetEvents.enabled }}
- name: FIFTYONE_WORKFLOW_DATASET_EVENTS_ENABLED
  value: "true"
{{- end }}
{{- end }}

{{/*
Emit the workflow dataset-event env entries for one Activity Core worker.
Renders empty when workflow dataset events are disabled.
Inputs (dict):
  ctx     — root context (.)
  worker  — the key from `activitySettings.workers` (e.g. "ingest")

The ingest and rollup workers publish sample changes for workflows. The
trigger worker reads them and posts each batch to the relay on teams-api,
naming itself with the deployment's auth secret, as teams-api does to CAS.
*/}}
{{- define "workflows.activity-worker-env" -}}
{{- if .ctx.Values.workflowSettings.datasetEvents.enabled }}
{{- if has .worker (list "ingest" "rollup") }}
- name: FIFTYONE_ACTIVITY_WORKFLOWS_TOPIC_ENABLED
  value: "true"
{{- end }}
{{- if eq .worker "trigger" }}
- name: FIFTYONE_ACTIVITY_TRIGGER_API_URI
  value: {{ include "workflows.trigger-relay-url" .ctx | quote }}
- name: FIFTYONE_ACTIVITY_TRIGGER_API_KEY
  valueFrom:
    secretKeyRef:
      name: {{ .ctx.Values.secret.name }}
      key: fiftyoneAuthSecret
{{- end }}
{{- end }}
{{- end }}
