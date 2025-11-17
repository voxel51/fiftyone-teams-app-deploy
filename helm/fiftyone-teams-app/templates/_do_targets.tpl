{{/*
Default execution profiles for all backends
*/}}
{{- define "fiftyone-teams-app.defaultExecutionProfiles" -}}
kubernetes:
  # https://learn.microsoft.com/en-us/azure/aks/use-nvidia-gpu
  "voxel51.com/aks":
    resources:
      limits:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
      requests:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
    tolerations:
      - effect: NoSchedule
        key: sku
        operator: Equal
        value: gpu

  # https://docs.aws.amazon.com/eks/latest/userguide/auto-accelerated.html
  "voxel51.com/eks-auto":
    resources:
      limits:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
      requests:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'

  # https://aws.amazon.com/blogs/compute/running-gpu-accelerated-kubernetes-workloads-on-p3-and-p2-ec2-instances-with-amazon-eks/
  "voxel51.com/eks-standard":
    resources:
      limits:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
      requests:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'

  # https://docs.cloud.google.com/kubernetes-engine/docs/how-to/autopilot-gpus
  "voxel51.com/gke-autopilot":
    resources:
      limits:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
      requests:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
    nodeSelector:
      cloud.google.com/gke-accelerator: '{{ `{{ .gpu.type }}` }}'
      cloud.google.com/gke-accelerator-count: '{{ `{{ .gpu.count }}` }}'

  # https://docs.cloud.google.com/kubernetes-engine/docs/how-to/gpus
  "voxel51.com/gke-standard":
    resources:
      limits:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
      requests:
        nvidia.com/gpu: '{{ `{{ .gpu.count }}` }}'
{{- end -}}

{{/*
Merge default profiles with user-provided profiles
*/}}
{{- define "fiftyone-teams-app.executionProfiles" -}}
{{- $defaults := include "fiftyone-teams-app.defaultExecutionProfiles" . | fromYaml -}}
{{- $userProfiles := .Values.delegatedOperatorJobTemplates.customExecutionProfiles | default dict -}}
{{- $merged := dict -}}

{{- range $backend, $profiles := $defaults }}
  {{- $userBackendProfiles := index $userProfiles $backend | default dict -}}
  {{- $mergedBackend := dict -}}

  {{- range $profileName, $profileConfig := $profiles }}
    {{- $_ := set $mergedBackend $profileName $profileConfig -}}
  {{- end -}}

  {{- range $profileName, $profileConfig := $userBackendProfiles }}
    {{- $_ := set $mergedBackend $profileName $profileConfig -}}
  {{- end -}}

  {{- $_ := set $merged $backend $mergedBackend -}}
{{- end -}}

{{- range $backend, $profiles := $userProfiles }}
  {{- if not (hasKey $defaults $backend) }}
    {{- $_ := set $merged $backend $profiles -}}
  {{- end -}}
{{- end -}}

{{- toYaml $merged -}}
{{- end -}}

{{/*
Get a specific execution profile
Usage: {{ include "fiftyone-teams-app.getExecutionProfile" (dict "backend" "kubernetes" "profile" "voxel51.com/gke-autopilot" "root" $) }}
*/}}
{{- define "fiftyone-teams-app.getExecutionProfile" -}}
{{- $allProfiles := include "fiftyone-teams-app.executionProfiles" .root | fromYaml -}}
{{- $backendProfiles := index $allProfiles .backend -}}
{{- if not $backendProfiles -}}
  {{- fail (printf "Backend '%s' not found in execution profiles" .backend) -}}
{{- end -}}
{{- $profile := index $backendProfiles .profile -}}
{{- if not $profile -}}
  {{- $availableProfiles := keys $backendProfiles | sortAlpha | join ", " -}}
  {{- fail (printf "Profile '%s' not found in backend '%s'. Available profiles: %s" .profile .backend $availableProfiles) -}}
{{- end -}}
{{- toYaml $profile -}}
{{- end -}}

{{/*
List all available profiles for a backend
Usage: {{ include "fiftyone-teams-app.listProfilesForBackend" (dict "backend" "kubernetes" "root" $) }}
*/}}
{{- define "fiftyone-teams-app.listProfilesForBackend" -}}
{{- $allProfiles := include "fiftyone-teams-app.executionProfiles" .root | fromYaml -}}
{{- $backendProfiles := index $allProfiles .backend | default dict -}}
{{- keys $backendProfiles | sortAlpha | join ", " -}}
{{- end -}}

{{/*
Validate that a profile exists
Usage: {{ include "fiftyone-teams-app.validateProfile" (dict "backend" "kubernetes" "profile" "voxel51.com/gke-autopilot" "jobName" "my-job" "root" $) }}
*/}}
{{- define "fiftyone-teams-app.validateProfile" -}}
{{- $allProfiles := include "fiftyone-teams-app.executionProfiles" .root | fromYaml -}}
{{- $backendProfiles := index $allProfiles .backend -}}
{{- if not $backendProfiles -}}
  {{- fail (printf "Job '%s': Backend '%s' not found. Available backends: %s" .jobName .backend (keys $allProfiles | sortAlpha | join ", ")) -}}
{{- end -}}
{{- if not (hasKey $backendProfiles .profile) -}}
  {{- $availableProfiles := keys $backendProfiles | sortAlpha | join ", " -}}
  {{- fail (printf "Job '%s': Profile '%s' not found in backend '%s'. Available profiles: %s" .jobName .profile .backend $availableProfiles) -}}
{{- end -}}
{{- end -}}

{{/*
Delegated Operator Templates Combined labels
*/}}
{{- define "delegated-operator-templates.labels" -}}
{{ include "fiftyone-teams-app.commonLabels" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Render job based on backend type
Usage: {{ include "fiftyone-teams-app.renderJob" (dict "jobName" $jobName "job" $jobConfig "root" $) }}
*/}}
{{- define "fiftyone-teams-app.renderJob" -}}
{{- $jobName := .jobName -}}
{{- $job := .job -}}
{{- $root := .root -}}

{{- /* Get the backend from job config */ -}}
{{- $backend := $job.backend | required (printf "Job '%s' must specify a backend" $jobName) -}}

{{- /* Get all profiles and validate */ -}}
{{- $allProfiles := include "fiftyone-teams-app.executionProfiles" $root | fromYaml -}}
{{- include "fiftyone-teams-app.validateProfile" (dict "backend" $backend "profile" $job.profile "jobName" $jobName "root" $root) -}}

{{- /* Get the specific profile */ -}}
{{- $profile := include "fiftyone-teams-app.getExecutionProfile" (dict "backend" $backend "profile" $job.profile "root" $root) | fromYaml -}}

{{- /* Route to backend-specific renderer */ -}}
{{- if eq $backend "kubernetes" -}}
  {{- include "fiftyone-teams-app.renderKubernetesJob" (dict "jobName" $jobName "job" $job "profile" $profile "root" $root) -}}
{{- end -}}
{{- end -}}

{{/*
Render Kubernetes Job
*/}}
{{- define "fiftyone-teams-app.renderKubernetesJob" -}}
{{- $jobName := .name -}}
{{- $jobConfig := .job -}}
{{- $profile := .profile -}}
{{- $root := .root -}}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ $jobName }}-{{ `{{ _id }}` }}
  {{- with $jobConfig.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  labels:
    {{- with $jobConfig.labels }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
    task-id: {{ `{{ _id }}` }}
    task-type: delegated_operation
spec:
  backoffLimit: {{ $jobConfig.backoffLimit | default 3 }}
  {{- if $jobConfig.ttlSecondsAfterFinished }}
  ttlSecondsAfterFinished: {{ $jobConfig.ttlSecondsAfterFinished }}
  {{- end }}
  {{- if $jobConfig.activeDeadlineSeconds }}
  activeDeadlineSeconds: {{ $jobConfig.activeDeadlineSeconds }}
  {{- end }}
  {{- if $jobConfig.completions }}
  completions: {{ $jobConfig.completions }}
  {{- end }}
  template:
    metadata:
      {{- with $jobConfig.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- with $jobConfig.podLabels }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
        task-id: {{ `{{ _id }}` }}
        task-type: delegated_operation
    spec:
      restartPolicy: {{ $jobConfig.restartPolicy | default "Never" }}
      {{- if $jobConfig.serviceAccountName }}
      serviceAccountName: {{ $jobConfig.serviceAccountName }}
      {{- end }}
      {{- if $jobConfig.podSecurityContext }}
      securityContext:
        {{- toYaml $jobConfig.podSecurityContext | nindent 8 }}
      {{- end }}
      {{- if $jobConfig.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml $jobConfig.imagePullSecrets | nindent 8 }}
      {{- end }}
      {{- /* Merge nodeSelector from profile and job */ -}}
      {{- $mergedNodeSelector := dict }}
      {{- if $profile.nodeSelector }}
        {{- range $key, $value := $profile.nodeSelector }}
          {{- if contains "{{ .gpu.count }}" $value }}
            {{- $_ := set $mergedNodeSelector $key ($jobConfig.gpu.count | toString) }}
          {{- else if contains "{{ .gpu.type }}" $value }}
            {{- $_ := set $mergedNodeSelector $key ($jobConfig.gpu.type | toString) }}
          {{- else }}
            {{- $_ := set $mergedNodeSelector $key $value }}
          {{- end }}
        {{- end }}
      {{- end }}
      {{- if $jobConfig.nodeSelector }}
        {{- range $key, $value := $jobConfig.nodeSelector }}
          {{- $_ := set $mergedNodeSelector $key $value }}
        {{- end }}
      {{- end }}
      {{- if $mergedNodeSelector }}
      nodeSelector:
        {{- toYaml $mergedNodeSelector | nindent 8 }}
      {{- end }}
      {{- /* Merge tolerations from profile and job */ -}}
      {{- $mergedTolerations := list }}
      {{- if $profile.tolerations }}
        {{- $mergedTolerations = concat $mergedTolerations $profile.tolerations }}
      {{- end }}
      {{- if $jobConfig.tolerations }}
        {{- $mergedTolerations = concat $mergedTolerations $jobConfig.tolerations }}
      {{- end }}
      {{- if $mergedTolerations }}
      tolerations:
        {{- toYaml $mergedTolerations | nindent 8 }}
      {{- end }}
      {{- with $jobConfig.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      containers:
      - name: {{ $jobName }}-{{ `{{ _id }}` }}
        image: "{{ $jobConfig.image.repository }}:{{ $jobConfig.image.tag | default $root.Chart.AppVersion }}"
        imagePullPolicy: {{ $jobConfig.image.pullPolicy | default "IfNotPresent" }}
        {{- if $jobConfig.command }}
        command:
          {{- toYaml $jobConfig.command | nindent 10 }}
        {{- else }}
        command:
          - {{ `{{ _command }}` }}
        {{- end }}
        {{- if $jobConfig.args }}
        args:
          {{- toYaml $jobConfig.args | nindent 10 }}
        {{- else }}
        args:
        {{ `{% for arg in _args %}` }}
          - {{ `{{ arg }}` }}
        {{ `{% endfor %}` }}
        {{- end }}
        {{- if $jobConfig.containerSecurityContext }}
        securityContext:
          {{- toYaml $jobConfig.containerSecurityContext | nindent 10 }}
        {{- end }}
        env:
          {{- with $jobConfig.env }}
          {{- toYaml . | nindent 10 }}
          {{- end }}
          - name: API_URL
            value: {{ printf "http://%s:%.0f" $root.Values.apiSettings.service.name (float64 $root.Values.apiSettings.service.port) | quote }}
          - name: FIFTYONE_DATABASE_ADMIN
            value: "false"
          - name: FIFTYONE_DATABASE_NAME
            valueFrom:
              secretKeyRef:
                name: {{ $root.Values.secret.name }}
                key: fiftyoneDatabaseName
          - name: FIFTYONE_DATABASE_URI
            valueFrom:
              secretKeyRef:
                name: {{ $root.Values.secret.name }}
                key: mongodbConnectionString
          - name: FIFTYONE_ENCRYPTION_KEY
            valueFrom:
              secretKeyRef:
                name: {{ $root.Values.secret.name }}
                key: encryptionKey
        {{- include "fiftyone-teams-app.mergeResources" (dict "job" $jobConfig "profile" $profile) | nindent 8 }}
        {{- with $jobConfig.volumeMounts }}
        volumeMounts:
          {{- toYaml . | nindent 10 }}
        {{- end }}
      {{- with $jobConfig.volumes }}
      volumes:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end -}}

{{/*
Merge resources from profile and job configuration
*/}}
{{- define "fiftyone-teams-app.mergeResources" -}}
{{- $limits := dict -}}
{{- $requests := dict -}}

{{- if .profile.resources -}}
  {{- if .profile.resources.limits -}}
    {{- range $key, $value := .profile.resources.limits -}}
      {{- if contains "{{ .gpu.count }}" $value -}}
        {{- $_ := set $limits $key ($.job.gpu.count | toString) -}}
      {{- else -}}
        {{- $_ := set $limits $key $value -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}

  {{- if .profile.resources.requests -}}
    {{- range $key, $value := .profile.resources.requests -}}
      {{- if contains "{{ .gpu.count }}" $value -}}
        {{- $_ := set $requests $key ($.job.gpu.count | toString) -}}
      {{- else -}}
        {{- $_ := set $requests $key $value -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- if .job.resources -}}
  {{- if .job.resources.limits -}}
    {{- range $key, $value := .job.resources.limits -}}
      {{- $_ := set $limits $key $value -}}
    {{- end -}}
  {{- end -}}

  {{- if .job.resources.requests -}}
    {{- range $key, $value := .job.resources.requests -}}
      {{- $_ := set $requests $key $value -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- if or $limits $requests -}}
resources:
  {{- if $limits }}
  limits:
    {{- range $key, $value := $limits }}
    {{ $key }}: {{ $value | quote }}
    {{- end }}
  {{- end }}
  {{- if $requests }}
  requests:
    {{- range $key, $value := $requests }}
    {{ $key }}: {{ $value | quote }}
    {{- end }}
  {{- end }}
{{- end -}}
{{- end -}}
