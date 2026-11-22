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
