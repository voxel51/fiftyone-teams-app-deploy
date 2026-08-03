{{/*
Create the name of the builtin-services ConfigMap
*/}}
{{- define "service-orchestrator.builtin-services-config-map-name" }}
{{- printf "%s-builtin-services" (include "fiftyone-teams-app.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "service-orchestrator.builtin-services-config-map-labels" }}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: {{ include "service-orchestrator.builtin-services-config-map-name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: service-orchestrator
{{- end }}

{{/*
Flatten an env/secrets value to the newline-joined KEY=VALUE string the
service definitions carry.
A map becomes one KEY=VALUE line per entry;
a string passes through;
anything absent becomes "".
*/}}
{{- define "service-orchestrator.env-string" -}}
{{- if kindIs "map" . -}}
{{- range $key, $value := . -}}
{{ printf "%s=%v" $key $value }}
{{ end -}}
{{- else -}}
{{- . | default "" -}}
{{- end -}}
{{- end }}

{{/*
The builtin service definitions derived from the entries under
`delegatedOperatorJobTemplates.serviceOrchestrators.<name>.services`,
as a YAML list.
teams-api mounts the generated file and reads it via
`FIFTYONE_BUILTIN_SERVICES_PATH`;
entries deep-merge by `id` onto the service definitions packaged in
fiftyone at reconcile time
(`builtin_version` must be bumped for a changed entry to re-apply to an
environment that already stores the builtin).

Nesting a service under an orchestrator derives its identity fields from
the map keys, so they are written once:
`id`, `kind`, `name`, and `label` come from the service key
(`id` is prefixed `builtin:` when `builtin` resolves to true),
and `delegation_target` is the enclosing orchestrator key.
Explicitly set fields win over every derived default.
`autoStart` (default false) maps to the definition's `enabled` field.
`env` and `secrets` accept maps, flattened to KEY=VALUE lines.
An `entrypoint.container.image` without a tag defaults to the chart's
appVersion, so entries track image bumps instead of pinning a tag.
Tag detection looks at the segment after the last "/" so a registry port
(registry:5000/image) is not mistaken for a tag, and digests
(image@sha256:...) are left alone.
*/}}
{{- define "service-orchestrator.builtin-services" -}}
{{- $services := list }}
{{- range $orcName, $orcConfig := .Values.delegatedOperatorJobTemplates.serviceOrchestrators }}
{{- if ne $orcConfig.enabled false }}
{{- range $svcName, $svcConfig := ($orcConfig.services | default dict) }}
{{- $svc := omit (deepCopy $svcConfig) "autoStart" "env" "secrets" }}
{{- /* Resolved before the merge below: sprig's merge treats false in the
       destination as absent, so a boolean cannot ride the defaults dict. */}}
{{- $builtin := $svcConfig.builtin | default false }}
{{- $_ := set $svc "builtin" $builtin }}
{{- $_ := set $svc "enabled" ($svcConfig.autoStart | default false) }}
{{- $_ := set $svc "env" (include "service-orchestrator.env-string" $svcConfig.env) }}
{{- $_ := set $svc "secrets" (include "service-orchestrator.env-string" $svcConfig.secrets) }}
{{- $image := dig "entrypoint" "container" "image" "" $svc }}
{{- if and $image (not (contains ":" (last (splitList "/" $image)))) }}
{{- $_ := set $svc.entrypoint.container "image" (printf "%s:%s" $image $.Chart.AppVersion) }}
{{- end }}
{{- $svc = merge $svc (dict
    "id" (ternary (printf "builtin:%s" $svcName) $svcName $builtin)
    "kind" $svcName
    "name" $svcName
    "label" $svcName
    "delegation_target" $orcName
    "builtin_version" 1
    "scope" "shared") }}
{{- $services = append $services $svc }}
{{- end }}
{{- end }}
{{- end }}
{{- toYaml $services -}}
{{- end }}
