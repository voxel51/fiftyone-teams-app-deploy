{{/*
Create the name of the ConfigMap to use
*/}}
{{- define "delegated-operator-templates.config-map-name" }}
{{- if .Values.delegatedOperatorJobTemplates.configMap.name }}
{{- .Values.delegatedOperatorJobTemplates.configMap.name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := "do-templates" }}
{{- printf "%s-%s" (include "fiftyone-teams-app.fullname" .) $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "delegated-operator-templates.config-map-labels" }}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "delegated-operator-templates.config-map-name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: on-demand-delegated-operators
{{- with .Values.delegatedOperatorJobTemplates.configMap.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create a merged list of environment variables for delegated-operator templates
*/}}
{{- define "delegated-operator-templates.env-vars-list" }}
- name: POD_NAME
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: metadata.name
- name: API_URL
  value: {{ printf "http://%s:%.0f" .apiServiceName .apiServicePort | quote }}
- name: FIFTYONE_DATABASE_ADMIN
  value: "false"
- name: FIFTYONE_INTERNAL_SERVICE
  value: "true"
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
{{- if and .ctx .ctx.Values.telemetry.enabled }}
- name: FIFTYONE_TELEMETRY_REDIS_URL
  value: {{ include "telemetry.redis.url" .ctx | quote }}
- name: TELEMETRY_SOCKET
  value: /tmp/telemetry/agent.sock
{{- end }}
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
Delegated Operator Executor Selector labels
*/}}
{{- define "delegated-operator-templates.templateLabels" -}}
app.voxel51.com/delegate-operator-task-id: {{ `{{ _id }}` }}
app.voxel51.com/delegate-operator-task-type: delegated_operation
{{- end }}

{{- define "delegated-operator-templates.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
{{ include "delegated-operator-templates.templateLabels" . }}
app.voxel51.com/delegate-operator-template-name: {{ .jobTemplateName }}
{{- end }}

{{/*
Orchestrator registrations are derived from the delegated-operator template
maps, as a JSON list consumed by the seeding job (see
`../../files/seed_orchestrators.py` and `./seed-orchestrators-job.yaml).`

When `delegatedOperatorJobTemplates.[jobs|serviceOrchestrators].*.enabled=true`
and `delegatedOperatorJobTemplates.[jobs|serviceOrchestrators].*.registerOrchestrator=true`
(or inherited via `delegatedOperatorJobTemplates.template.registerOrchestrator=true`),
the orchestrator will be registered via the `seed_orchestrators.py` script.
The job or service key name is used as the `instance_id`.
`config.execution_tmpl_uri` contains the path to the entry's
rendered template file (mounted in teams-api from the do-templates
ConfigMap).
`available_operators` is the list of operator URIs an orchestrator may execute.
`job` entries omit it because the app's Refresh action discovers
the operators installed in the worker image and owns the list.
`service` orchestrators only ever execute one operator —
the builtin @voxel51/operators/run_service, which launches a service —
so seeding writes that single-entry list and re-applies it on every run.
*/}}
{{- define "delegated-operator-templates.seed-orchestrators" -}}
{{- $baseTpl := .Values.delegatedOperatorJobTemplates.template }}
{{- $namespace := .Values.namespace.name }}
{{- $orchestrators := list }}
{{- range $name, $config := .Values.delegatedOperatorJobTemplates.jobs }}
{{- $register := ternary $config.registerOrchestrator ($baseTpl.registerOrchestrator | default false) (hasKey $config "registerOrchestrator") }}
{{- if and (ne $config.enabled false) $register }}
{{- $mergedImage := merge (deepCopy ($config.image | default dict)) ($baseTpl.image) }}
{{- $orchestrators = append $orchestrators (dict
    "instance_id" $name
    "description" ($config.description | default (printf "Chart-managed job orchestrator %s" $name))
    "environment" "kubernetes"
    "config" (dict
      "image" (printf "%s:%s" $mergedImage.repository ($mergedImage.tag | default $.Chart.AppVersion))
      "execution_tmpl_uri" (printf "/tmp/do-targets/%s.yaml" $name)
      "namespace" $namespace)
    "secrets" (dict "kube_config" "")) }}
{{- end }}
{{- end }}
{{- range $name, $config := .Values.delegatedOperatorJobTemplates.serviceOrchestrators }}
{{- $register := ternary $config.registerOrchestrator ($baseTpl.registerOrchestrator | default false) (hasKey $config "registerOrchestrator") }}
{{- if and (ne $config.enabled false) $register }}
{{- $orchestrators = append $orchestrators (dict
    "instance_id" $name
    "description" ($config.description | default (printf "Chart-managed service orchestrator %s" $name))
    "environment" "kubernetes-service"
    "config" (dict
      "execution_tmpl_uri" (printf "/tmp/do-targets/%s.yaml" $name)
      "namespace" $namespace)
    "secrets" (dict "kube_config" "")
    "available_operators" (list "@voxel51/operators/run_service")) }}
{{- end }}
{{- end }}
{{- toJson $orchestrators -}}
{{- end }}
