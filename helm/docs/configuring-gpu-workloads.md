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
- [Utilizing Azure AKS GPUs For Delegated Operations](#utilizing-azure-aks-gpus-for-delegated-operations)
  - [Prerequisites](#prerequisites-1)
  - [Deploying GPU-enabled Delegated Operator Pods](#deploying-gpu-enabled-delegated-operator-pods-1)
  - [Deploying GPU-enabled On-Demand Jobs](#deploying-gpu-enabled-on-demand-jobs-1)
- [Utilizing AWS EKS GPUs For Delegated Operations](#utilizing-aws-eks-gpus-for-delegated-operations)
  - [Prerequisites](#prerequisites-2)
  - [Deploying GPU-enabled Delegated Operator Pods](#deploying-gpu-enabled-delegated-operator-pods-2)
  - [Deploying GPU-enabled On-Demand Jobs](#deploying-gpu-enabled-on-demand-jobs-2)

<!-- tocstop -->

## Overview

Many machine learning applications utilize GPU hardware for
intensive computations.
The FiftyOne Enterprise helm chart allows users to schedule pods on
GPU-enabled nodes using the `nodeSelector`, `resource`, and `toleration`
settings for individual services.

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
as a GPU-based delegated operator (`teamsDoWithGpu`):

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}  # A CPU Based Deployment
    teamsDoWithGpu:
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
```

Upgrade your deployment via `helm upgrade` and wait for the
`teams-do-with-gpu` pods to be scheduled and deployed.

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
    cpu-default: {}  # A CPU Based Job
    # https://docs.cloud.google.com/kubernetes-engine/docs/how-to/gpus
    gpu-gcp-gke-autopilot:
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
    gpu-gcp-gke-standard:
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
as a GPU-based delegated operator (`teamsDoWithGpu`):

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}  # A CPU Based Deployment
    teamsDoWithGpu:
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
`teams-do-with-gpu` pods to be scheduled and deployed.

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
    cpu-default: {}  # A CPU Based Job
    gpu-azure-aks:
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
as GPU-based delegated operators for both EKS auto mode (`teamsDoWithGpuAuto`)
and standard EKS (`teamsDoWithGpuStandard`):

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDo: {}  # A CPU Based Deployment
    # https://docs.aws.amazon.com/eks/latest/userguide/auto-accelerated.html
    teamsDoWithGpuAuto:  # For EKS Auto Mode
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
    teamsDoWithGpuStandard:  # For Standard EKS
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
    cpu-default: {}  # A CPU Based Job
    # https://docs.aws.amazon.com/eks/latest/userguide/auto-accelerated.html
    gpu-aws-eks-auto:  # For EKS Auto Mode
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
    gpu-aws-eks-standard:  # For Standard EKS
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

<!-- Reference Links -->
[gpu-azure-aks]: https://learn.microsoft.com/en-us/azure/aks/use-nvidia-gpu
[gpu-aws-eks-auto]: https://docs.aws.amazon.com/eks/latest/userguide/auto-accelerated.html
[gpu-aws-eks-standard]: https://aws.amazon.com/blogs/compute/running-gpu-accelerated-kubernetes-workloads-on-p3-and-p2-ec2-instances-with-amazon-eks/
[gpu-gcp-gke-autopilot]: https://cloud.google.com/kubernetes-engine/docs/how-to/autopilot-gpus
[gpu-gcp-gke-standard]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus
[gpu-gcp-gke-standard-cuda]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#cuda
[gpu-gcp-gke-standard-multi]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#multiple_gpus
