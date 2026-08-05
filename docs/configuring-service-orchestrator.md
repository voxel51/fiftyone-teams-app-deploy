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

Both services need a GPU.
The default configurations
(GPU reservation on Compose, and `nvidia.com/gpu: 1` on Kubernetes)
express a device count and cannot specify VRAM requirements,
so neither prevents a service from starting on an accelerator too small
for its model.
To prevent GPU out of memory errors,
please note the following minimum recommended GPU requirements:

| Service | Minimum accelerator | Notes |
| --- | --- | --- |
| `annotation-ai` | 16 GB VRAM (T4, L4) | SAM2 is a small model. |
| `agentic-labeler` | 24 GB VRAM, Ampere or newer (L4, A10G, L40S, A100) | The default `gemma4-31B-qat-maxvision` config is roughly 17-20 GB of int4 weights before KV cache, and vLLM's int4 kernels require compute capability 8.0+, so a T4 is unsuitable on both counts. |

Host memory matters as much as VRAM,
because weights are read into host memory before they reach the device.
On Kubernetes the chart sets no cpu or memory request on the
orchestrator, so
[set them for the model you run](#gpu-requirements).
On Compose, size the host itself; see the
[worker requirements](../docker/docs/configuring-agentic-labeler.md#requirements).

## Docker Compose

### The builtin services list

The deployment's services are declared in
[builtin_services.yaml](../docker/builtin_services.yaml),
which `common-services.yaml` mounts into `teams-api` and which
`teams-api` reconciles at startup.
Entries deep-merge by `id` onto the definitions packaged in `fiftyone`,
so that file is the deployment's full list.
Add, remove, or retarget services by editing it.

Bump an entry's `builtin_version` to re-apply a change to an environment
that has already stored that service.
Without the bump, reconciliation keeps the stored definition.

`entrypoint.container.port` is the port the `teams-api` `/service` proxy
dials.
It is required even on Compose,
where no container is provisioned and the shell command is what runs.

### Where each service runs

A service's `delegation_target` names the worker that hosts it,
and the two builtin services target different workers by default:

| Service | `delegation_target` | Worker |
| --- | --- | --- |
| `annotation-ai` | `builtin` | The default `teams-do` worker |
| `agentic-labeler` | `agentic-labeler` | The dedicated GPU worker added by [`compose.agenticlabeler.yaml`](../docker/docs/configuring-agentic-labeler.md) |

`annotation-ai` therefore lands on the default `teams-do` worker, which
needs GPU access before the service can start.
Either give that worker GPU access, see
[configuring GPU workloads](../docker/docs/configuring-gpu-workloads.md),
or retarget the service by pointing its `delegation_target` at a worker
that already has one.

This differs from Kubernetes,
where both services target the chart's `gpuServiceOrc` and so both get a
GPU by default.

### `FIFTYONE_SERVICE_POD_IP`

The worker hosts the service in-process and publishes the address the
`teams-api` proxy uses to reach it.
It auto-detects its own container IP at runtime,
which is correct on a standard single-network Compose host.
On multi-homed or non-default-network hosts,
where auto-detect can pick the wrong interface,
set `FIFTYONE_SERVICE_POD_IP` on the worker to the reachable address.
This mirrors the Kubernetes path,
where the resolver reads the injected `POD_IP`.

## Kubernetes

### GPU requirements

The chart ships two service orchestrators,
`cpuServiceOrc` and `gpuServiceOrc`,
under `delegatedOperatorJobTemplates.serviceOrchestrators`.
Both are registered on every `helm install` and `helm upgrade`,
so both builtin services appear under `Settings -> Services`
whether or not the cluster has GPU nodes.
Neither starts on its own.
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
This is enough for AKS GPU node pools,
EKS,
GKE Standard with an existing GPU node pool,
and on-premises GPU Operator installs.

However, some cases require more than the generic request.
For example:

- **GKE Autopilot** rejects GPU pods that do not set
  `cloud.google.com/gke-accelerator`.
- **Scaling from zero**,
  including GKE node auto-provisioning,
  needs the accelerator label to decide which node pool to grow.
- **Mixed-accelerator clusters** cannot be steered by a device count.
  Pin the accelerator with a `nodeSelector` to meet the
  [minimums above](#accelerator-sizing).

The chart also sets no cpu or memory request on `gpuServiceOrc`,
so a service pod is placed on GPU availability alone.
Consider setting both:
a pod with no memory request is an early eviction candidate under node
memory pressure,
the cluster autoscaler has nothing to size a node from when scaling up
for a service,
and a namespace with a `LimitRange` or `ResourceQuota` that requires
requests will reject the pod outright.

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      resources:
        requests:
          cpu: 2        # Modify For Your Needs
          memory: 12Gi  # Modify For Your Needs
```

Leaving the memory limit unset avoids OOMKilling a model server that
grows past its request.
Both services share `gpuServiceOrc`,
so one `resources` block covers whichever service is running.
To size them independently,
declare a second orchestrator and move one service's entry under it.

### Targeting specific GPU nodes

Add the cloud-specific settings to the shipped `gpuServiceOrc` rather
than declaring a replacement,
so the `services` entries are inherited instead of restated.
Per-cloud examples are in
[Leveraging GPU Workloads](../helm/docs/configuring-gpu-workloads.md).

Overrides follow the same rules as the rest of
`delegatedOperatorJobTemplates`:
maps merge key-wise and lists are replaced wholesale, see
[merge examples](../helm/docs/configuring-delegated-operators.md#examples).
For GPU settings that means:

- Adding `cloud.google.com/gke-accelerator` to `nodeSelector` merges with
  the shipped value.
- Supplying `tolerations` replaces the chart's `nvidia.com/gpu`
  toleration entirely.
  Restate it if the cluster still needs it.
- `nvidia.com/gpu` survives a partial `resources` override.
  Removing it requires setting it to `null` explicitly.

### CPU-only clusters

Starting a GPU service where no node advertises `nvidia.com/gpu` leaves
the pod `Pending` until `FIFTYONE_SERVICE_POD_READY_TIMEOUT_S` expires,
30 minutes by default,
before the service is marked with an error.
On clusters that will never have GPU nodes,
stop registering the orchestrator,
and delete them from the Settings -> Orchestrators UI in FiftyOne Enterprise:

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      enabled: false
```

This drops the pod template and the orchestrator registration.
Services that `teams-api` has already reconciled into the deployment
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
