{{/*
Name of the telemetry redis Deployment/Service/PVC.
*/}}
{{- define "telemetry.redis.name" -}}
{{- printf "%s-telemetry-redis" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Resolves the URL for the telemetry Redis backend.

If `telemetry.redis.external.url` is set, returns it (chart skips the
bundled Redis Deployment/Service/PVC and consumer workloads + sidecars
are wired at the external URL instead — e.g. for managed Redis like
ElastiCache or MemoryStore). Otherwise returns the in-cluster Service
URL of the bundled Redis as a fully-qualified `<svc>.<ns>.svc.cluster.local`
hostname so cross-namespace consumers (e.g. delegated-operator Jobs
scheduled into a different namespace) still resolve it.

Always returns a non-empty URL when telemetry is enabled.
*/}}
{{- define "telemetry.redis.url" -}}
{{- if .Values.telemetry.redis.external.url -}}
{{- .Values.telemetry.redis.external.url -}}
{{- else -}}
{{- printf "redis://%s.%s.svc.cluster.local:6379" (include "telemetry.redis.name" .) .Values.namespace.name -}}
{{- end -}}
{{- end }}

{{/*
Selector labels for the telemetry redis Deployment/Service.
*/}}
{{- define "telemetry.redis.selectorLabels" -}}
app.kubernetes.io/name: telemetry-redis
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Combined labels for the telemetry redis Deployment/Service/PVC.
*/}}
{{- define "telemetry.redis.labels" -}}
{{- include "fiftyone-teams-app.commonLabels" . }}
{{ include "telemetry.redis.selectorLabels" . }}
app.voxel51.com/component: telemetry-redis
{{- end }}

{{/*
Name of the telemetry Role and RoleBinding for the sidecar's pods/log access.
*/}}
{{- define "telemetry.role.name" -}}
{{- printf "%s-telemetry-pod-logs" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Combined labels for the telemetry Role and RoleBinding.
*/}}
{{- define "telemetry.role.labels" -}}
{{- include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/name: telemetry
app.kubernetes.io/instance: {{ .Release.Name }}
app.voxel51.com/component: telemetry
{{- end }}

{{/*
Default subjects for the telemetry RoleBinding. When .Values.telemetry.serviceAccounts
is empty, bind to the main app service account — the SA used by the auto-injected
sidecars on app/plugins/delegated-operator pods.

The teams-api sidecar uses the teams-api RBAC service account, which is already
granted `pods/log` GET by api-role.yaml; binding it here would be a redundant
duplicate. When `apiSettings.rbac.create` is false, api-deployment falls back to
the main app SA anyway, which this RoleBinding still covers.
*/}}
{{- define "telemetry.role.subjects" -}}
{{- $appSA := include "fiftyone-teams-app.serviceAccountName" . | trim -}}
{{- $defaultSubjects := list $appSA -}}
{{- $subjects := .Values.telemetry.serviceAccounts | default $defaultSubjects -}}
{{- range $subjects }}
- kind: ServiceAccount
  name: {{ . }}
  namespace: {{ $.Values.namespace.name }}
{{- end }}
{{- end }}

{{/*
The shared base env vars for any telemetry-sidecar container.
Inputs (dict):
  ctx              — root context (.)
  serviceType      — value for SERVICE_TYPE env var (e.g. "teams-api")
  targetName       — value for TARGET_NAME env var (e.g. "fiftyone-teams-api")
  podName          — value for POD_NAME env var (defaults to fieldRef metadata.name)
  executor         — bool, when true emit EXECUTOR_SIDECAR=true and TELEMETRY_SOCKET env
  targetContainer  — when set, emit TARGET_CONTAINER env var (used by job sidecars)
*/}}
{{- define "telemetry.sidecar-env" -}}
{{- $secretName := .ctx.Values.secret.name -}}
- name: POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: SERVICE_TYPE
  value: {{ .serviceType | quote }}
- name: TARGET_NAME
  value: {{ .targetName | quote }}
{{- if .targetContainer }}
- name: TARGET_CONTAINER
  value: {{ .targetContainer | quote }}
{{- end }}
{{- if .executor }}
- name: EXECUTOR_SIDECAR
  value: "true"
- name: TELEMETRY_SOCKET
  value: /tmp/telemetry/agent.sock
{{- end }}
- name: FIFTYONE_TELEMETRY_REDIS_URL
  value: {{ include "telemetry.redis.url" .ctx | quote }}
- name: FIFTYONE_DATABASE_URI
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: mongodbConnectionString
- name: FIFTYONE_DATABASE_NAME
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: fiftyoneDatabaseName
{{- end }}

{{/*
A regular spec.containers entry: telemetry-sidecar for api/app/plugins/DO deployments.
Inputs: same dict as telemetry.sidecar-env.
*/}}
{{- define "telemetry.sidecar" -}}
- name: telemetry-sidecar
  image: "{{ .ctx.Values.telemetry.sidecar.image.repository }}:{{ .ctx.Values.telemetry.sidecar.image.tag | default .ctx.Chart.AppVersion }}"
  imagePullPolicy: {{ .ctx.Values.telemetry.sidecar.image.pullPolicy | default "Always" }}
  env:
    {{- include "telemetry.sidecar-env" . | nindent 4 }}
  {{- with .ctx.Values.telemetry.sidecar.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  securityContext:
    # Match the paired workload's UID so same-UID /proc reads work without elevated caps.
    {{- if .targetUid }}
    runAsUser: {{ .targetUid }}
    runAsNonRoot: {{ ne (int .targetUid) 0 }}
    {{- else }}
    runAsNonRoot: false
    runAsUser: 0
    {{- end }}
    allowPrivilegeEscalation: false
    capabilities:
      drop: ["ALL"]
      {{- if .executor }}
      add:
        - SYS_PTRACE
      {{- end }}
  {{- if .executor }}
  volumeMounts:
    - name: telemetry-socket
      mountPath: /tmp/telemetry
  {{- end }}
{{- end }}

{{/*
A native-sidecar (initContainer with restartPolicy: Always) variant — for use under
spec.initContainers in DO Jobs, where a regular sidecar container that does not exit
would block Job completion.
*/}}
{{- define "telemetry.native-sidecar" -}}
- name: telemetry-sidecar
  image: "{{ .ctx.Values.telemetry.sidecar.image.repository }}:{{ .ctx.Values.telemetry.sidecar.image.tag | default .ctx.Chart.AppVersion }}"
  imagePullPolicy: {{ .ctx.Values.telemetry.sidecar.image.pullPolicy | default "Always" }}
  restartPolicy: Always
  env:
    {{- include "telemetry.sidecar-env" . | nindent 4 }}
  {{- with .ctx.Values.telemetry.sidecar.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  readinessProbe:
    exec:
      command:
        - sh
        - -c
        - test -S /tmp/telemetry/agent.sock
    failureThreshold: 30
    periodSeconds: 1
  securityContext:
    # Match the paired workload's UID so same-UID /proc reads work without elevated caps.
    {{- if .targetUid }}
    runAsUser: {{ .targetUid }}
    runAsNonRoot: {{ ne (int .targetUid) 0 }}
    {{- else }}
    runAsNonRoot: false
    runAsUser: 0
    {{- end }}
    allowPrivilegeEscalation: false
    capabilities:
      drop: ["ALL"]
      {{- if .executor }}
      add:
        - SYS_PTRACE
      {{- end }}
  volumeMounts:
    - name: telemetry-socket
      mountPath: /tmp/telemetry
{{- end }}

{{/*
Emit a `FIFTYONE_TELEMETRY_REDIS_URL` env entry for a main workload container.
Renders empty when telemetry is disabled, allowing safe inclusion from
env-vars-list helpers.
*/}}
{{- define "telemetry.redis-url-env" -}}
{{- if .Values.telemetry.enabled }}
- name: FIFTYONE_TELEMETRY_REDIS_URL
  value: {{ include "telemetry.redis.url" . | quote }}
{{- end }}
{{- end }}
