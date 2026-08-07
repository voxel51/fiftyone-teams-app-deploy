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

Configuration is deployment-specific:

- [Docker Compose](#docker-compose)
- [Kubernetes](#kubernetes)

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
| `agentic-labeler` | 24 GB VRAM, Ampere or newer (L4, A10G, L40S, A100) | The default `gemma4-31B-qat-maxvision` config is ~17-20 GB of int4 weights before KV cache, and vLLM's int4 kernels require compute capability 8.0+. |

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

The `teams-api`'s `/service` proxy connects to the builtin service's `entrypoint.container.port`.
This is required even when `entrypoint.kind=shell`.

### Where each service runs

A service's `delegation_target` sets the host worker.
The builtin services target different workers by default:

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

On a single-network Compose host, the service publishes its
(auto-detected) IP address to which the `teams-api` proxy connects.
On multi-homed or non-default-network hosts
(where auto-detect can pick the wrong interface),
set the `FIFTYONE_SERVICE_POD_IP` to the reachable address.

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
- `nvidia.com/gpu` survives a partial `resources` override.
  - To remove it, set it to `null`.

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
