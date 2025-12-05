<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">
<img alt="Voxel51 Logo" src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img alt="Voxel51 FiftyOne" src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">
</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# Kubernetes On-Demand Orchestrator Setup

<!-- toc -->

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Kubernetes Credentials](#kubernetes-credentials)
- [Create Job Template](#create-job-template)
- [Container Image](#container-image)
  - [Required Environment Variables](#required-environment-variables)
- [Register Orchestrator in FiftyOne](#register-orchestrator-in-fiftyone)
  - [Configuration Options](#configuration-options)
  - [Template Storage Options](#template-storage-options)
  - [Secrets Options](#secrets-options)
- [Separate CPU and GPU Templates](#separate-cpu-and-gpu-templates)
- [Refresh Orchestrator Operators](#refresh-orchestrator-operators)
- [Additional Considerations](#additional-considerations)
- [Credential Rotation](#credential-rotation)
- [Full Production Template Example](#full-production-template-example)

<!-- tocstop -->

This document provides a step-by-step guide to configuring FiftyOne Enterprise
to use [Kubernetes](https://kubernetes.io/) as an orchestrator for running
delegated operations on-demand.

## Introduction

This document outlines the steps necessary to configure your FiftyOne
Enterprise system to send Delegated Operations to your Kubernetes cluster
for execution, on-demand. Jobs are submitted as Kubernetes
[Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
using a Jinja2 template that you provide.

## Prerequisites

Your FiftyOne API deployment must have the `kubernetes` Python package
installed. This is not included by default, so you will need to add it as an
extra dependency. See the
[Custom Plugins Images docs](../custom-plugins.md#custom-plugins-images).

## Kubernetes Credentials

The orchestrator can authenticate to Kubernetes in two ways:

1. **In-cluster credentials** (recommended for production): If the FiftyOne
   API is running inside the same Kubernetes cluster, leave `kubeConfig` empty
   and it will automatically use the service account credentials or from a kube
   config file that is present on the API pod.

2. **Kubeconfig file**: Provide the contents of a kubeconfig file as a string.
   This is useful for connecting to remote clusters or for testing.

## Create Job Template

The Kubernetes orchestrator uses a Jinja2 template to generate Job manifests.
This gives you full control over the Job spec, including resource requests,
node selectors, tolerations, volumes, and environment variables.

The template must contain the following variables that are replaced by the API
at runtime:

| Variable   | Description                                                                                                                                  |
|------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `_id`      | Task ID                                                                                                                                      |
| `_name`    | Generated job name                                                                                                                           |
| `_image`   | Container image (only if you prefer setting the image in the orchestrator's configuration, otherwise set the image directly in the template) |
| `_command` | Command to run                                                                                                                               |
| `_args`    | Arguments for the command                                                                                                                    |

Here is a minimal example template:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ _name }}
  namespace: fiftyone
  labels:
    task-id: {{ _id }}
spec:
  ttlSecondsAfterFinished: 60
  backoffLimit: 0
  template:
    spec:
      containers:
      - name: task-worker
        image: registry/image:tag
        command:
          - {{ _command }}
        args:
        {% for arg in _args %}
          - {{ arg }}
        {% endfor %}
        env:
          - name: API_URL
            value: "http://teams-api.fiftyone.svc.cluster.local:8000"
          - name: FIFTYONE_ENCRYPTION_KEY
            valueFrom:
              secretKeyRef:
                name: fiftyone-secrets
                key: encryption-key
          - name: FIFTYONE_INTERNAL_SERVICE
            value: "1"
          - name: FIFTYONE_DATABASE_URI
            valueFrom:
              secretKeyRef:
                name: fiftyone-secrets
                key: database-uri
      restartPolicy: Never
```

See the
[Full Production Template Example](#full-production-template-example)
at the end of this document for a complete setup with volumes, security
contexts, and GCS FUSE.

## Container Image

You need a container image with FiftyOne installed that will run your
delegated operations. This image should include:

1. FiftyOne Enterprise Python package
1. Any additional dependencies required by your operators
1. Custom operators (if not using a plugins directory)
1. Pushed to a container registry accessible by your Kubernetes cluster

### Required Environment Variables

Your job template must set the following environment variables for the
delegated operation to connect back to FiftyOne:

| Variable                    | Description                                                    |
|-----------------------------|----------------------------------------------------------------|
| `API_URL`                   | URL of the FiftyOne Teams API (must be reachable from the pod) |
| `FIFTYONE_DATABASE_URI`     | MongoDB connection URI                                         |
| `FIFTYONE_ENCRYPTION_KEY`   | FiftyOne encryption key                                        |
| `FIFTYONE_INTERNAL_SERVICE` | Set to `1`                                                     |

For cloud storage access, you will also need to configure the appropriate
credentials (e.g., `GOOGLE_APPLICATION_CREDENTIALS` for GCP,
`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` for AWS) or ensure the delegated
operator pod's service account is permitted to cloud storage.

## Register Orchestrator in FiftyOne

To register your orchestrator with FiftyOne, you may use the
[FiftyOne Management SDK](https://docs.voxel51.com/enterprise/management_sdk.html#module-fiftyone.management.orchestrator).
Supply the environment you want to run your orchestrator
(`fom.OrchestratorEnvironment.KUBERNETES`), the configuration, and
credential to access that runner. To use the FiftyOne
Management SDK, set the `API_URI` environment variable or
FiftyOne configuration variable.

When registering your orchestrator with FiftyOne, supply the
credential information stored as a
[FiftyOne Secret](https://docs.voxel51.com/enterprise/secrets.html).
The `secrets` parameter to
[`fom.register_orchestrator()`](https://docs.voxel51.com/enterprise/management_sdk.html#fiftyone.management.orchestrator.register_orchestrator)
takes a top-level key that must match your orchestrator environment. The
object that follows has key and value pairs specific to the
credentials needed to access your orchestrator.

When supplying one of the values, a new secret will be created for you that
securely stores the information provided. These can be managed as
[FiftyOne Secrets](https://docs.voxel51.com/enterprise/secrets.html).

Optionally, if you have an existing secret containing the credentials,
provide that secret name, and it will be used
instead of creating a new one. Examples of both options are below.

Example snippet using the Management SDK to register a Kubernetes orchestrator:

```python
import fiftyone.management as fom

fom.register_orchestrator(
    instance_id="kubernetes-gpu",
    description="Kubernetes GPU cluster for ML operations",
    environment=fom.OrchestratorEnvironment.KUBERNETES,
    config={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "image": "your-registry/fiftyone-worker:latest", # optional, should generally be set in template
            "executionTmplUri": "/path/to/gpu-job-template.yaml.j2",
            "registrationTmplUri": "/path/to/cpu-job-template.yaml.j2",  # optional
            "namespace": "fiftyone",  # optional, can also be set in template or from the kube config
            "context": "my-cluster-context",  # optional, will use the default context otherwise
        }
    },
    secrets={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "kubeConfig": "",  # optional, absent or empty for in-cluster auth
        }
    },
)
```

This will register a new orchestrator with the identifier `kubernetes-gpu`.

Additionally, it will save a new secret for the value supplied in kubeConfig.
That new secret will have the name `KUBE_CONFIG_KUBERNETES_GPU`.

If you already have a secret with values,
supply the name in the `secrets` parameter.
Here is an example:

```python
import fiftyone.management as fom

fom.register_orchestrator(
    instance_id="kubernetes-gpu",
    description="Kubernetes GPU cluster for ML operations",
    environment=fom.OrchestratorEnvironment.KUBERNETES,
    config={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "executionTmplUri": "/path/to/gpu-job-template.yaml.j2",
        }
    },
    secrets={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "kubeConfig": "EXISTING_KUBECONFIG_SECRET",
        }
    },
)
```

In this case, a new secret will not be created.
The existing secrets will be associated with the orchestrator.

### Configuration Options

| Parameter             | Required | Description                                                             |
|-----------------------|----------|-------------------------------------------------------------------------|
| `image`               | No       | Container image to use for jobs, generally set in template              |
| `executionTmplUri`    | Yes      | Path/URI to the Job template or base64 encoded (see [below](#b64))      |
| `registrationTmplUri` | No       | Path/URI to a separate template for registration jobs or base64 encoded |
| `namespace`           | No       | Kubernetes namespace (can also be set in template or kubeconfig)        |
| `context`             | No       | Kubeconfig context to use if different from default                     |

### Template Storage Options

The template can be provided in one of the following ways:

- **File path accessible to the API**: Store the template file somewhere the
   FiftyOne API can read it (local filesystem, mounted volume, etc.) and
   provide the path. The templates can be mounted to the API pod via configmaps.

   ```python
   # ...
   "executionTmplUri": "/path/to/template.yaml.j2"
   # ...
   ```

- **Cloud storage URI**: Store the template in GCS, S3, or other supported
   storage and provide the URI.

   ```python
   # ...
   "executionTmplUri": "gs://my-bucket/templates/job-template.yaml.j2"
   # ...
   ```

<!-- markdownlint-disable no-inline-html line-length -->
- <a id="b64"></a>**Base64-encoded data URI**: Embed the template's
   content directly in the config as a base64-encoded string.

   ```python
   import base64

   template = """..."""  # your full template as a string here, or loaded from disk
   encoded = base64.b64encode(template.encode()).decode()

   # ...
   "executionTmplUri": f"data:text/yaml;base64,{encoded}"
   # ...
   ```

### Secrets Options

| Parameter    | Required | Description                                                          |
|--------------|----------|----------------------------------------------------------------------|
| `kubeConfig` | No       | Kubeconfig file contents as string. Leave empty for in-cluster auth. |

## Separate CPU and GPU Templates

A common pattern is to register two orchestrators: one for GPU-heavy ML
operations and one for lightweight CPU tasks (including operator registration).
This avoids wasting expensive GPU resources on simple jobs.

**GPU template** (`gpu-job-template.yaml.j2`):

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ _name }}
  namespace: fiftyone
  labels:
    task-id: {{ _id }}
    task-type: delegated_operation
spec:
  ttlSecondsAfterFinished: 60
  backoffLimit: 0
  template:
    metadata:
      labels:
        task-id: {{ _id }}
        task-type: delegated_operation
    spec:
      containers:
      - name: task-worker
        image: registry/image:tag
        command:
          - {{ _command }}
        args:
        {% for arg in _args %}
          - {{ arg }}
        {% endfor %}
        env:
          - name: API_URL
            value: "http://teams-api.fiftyone.svc.cluster.local:8000"
          - name: FIFTYONE_ENCRYPTION_KEY
            valueFrom:
              secretKeyRef:
                name: fiftyone-secrets
                key: encryption-key
          - name: FIFTYONE_INTERNAL_SERVICE
            value: "1"
          - name: FIFTYONE_DATABASE_URI
            valueFrom:
              secretKeyRef:
                name: fiftyone-secrets
                key: database-uri
        resources:
          requests:
            cpu: "4"
            memory: "16Gi"
            nvidia.com/gpu: "1"
          limits:
            cpu: "8"
            memory: "32Gi"
            nvidia.com/gpu: "1"
      nodeSelector:
        cloud.google.com/gke-accelerator: nvidia-tesla-t4
      tolerations:
        - key: nvidia.com/gpu
          operator: Exists
          effect: NoSchedule
      restartPolicy: Never
```

**CPU template** (`cpu-job-template.yaml.j2`):

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ _name }}
  namespace: fiftyone
  labels:
    task-id: {{ _id }}
    task-type: delegated_operation
spec:
  ttlSecondsAfterFinished: 60
  backoffLimit: 0
  template:
    metadata:
      labels:
        task-id: {{ _id }}
        task-type: delegated_operation
    spec:
      containers:
      - name: task-worker
        image: registry/image:tag
        command:
          - {{ _command }}
        args:
        {% for arg in _args %}
          - {{ arg }}
        {% endfor %}
        env:
          - name: API_URL
            value: "http://teams-api.fiftyone.svc.cluster.local:8000"
          - name: FIFTYONE_ENCRYPTION_KEY
            valueFrom:
              secretKeyRef:
                name: fiftyone-secrets
                key: encryption-key
          - name: FIFTYONE_INTERNAL_SERVICE
            value: "1"
          - name: FIFTYONE_DATABASE_URI
            valueFrom:
              secretKeyRef:
                name: fiftyone-secrets
                key: database-uri
        resources:
          requests:
            cpu: "1"
            memory: "2Gi"
          limits:
            cpu: "2"
            memory: "4Gi"
      restartPolicy: Never
```

Register the GPU orchestrator with the CPU template for registration:

```python
import fiftyone.management as fom

fom.register_orchestrator(
    instance_id="kubernetes-gpu",
    description="Kubernetes GPU cluster for ML operations",
    environment=fom.OrchestratorEnvironment.KUBERNETES,
    config={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "executionTmplUri": "/templates/gpu-job-template.yaml.j2",
            "registrationTmplUri": "/templates/cpu-job-template.yaml.j2",
        }
    },
    secrets={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "kubeConfig": "",
        }
    },
)
```

You may also register a separate CPU-only orchestrator for operations that
do not require GPU:

```python
fom.register_orchestrator(
    instance_id="kubernetes-cpu",
    description="Kubernetes CPU cluster for lightweight operations",
    environment=fom.OrchestratorEnvironment.KUBERNETES,
    config={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "executionTmplUri": "/templates/cpu-job-template.yaml.j2",
        }
    },
    secrets={
        fom.OrchestratorEnvironment.KUBERNETES: {
            "kubeConfig": "",
        }
    },
)
```

## Refresh Orchestrator Operators

This step is only required if you've added a plugin directory with custom
plugins to your Kubernetes environment.

Once your orchestrator is registered in FiftyOne you may now refresh the
available operators for that environment. To do so:

1. Go to any dataset/runs page and select your orchestrator on the right-hand side.
1. Select the "refresh" button and click "confirm" when prompted.
    - This will kick off a job in your Kubernetes cluster that will tell
       FiftyOne what operators are available in that environment.
1. Once you see the job is complete, reload the page and verify your
   "available operators" show the ones that you have configured.

In the future, anytime you add new operators to your environment, you will go
through this same workflow.

## Additional Considerations

Your Kubernetes service account (or the credentials in kubeConfig) will need
appropriate RBAC permissions to create and delete jobs.

For cloud storage access, you may need to configure:

- Storage Bucket Viewer
- Storage Object Viewer
- Write permissions, if you setup
  [cloud storage logging](https://docs.voxel51.com/enterprise/plugins.html#logs)
- Blob sign permission, if the plugin uses signed URLs and your cloud platform
  requires additional permissions

Additionally:

- The `ttlSecondsAfterFinished` setting in your Job spec controls how long
  completed jobs persist before being cleaned up. A short value (60s)
  keeps the cluster tidy while a longer value makes debugging easier.

## Credential Rotation

If you are using a kubeConfig secret and need to rotate credentials:

```python
import fiftyone.management as fom

orc = fom.get_orchestrator("kubernetes-gpu")
fom.update_secret(
    key=orc.secrets['kube_config'],
    value="<new_kubeconfig_contents>",
)
```

For in-cluster authentication, credential rotation is handled by your
Kubernetes cluster's service account management.

## Full Production Template Example

The following template shows a complete production setup with GCS FUSE volumes,
security contexts, resource limits, and all recommended environment variables:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ _name }}
  namespace: your-org-fiftyone-ai
  labels:
    task-id: {{ _id }}
    task-type: delegated_operation
spec:
  ttlSecondsAfterFinished: 60
  backoffLimit: 0
  template:
    metadata:
      labels:
        task-id: {{ _id }}
        task-type: delegated_operation
      annotations:
          gke-gcsfuse/volumes: 'true'
    spec:
      serviceAccountName: your-org-fiftyone-teams
      podSecurityContext:
          runAsNonRoot: false
      containers:
      - name: task-worker
        image: registry/image:tag
        command:
          - {{ _command }}
        args:
        {% for arg in _args %}
          - {{ arg }}
        {% endfor %}
        env:
          - name: API_URL
            value: http://teams-api:80
          - name: FIFTYONE_ENCRYPTION_KEY
            valueFrom:
              secretKeyRef:
                key: encryptionKey
                name: your-org-teams-secrets
          - name: FIFTYONE_INTERNAL_SERVICE
            value: "1"
          - name: FIFTYONE_DATABASE_URI
            valueFrom:
              secretKeyRef:
                key: mongodbConnectionString
                name: your-org-teams-secrets
          - name: FIFTYONE_DATABASE_NAME
            valueFrom:
              secretKeyRef:
                key: fiftyoneDatabaseName
                name: your-org-teams-secrets
          - name: FIFTYONE_DATABASE_ADMIN
            value: "false"
          - name: FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH
            value: gs://bucket/name/path
          - name: FIFTYONE_FEATURE_FLAG_ENABLE_GEN_LABELING
            value: "true"
          - name: FIFTYONE_MEDIA_CACHE_DIR
            value: /opt/media_cache
          - name: FIFTYONE_MEDIA_CACHE_SIZE_BYTES
            value: "2147483648"
          - name: FIFTYONE_MODEL_ZOO_DIR
            value: /opt/fiftyone_zoo/your-org/models
          - name: FIFTYONE_PLUGINS_CACHE_ENABLED
            value: "true"
          - name: FIFTYONE_PLUGINS_DIR
            value: /opt/plugins
          - name: NUMBA_CACHE_DIR
            value: /tmp/numba
          - name: TORCH_HOME
            value: /opt/fiftyone_zoo/your-org/torch
        resources:
          limits:
            cpu: "1"
            memory: 6656Mi
          requests:
            cpu: "1"
            memory: 6656Mi
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
        volumeMounts:
          - mountPath: /opt/fiftyone
            name: opt-fiftyone
          - mountPath: /opt/.cache
            name: opt-dot-cache
          - mountPath: /opt/plugins
            name: nfs-plugins-ro-vol
          - mountPath: /opt/.fiftyone
            name: fiftyone-home-vol
          - mountPath: /opt/.config
            name: matplotlib-config-vol
          - mountPath: /tmp
            name: tmpdir
          - mountPath: /opt/fiftyone_zoo
            name: fuse-do-models-vol
          - mountPath: /opt/media_cache
            name: memory-media-cache-vol
          - mountPath: /dev/shm
            name: shm-vol
      volumes:
        - name: nfs-plugins-ro-vol
          persistentVolumeClaim:
            claimName: your-org-pvc
            readOnly: true
        - emptyDir:
            sizeLimit: 10Mi
          name: opt-fiftyone
        - emptyDir:
            sizeLimit: 500Mi
          name: opt-dot-cache
        - emptyDir:
            sizeLimit: 10Mi
          name: fiftyone-home-vol
        - emptyDir:
            sizeLimit: 10Mi
          name: matplotlib-config-vol
        - emptyDir:
            sizeLimit: 10Mi
          name: tmpdir
        - csi:
            driver: gcsfuse.csi.storage.gke.io
            volumeAttributes:
              bucketName: your-bucket-name
              mountOptions: implicit-dirs,file-mode=640,dir-mode=750,uid=1000,gid=1000
          name: fuse-do-models-vol
        - emptyDir:
            medium: Memory
            sizeLimit: 2.5Gi
          name: memory-media-cache-vol
        - emptyDir:
            medium: Memory
            sizeLimit: 2Gi
          name: shm-vol
      restartPolicy: Never
```
