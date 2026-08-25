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

# Configuring Service Orchestrators

Some builtin operators run as long-lived services (always-on model servers)
rather than on-demand delegated jobs. A service orchestrator is the
delegated-operator worker that hosts these services and keeps them reachable
from `teams-api`.

Services are declared in the deployment's builtin services list, which
`teams-api` mounts and reconciles at startup. Each ships created stopped. Start
one from the FiftyOne Enterprise UI under `Settings -> Services`.

The builtin services are:

- `annotation-ai` (SAM2) powers AI-assisted segmentation in the annotation
  editor. Needs a GPU.
- `agentic-labeler` powers few-shot VLM labeling. Needs a GPU.

## Accelerator sizing

The builtin services need a GPU.
The default configurations (Compose GPU reservation and Kubernetes
`nvidia.com/gpu: 1`) define device count but cannot specify VRAM requirements.
The resulting accelerator may be too small for its model.
To prevent GPU out of memory errors, please specify
an accelerator that exceeds the minimum recommended requirements:

| Service | Minimum accelerator | Notes |
| --- | --- | --- |
| `annotation-ai` | 16 GB VRAM (T4, L4) | SAM2 is a small model. |
| `agentic-labeler` | 48 GB VRAM, Ampere or newer (L40S, A100, H100), or 2 x 24 GB (L4, A10G) | The pinned `gemma4-31B-qat-maxvision` config is ~22 GB of int4 weights before KV cache, and vLLM's int4 kernels require compute capability 8.0+. |

A single 24 GB card cannot run the pinned default: the weights alone leave
under 2 GB for the KV cache and activations.
On two 24 GB cards the launcher shards the model automatically.
To run the Agentic Labeler on one 24 GB card, point `LABELER_CONFIG_FILE` at a
smaller config; see
[choosing a model](../docker/docs/configuring-agentic-labeler.md#choosing-a-model).

Host memory matters as much as VRAM.
The weights are read into host memory before they reach the device.
On Kubernetes the chart sets no cpu or memory request on the orchestrator, so
[set them for the model you run](#gpu-requirements).
On Docker Compose, size the host itself; see the
[worker requirements](../docker/docs/configuring-agentic-labeler.md#requirements).

## Docker Compose

### The builtin services list

The builtin services are declared in
[docker/builtin_services.yaml](../docker/builtin_services.yaml)
and bind mounted into the `teams-api` service
(see [docker/common-services.yaml](../docker/common-services.yaml)).
At startup, the `teams-api` container performs a reconciliation where
entries are deep-merged by `id` onto the definitions packaged in `fiftyone`.
Edit
[docker/builtin_services.yaml](../docker/builtin_services.yaml)
to add, remove, or retarget services.

To update a builtin service's definition, you must
increment the service's `builtin_version`.
Otherwise the cached definition will not be updated.

A version bump re-applies the **entire** entry, not just the fields you
changed.
Any value set through the UI rather than the file is overwritten —
`delegation_target` in particular.
Write the current UI values into the file before incrementing.

### Service ports

Two ports are configured per service, and they are not interchangeable:

| Field | Dialed by | Must be |
| --- | --- | --- |
| `entrypoint.container.port` | The `teams-api` `/service` proxy, over the network | The address the proxy can reach the service on. Required even when `entrypoint.kind=shell`. |
| `entrypoint.container.healthcheck.port` | The health probe, on `127.0.0.1` from inside the worker | The port the service listens on **in the container**. |

On a single-host deployment these are usually the same number.
When `teams-api` and the worker are on different hosts they are not:
`container.port` is the port published on the worker's host, while
`healthcheck.port` is still the in-container port.

Setting `healthcheck.port` to a published host port is a common mistake.
Nothing is listening on that port inside the container, so the service is
never marked healthy and the Services page reports `STARTING` indefinitely
with no error in any log.

For a worked split-host example, see
[running `teams-api` and the worker on separate hosts](../docker/docs/configuring-agentic-labeler.md#running-teams-api-and-the-worker-on-separate-hosts).

### Where each service runs

A service's `delegation_target` specifies the worker,
and the builtin services have different targets by default:

| Service | `delegation_target` | Worker |
| --- | --- | --- |
| `annotation-ai` | `builtin` | The default `teams-do` worker |
| `agentic-labeler` | `agentic-labeler` | The dedicated GPU worker added by [`compose.agenticlabeler.yaml`](../docker/docs/configuring-agentic-labeler.md) |

`annotation-ai` therefore lands on the default `teams-do` worker, which
needs a GPU (before the service can start).
Either give that worker a GPU (see
[configuring GPU workloads](../docker/docs/configuring-gpu-workloads.md))
or retarget the service by pointing its `delegation_target` at a GPU worker.

### `FIFTYONE_SERVICE_POD_IP`

The service attempts to auto-detect its own container IP at runtime,
which is reliable on a standard single-network Compose host.
On multi-homed or non-default-network hosts,
where auto-detect can pick the wrong interface,
set `FIFTYONE_SERVICE_POD_IP` on the worker to the reachable address.

## Kubernetes

### GPU requirements

The chart provides two service orchestrators (`cpuServiceOrc` and `gpuServiceOrc`)
under `delegatedOperatorJobTemplates.serviceOrchestrators`.
Both service orchestrators are registered on every
`helm install` and `helm upgrade` invocation.
Both builtin services appear under `Settings -> Services`
(even when the cluster has no GPU nodes).
Neither service automatically starts.
For the values structure, see
[`serviceOrchestrators`](../helm/docs/configuring-delegated-operators.md#long-lived-services-with-serviceorchestrators)
and the default service specs in
[`values.yaml`](../helm/fiftyone-teams-app/values.yaml)
under `serviceOrchestrators.services`.

`gpuServiceOrc` hosts both services.
It requests one GPU through the standard Kubernetes extended resource and
tolerates the taint the NVIDIA device plugin and GPU Operator apply:

```yaml
resources:
  limits:
    nvidia.com/gpu: 1
  requests:
    nvidia.com/gpu: 1
tolerations:
  - key: nvidia.com/gpu
    effect: NoSchedule
    operator: Exists
```

There is deliberately no `nodeSelector`,
so the pod schedules onto any node advertising `nvidia.com/gpu`.
The cluster must already have GPU nodes running the NVIDIA driver
and either the device plugin or the GPU Operator to advertise that
resource.
This supports AKS GPU node pools, EKS, GKE Standard (with
an existing GPU node pool), and on-premises GPU Operator.

Some cases may require configurations beyond the generic request.
For example:

- **GKE Autopilot** rejects GPU pods that do not contain
  `nodeSelector.cloud.google.com/gke-accelerator`.
- **Scaling from zero**
  (including GKE node auto-provisioning)
  needs the accelerator label to decide which node pool to grow.
- **Mixed-accelerator clusters** cannot be steered by a device count.
  Set the accelerator with a `nodeSelector` that meets the
  [Accelerator sizing](#accelerator-sizing) minimums.

The chart does not set no cpu or memory request on `gpuServiceOrc`.
so a service pod is placed on GPU availability alone.
We recommend setting both, because:

- A pod with no memory request is an early eviction candidate
  under node memory pressure.
- The cluster autoscaler has nothing to size a node from
  when scaling up for a service.
- A namespace with a `LimitRange` or `ResourceQuota` that requires
  requests will reject the pod outright.

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      resources:
        # Modify as needed
        requests:
          cpu: 2
          memory: 12Gi
```

Leaving the memory limit unset avoids OOMKilling a model server
that grows beyond its request.
By default, both builtin services share the `gpuServiceOrc` `resources` block.
To set service-specific values, declare a second
orchestrator and move one service's entry under it.

### Persisting the model cache

Starting `agentic-labeler` downloads the model weights, which for the
pinned config are tens of GB.
Without persistent storage that download repeats every time the pod is
recreated.

The chart does not ship a `PersistentVolumeClaim` for this, because the
storage class and access mode are deployment-specific.
Mount your own through the existing `volumes` and `volumeMounts` keys:

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      volumes:
        - name: labeler-cache
          persistentVolumeClaim:
            claimName: agentic-labeler-cache
      volumeMounts:
        - name: labeler-cache
          mountPath: /home/voxel51/.cache/huggingface
          subPath: huggingface
        - name: labeler-cache
          mountPath: /home/voxel51/.cache/torchinductor
          subPath: torchinductor
        - name: labeler-cache
          mountPath: /home/voxel51/.cache/triton
          subPath: triton
```

Mount each path the image declares as a `VOLUME` rather than their shared
parent; a mount on `/home/voxel51/.cache` alone is shadowed and has no
effect.
See
[persist the model cache](../docker/docs/configuring-agentic-labeler.md#persist-the-model-cache).

`podSecurityContext.fsGroup` is already `1000` on the delegated-operator
templates, so the mounted volume is writable by the service with no
further configuration.

### Targeting specific GPU nodes

`services` entries inherit from `gpuServiceOrc`.
Set your cloud-specific GPU settings in `gpuServiceOrc`.
For cloud-specific examples, see
[Leveraging GPU Workloads](../helm/docs/configuring-gpu-workloads.md).

Overrides follow the same rules as the rest of
`delegatedOperatorJobTemplates`
(maps merge key-wise and lists are replaced wholesale, see
[merge examples](../helm/docs/configuring-delegated-operators.md#examples)).
For the GPU settings,

- Adding `cloud.google.com/gke-accelerator` to `nodeSelector` merges with
  the default values.
- Overriding `tolerations` replaces the chart default's `nvidia.com/gpu` toleration.
  - If your cluster requires this toleration, restate it.

### CPU-only clusters

When no Kubernetes node advertises `nvidia.com/gpu`, the service pod will be stuck
in a `Pending` state until the `FIFTYONE_SERVICE_POD_READY_TIMEOUT_S` duration
(30 minutes by default) expires.
After expiry, the service will be marked with an error.
For clusters without GPU nodes,
you may disable the service orchestrator registration.

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      enabled: false
```

When you disable the service orchestrator registration,
delete them in the FiftyOne Enterprise UI via *Settings* -> *Orchestrators*.

Services already reconciled by the `teams-api`
remain in the `Settings -> Services` list.

## Broker settings

The service broker runs inside `teams-api`. Set these variables on the
`teams-api` deployment (Kubernetes) or the `teams-api` service (Docker Compose)
to tune it.

| Variable | Default | Description |
| --- | --- | --- |
| `FIFTYONE_SERVICE_RECONCILE_DELAY_S` | `300` | Seconds after `teams-api` starts before it reconciles builtin services. Auto-start can queue heavy GPU work, so it is held off until the deployment settles. Set to `0` to reconcile as soon as the server starts. |
| `FIFTYONE_SERVICE_POD_READY_TIMEOUT_S` | `1800` (chart), `900` (broker) | Seconds the broker waits for a service pod to become ready before it times out and tears the pod down. A cold GPU start that provisions a node from zero can take several minutes, so raise this if pods are killed before they finish starting. Note that an unschedulable pod also consumes this full timeout. The chart sets `1800` in `apiSettings.env`, overriding the broker's own `900` default. Kubernetes service broker only. |
| `FIFTYONE_SERVICE_POD_READY_POLL_INTERVAL_S` | `2` | Seconds between pod readiness checks while the broker waits. Kubernetes service broker only. |
| `FIFTYONE_SERVICE_POD_SERVICE_ACCOUNT` | auto | ServiceAccount for service pods. Defaults to the `teams-api` ServiceAccount so pods inherit its Workload Identity and cloud access. Set to override. Kubernetes service broker only. |
