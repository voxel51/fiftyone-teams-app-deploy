{{/*
Name of the telemetry redis Deployment/Service/PVC.
*/}}
{{- define "telemetry.redis.name" -}}
{{- printf "%s-telemetry-redis" .Release.Name | trunc 63 | trimSuffix "-" -}}
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
is empty, bind to both the main app service account and the teams-api RBAC service
account (covering the SAs used by the auto-injected sidecars).
*/}}
{{- define "telemetry.role.subjects" -}}
{{- $appSA := include "fiftyone-teams-app.serviceAccountName" . | trim -}}
{{- $apiSA := include "teams-api-rbac.service-account-name" . | trim -}}
{{- $defaultSubjects := list $appSA $apiSA | uniq -}}
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
  value: {{ printf "redis://%s:6379" (include "telemetry.redis.name" .ctx) | quote }}
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
  image: {{ .ctx.Values.telemetry.sidecar.image | default "us-central1-docker.pkg.dev/computer-vision-team/dev-docker/fiftyone-telemetry-sidecar:v0.1.62" | quote }}
  imagePullPolicy: {{ .ctx.Values.telemetry.sidecar.imagePullPolicy | default "Always" }}
  env:
    {{- include "telemetry.sidecar-env" . | nindent 4 }}
  {{- with .ctx.Values.telemetry.sidecar.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  securityContext:
    # The sidecar image runs as root (SYS_PTRACE + /proc/<pid>/fd/1 access
    # require it). Explicit container-level override so it works even when
    # the pod's podSecurityContext sets runAsNonRoot: true.
    runAsNonRoot: false
    runAsUser: 0
    capabilities:
      add:
        - SYS_PTRACE
  {{- if .executor }}
  volumeMounts:
    - name: telemetry-socket
      mountPath: /tmp/telemetry
  {{- end }}
{{- end }}

{{/*
A native-sidecar (initContainer with restartPolicy: Always) variant — for use under
spec.initContainers, primarily in DO Jobs where a regular extraContainer that does
not exit would block Job completion.
*/}}
{{- define "telemetry.native-sidecar" -}}
- name: telemetry-sidecar
  image: {{ .ctx.Values.telemetry.sidecar.image | default "us-central1-docker.pkg.dev/computer-vision-team/dev-docker/fiftyone-telemetry-sidecar:v0.1.62" | quote }}
  imagePullPolicy: {{ .ctx.Values.telemetry.sidecar.imagePullPolicy | default "Always" }}
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
    # The sidecar image runs as root (SYS_PTRACE + /proc/<pid>/fd/1 access
    # require it). Explicit container-level override so it works even when
    # the pod's podSecurityContext sets runAsNonRoot: true.
    runAsNonRoot: false
    runAsUser: 0
    capabilities:
      add:
        - SYS_PTRACE
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
  value: {{ printf "redis://%s:6379" (include "telemetry.redis.name" .) | quote }}
{{- end }}
{{- end }}

{{/*
Emit the `telemetry-socket` shared emptyDir volume entry for spec.volumes.
Used by DO deployments and DO jobs so the sidecar's unix socket can be
reached by both the executor and the sidecar.
*/}}
{{- define "telemetry.socket-volume" -}}
- name: telemetry-socket
  emptyDir: {}
{{- end }}

{{/*
Emit the `telemetry-socket` volumeMount for a workload container that needs to
reach the sidecar's unix socket. Pair with telemetry.socket-volume.
*/}}
{{- define "telemetry.socket-volume-mount" -}}
- name: telemetry-socket
  mountPath: /tmp/telemetry
{{- end }}
