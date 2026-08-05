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

# Leveraging GPU Workloads

<!-- toc -->

- [Overview](#overview)
- [Utilizing GKE GPUs For Delegated Operations](#utilizing-gke-gpus-for-delegated-operations)
  - [Prerequisites](#prerequisites)
  - [Deploying GPU-enabled Delegated Operator Pods](#deploying-gpu-enabled-delegated-operator-pods)
  - [Deploying GPU-enabled On-Demand Jobs](#deploying-gpu-enabled-on-demand-jobs)
  - [Deploying GPU-enabled Service Orchestrators](#deploying-gpu-enabled-service-orchestrators)
- [Utilizing Azure AKS GPUs For Delegated Operations](#utilizing-azure-aks-gpus-for-delegated-operations)
  - [Prerequisites](#prerequisites-1)
  - [Deploying GPU-enabled Delegated Operator Pods](#deploying-gpu-enabled-delegated-operator-pods-1)
  - [Deploying GPU-enabled On-Demand Jobs](#deploying-gpu-enabled-on-demand-jobs-1)
  - [Deploying GPU-enabled Service Orchestrators](#deploying-gpu-enabled-service-orchestrators-1)
- [Utilizing AWS EKS GPUs For Delegated Operations](#utilizing-aws-eks-gpus-for-delegated-operations)
  - [Prerequisites](#prerequisites-2)
  - [Deploying GPU-enabled Delegated Operator Pods](#deploying-gpu-enabled-delegated-operator-pods-2)
  - [Deploying GPU-enabled On-Demand Jobs](#deploying-gpu-enabled-on-demand-jobs-2)
  - [Deploying GPU-enabled Service Orchestrators](#deploying-gpu-enabled-service-orchestrators-2)

<!-- tocstop -->

## Overview

Many machine learning applications utilize GPU hardware for
intensive computations.
The FiftyOne Enterprise helm chart allows users to schedule pods on
GPU-enabled nodes using the `nodeSelector`, `resource`, and `toleration`
settings for individual services.

The same three settings apply to always-running delegated operators,
on-demand jobs,
and
[service orchestrators](../../docs/configuring-service-orchestrator.md).
The chart ships a `gpuServiceOrc` service orchestrator that requests a
GPU generically,
with no cloud-specific `nodeSelector`;
the service orchestrator section for each cloud below covers what to add.

## Utilizing GKE GPUs For Delegated Operations

### Prerequisites

This example assumes you have a GKE cluster with GPU-available nodes.
Please refer to the
[autopilot documentation][gpu-gcp-gke-autopilot]
or the
[standard node pool documentation][gpu-gcp-gke-standard]
for assistance in setting up those clusters.

### Deploying GPU-enabled Delegated Operator Pods

We will configure the
[always-running delegated operators](./configuring-delegated-operators.md)
with GKE GPUs.

In your `values.yaml`,
under `.Values.delegatedOperatorDeployments.deployments`, add a new delegated
operator deployment.
Ensure the delegated operator deployment has `nodeSelector` set to valid
[GKE accelerator values and counts][gpu-gcp-gke-standard-multi].
The deployment should also set the `LD_LIBRARY_PATH` variable to the
corresponding
[google GPU driver][gpu-gcp-gke-standard-cuda].
Also be sure to modify the deployment's `resources.requests` to request
the desired amount of GPUs from the Kubernetes scheduler.

The below will deploy a CPU-based delegated operator (`teamsDo`) as well
as GPU-based delegated operators for both GKE auto-pilot mode
(`gpuGcpGkeAutopilot`) and standard GKE (`gpuGcpGkeStandard`):

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}  # A CPU Based Deployment
    gpuGcpGkeAutopilot:
      nodeSelector:
        cloud.google.com/gke-accelerator: nvidia-l4  # Modify For Your Needs
        cloud.google.com/gke-accelerator-count: "1"  # Modify For Your Needs
      resources:
        limits:
          cpu: 4        # Modify For Your Needs
          memory: 12Gi  # Modify For Your Needs
        requests:
          cpu: 4             # Modify For Your Needs
          memory: 12Gi       # Modify For Your Needs
          nvidia.com/gpu: 1  # Modify For Your Needs
      env:
        [...existing environment variables...]
        LD_LIBRARY_PATH: /usr/local/nvidia/lib64  # Modify For Your Needs
    gpuGcpGkeStandard:
      resources:
        limits:
          cpu: 4        # Modify For Your Needs
          memory: 12Gi  # Modify For Your Needs
        requests:
          cpu: 4             # Modify For Your Needs
          memory: 12Gi       # Modify For Your Needs
          nvidia.com/gpu: 1  # Modify For Your Needs
```

Upgrade your deployment via `helm upgrade` and wait for the
`gpu-gcp-gke-autopilot` or the `gpu-gcp-gke-standard`
pods to be scheduled and deployed.

### Deploying GPU-enabled On-Demand Jobs

We will configure the
[on-demand delegated operators](./configuring-delegated-operators.md)
with GKE GPUs.

In your `values.yaml`,
under `.Values.delegatedOperatorJobTemplates.jobs`, add a new delegated
operator job.
Ensure the delegated operator job has `nodeSelector` set to valid
[GKE accelerator values and counts][gpu-gcp-gke-standard-multi].
The deployment should also set the `LD_LIBRARY_PATH` variable to the
corresponding
[google GPU driver][gpu-gcp-gke-standard-cuda].
Also be sure to modify the deployment's `resources.requests` to request
the desired amount of GPUs from the Kubernetes scheduler.

The below will deploy a CPU-based delegated operator template
(`cpu-default`) as well as a GPU-based delegated operator
template (`gpu-gcp-gke-autopilot`):

```yaml
delegatedOperatorJobTemplates:
  jobs:
    cpuDefault: {}  # A CPU Based Job
    # https://docs.cloud.google.com/kubernetes-engine/docs/how-to/gpus
    gpuGcpGkeAutopilot:
      nodeSelector:
        cloud.google.com/gke-accelerator: nvidia-l4  # Modify For Your Needs
        cloud.google.com/gke-accelerator-count: "1"  # Modify For Your Needs
      resources:
        limits:
          cpu: 4        # Modify For Your Needs
          memory: 12Gi  # Modify For Your Needs
        requests:
          cpu: 4             # Modify For Your Needs
          memory: 12Gi       # Modify For Your Needs
          nvidia.com/gpu: 1  # Modify For Your Needs
      env:
        [...existing environment variables...]
        LD_LIBRARY_PATH: /usr/local/nvidia/lib64  # Modify For Your Needs
    # https://cloud.google.com/kubernetes-engine/docs/how-to/gpus
    gpuGcpGkeStandard:
      resources:
        limits:
          cpu: 4             # Modify For Your Needs
          memory: 12Gi       # Modify For Your Needs
          nvidia.com/gpu: 1  # Modify For Your Needs
        requests:
          cpu: 4             # Modify For Your Needs
          memory: 12Gi       # Modify For Your Needs
          nvidia.com/gpu: 1  # Modify For Your Needs
```

Upgrade your deployment via `helm upgrade` and wait for the
`k8s-job-manifests` ConfigMap to be updated.

### Deploying GPU-enabled Service Orchestrators

The chart ships a `gpuServiceOrc`
[service orchestrator](../../docs/configuring-service-orchestrator.md)
that requests `nvidia.com/gpu: 1` without a `nodeSelector`.
On GKE Standard with an existing GPU node pool that is enough,
because GKE tolerates the GPU taint for you.
Two GKE configurations need the accelerator labels:

- Autopilot rejects GPU pods that do not set
  `cloud.google.com/gke-accelerator`.
- Node auto-provisioning needs the label to decide which node pool
  to create or grow when scaling from zero.

Add the labels to the shipped orchestrator instead of declaring a new
one,
so its `services` entries are inherited rather than restated.
`nodeSelector` is a map,
so these keys merge into the chart's value:

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      nodeSelector:
        cloud.google.com/gke-accelerator: nvidia-l4  # Modify For Your Needs
        cloud.google.com/gke-accelerator-count: "1"  # Modify For Your Needs
      resources:
        requests:
          cpu: 4          # Modify For Your Needs
          memory: 32Gi    # Modify For Your Needs
```

Pick the accelerator from the
[minimums for each service](../../docs/configuring-service-orchestrator.md#accelerator-sizing).
`nvidia.com/gpu: 1` expresses a GPU count only,
so without these labels a pod can be scheduled onto an accelerator with
too little VRAM for the model,
which fails at model load rather than at scheduling time.

The chart already sets `LD_LIBRARY_PATH` on the `annotation-ai` service
for the
[google GPU driver][gpu-gcp-gke-standard-cuda],
and the `agentic-labeler` service has it built into its image.
A service whose image does not resolve the driver libraries itself needs
the same variable under its `entrypoint.container.env`.

Upgrade your deployment via `helm upgrade` and wait for the
`k8s-job-manifests` ConfigMap to be updated.

## Utilizing Azure AKS GPUs For Delegated Operations

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Prerequisites

This example assumes you have an AKS cluster with GPU-enabled node pools.
Please refer to the
[Azure AKS GPU documentation][gpu-azure-aks]
for assistance in setting up GPU-enabled clusters.

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deploying GPU-enabled Delegated Operator Pods

We will configure the
[always-running delegated operators](./configuring-delegated-operators.md)
with Azure AKS GPUs.

In your `values.yaml`,
under `.Values.delegatedOperatorDeployments.deployments`, add a new delegated
operator deployment.
Ensure the delegated operator deployment has appropriate `tolerations` for
GPU nodes and `resources.requests` to request the desired amount of GPUs
from the Kubernetes scheduler.

The below will deploy a CPU-based delegated operator (`teamsDo`) as well
as a GPU-based delegated operator (`gpuAzureAks`):

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}  # A CPU Based Deployment
    gpuAzureAks:
      resources:
        limits:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
        requests:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
      tolerations:
        - effect: NoSchedule
          key: sku
          operator: Equal
          value: gpu
```

Upgrade your deployment via `helm upgrade` and wait for the
`gpu-azure-aks` pods to be scheduled and deployed.

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deploying GPU-enabled On-Demand Jobs

We will configure the
[on-demand delegated operators](./configuring-delegated-operators.md)
with Azure AKS GPUs.

In your `values.yaml`,
under `.Values.delegatedOperatorJobTemplates.jobs`, add a new delegated
operator job.
Ensure the delegated operator job has appropriate `tolerations` for
GPU nodes and `resources.requests` to request the desired amount of GPUs
from the Kubernetes scheduler.

The below will deploy a CPU-based delegated operator template
(`cpu-default`) as well as a GPU-based delegated operator
template (`gpu-azure-aks`):

```yaml
delegatedOperatorJobTemplates:
  jobs:
    cpuDefault: {}  # A CPU Based Job
    gpuAzureAks:
      resources:
        limits:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
        requests:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
      tolerations:
        - effect: NoSchedule
          key: sku
          operator: Equal
          value: gpu
```

Upgrade your deployment via `helm upgrade` and wait for the
`k8s-job-manifests` ConfigMap to be updated.

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deploying GPU-enabled Service Orchestrators

The chart ships a `gpuServiceOrc`
[service orchestrator](../../docs/configuring-service-orchestrator.md)
that tolerates the `nvidia.com/gpu` taint.
AKS GPU node pools are conventionally tainted `sku=gpu` instead,
so the orchestrator needs that toleration to schedule.

`tolerations` is a list,
which Helm replaces rather than merges,
so the value below supersedes the chart's `nvidia.com/gpu` toleration.
Include both entries if the cluster has node pools using either taint:

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      tolerations:
        - effect: NoSchedule
          key: sku
          operator: Equal
          value: gpu
      resources:
        requests:
          cpu: 4          # Modify For Your Needs
          memory: 32Gi    # Modify For Your Needs
```

Size the node pool's accelerator from the
[minimums for each service](../../docs/configuring-service-orchestrator.md#accelerator-sizing).
On a cluster with more than one accelerator type,
add a `nodeSelector` for the node pool as well;
`nvidia.com/gpu: 1` requests a GPU count and cannot express VRAM.

Upgrade your deployment via `helm upgrade` and wait for the
`k8s-job-manifests` ConfigMap to be updated.

---

## Utilizing AWS EKS GPUs For Delegated Operations

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Prerequisites

This example assumes you have an EKS cluster with GPU-enabled nodes.
Please refer to the
[EKS auto mode accelerated compute documentation][gpu-aws-eks-auto]
or the
[standard EKS GPU workloads documentation][gpu-aws-eks-standard]
for assistance in setting up those clusters.

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deploying GPU-enabled Delegated Operator Pods

We will configure the
[always-running delegated operators](./configuring-delegated-operators.md)
with AWS EKS GPUs.

In your `values.yaml`,
under `.Values.delegatedOperatorDeployments.deployments`, add a new delegated
operator deployment.
Ensure the delegated operator deployment has `resources.requests` set to
request the desired amount of GPUs from the Kubernetes scheduler.

The below will deploy a CPU-based delegated operator (`teamsDo`) as well
as GPU-based delegated operators for both EKS auto mode (`gpuAwsEksAuto`)
and standard EKS (`gpuAwsEksStandard`):

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}  # A CPU Based Deployment
    # https://docs.aws.amazon.com/eks/latest/userguide/auto-accelerated.html
    gpuAwsEksAuto:  # For EKS Auto Mode
      resources:
        limits:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
        requests:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
      tolerations:
        - key: nvidia.com/gpu
          effect: NoSchedule
          operator: Exists
    # https://aws.amazon.com/blogs/compute/running-gpu-accelerated-kubernetes-workloads-on-p3-and-p2-ec2-instances-with-amazon-eks/
    gpuAwsEksStandard:  # For Standard EKS
      resources:
        limits:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
        requests:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
```

Upgrade your deployment via `helm upgrade` and wait for the
GPU-enabled pods to be scheduled and deployed.

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deploying GPU-enabled On-Demand Jobs

We will configure the
[on-demand delegated operators](./configuring-delegated-operators.md)
with AWS EKS GPUs.

In your `values.yaml`,
under `.Values.delegatedOperatorJobTemplates.jobs`, add a new delegated
operator job.
Ensure the delegated operator job has `resources.requests` set to
request the desired amount of GPUs from the Kubernetes scheduler.

The below will deploy a CPU-based delegated operator template
(`cpu-default`) as well as GPU-based delegated operator
templates for both EKS modes:

```yaml
delegatedOperatorJobTemplates:
  jobs:
    cpuDefault: {}  # A CPU Based Job
    # https://docs.aws.amazon.com/eks/latest/userguide/auto-accelerated.html
    gpuAwsEksAuto:  # For EKS Auto Mode
      resources:
        limits:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
        requests:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
      tolerations:
        - key: nvidia.com/gpu
          effect: NoSchedule
          operator: Exists
    # https://aws.amazon.com/blogs/compute/running-gpu-accelerated-kubernetes-workloads-on-p3-and-p2-ec2-instances-with-amazon-eks/
    gpuAwsEksStandard:  # For Standard EKS
      resources:
        limits:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
        requests:
          cpu: 4               # Modify For Your Needs
          memory: 12Gi         # Modify For Your Needs
          nvidia.com/gpu: 1    # Modify For Your Needs
```

Upgrade your deployment via `helm upgrade` and wait for the
`k8s-job-manifests` ConfigMap to be updated.

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deploying GPU-enabled Service Orchestrators

The chart's `gpuServiceOrc`
[service orchestrator](../../docs/configuring-service-orchestrator.md)
already carries the `nvidia.com/gpu` request and the toleration that EKS
GPU nodes use,
so no scheduling changes are required beyond having GPU nodes that
advertise `nvidia.com/gpu`.

The chart sets no cpu or memory request on the orchestrator, so set them
for the model you intend to run,
and add a `nodeSelector` on a cluster with more than one instance type
so the pod lands on an accelerator that meets the
[minimum for the service](../../docs/configuring-service-orchestrator.md#accelerator-sizing):

```yaml
delegatedOperatorJobTemplates:
  serviceOrchestrators:
    gpuServiceOrc:
      nodeSelector:
        node.kubernetes.io/instance-type: g6.2xlarge  # Modify For Your Needs
      resources:
        requests:
          cpu: 4          # Modify For Your Needs
          memory: 32Gi    # Modify For Your Needs
```

Upgrade your deployment via `helm upgrade` and wait for the
`k8s-job-manifests` ConfigMap to be updated.

<!-- Reference Links -->
[gpu-azure-aks]: https://learn.microsoft.com/en-us/azure/aks/use-nvidia-gpu
[gpu-aws-eks-auto]: https://docs.aws.amazon.com/eks/latest/userguide/auto-accelerated.html
[gpu-aws-eks-standard]: https://aws.amazon.com/blogs/compute/running-gpu-accelerated-kubernetes-workloads-on-p3-and-p2-ec2-instances-with-amazon-eks/
[gpu-gcp-gke-autopilot]: https://cloud.google.com/kubernetes-engine/docs/how-to/autopilot-gpus
[gpu-gcp-gke-standard]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus
[gpu-gcp-gke-standard-cuda]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#cuda
[gpu-gcp-gke-standard-multi]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#multiple_gpus
