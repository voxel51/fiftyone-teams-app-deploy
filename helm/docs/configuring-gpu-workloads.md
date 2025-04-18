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
  - [Validating GPU Access](#validating-gpu-access)

<!-- tocstop -->

## Overview

Many machine learning applications utilize
GPU hardware for intensive computations.
The FiftyOne Enterprise helm chart allows users to schedule pods on
GPU-enabled nodes using the `nodeSelector` values for individual services.

The below will show an example on Google Kubernetes Engine (GKE)
following their
[documentation][gke-gpu-how-to]
in a FiftyOne Enterprise context.

## Utilizing GKE GPUs For Delegated Operations

### Prerequisites

This example assumes you have a GKE cluster with GPU-available nodes.
Please refer to the
[autopilot documentation][gke-autopilot-gke-how-to]
or the
[standard node pool documentation][gke-gpu-how-to]
for assistance in setting up those clusters.

### Deploying GPU-enabled Delegated Operator Pods

We will configure the
[delegated operators](./configuring-delegated-operators.md)
with GKE GPUs.

In your `values.yaml`,
under `.Values.delegatedOperatorDeployments.deployments`, add a new delegated
operator deployment.
Ensure the delegated operator deployment has `nodeSelector` set to valid
[GKE accelerator values and counts][gke-gpu-how-to-multi].
The deployment should also set the `LD_LIBRARY_PATH` variable to the
corresponding
[google GPU driver][gke-gpu-how-to-cuda].
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

### Validating GPU Access

You may validate that the pod can access the GPU drivers using
PyTorch's
[cuda.is_available method][pytorch-cuda-is-available]
by execing into the pod and running

```shell
$ kubectl exec -it -n <YOUR_NAMESPACE> \
    <YOUR_POD> -- python -c 'import torch; print(torch.cuda.is_available())'
True
```

If `True` is printed, then computations may run on GPU hardware.

<!-- Reference Links -->

[gke-autopilot-gke-how-to]: https://cloud.google.com/kubernetes-engine/docs/how-to/autopilot-gpus
[gke-gpu-how-to]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus
[gke-gpu-how-to-cuda]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#cuda
[gke-gpu-how-to-multi]: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#multiple_gpus
[pytorch-cuda-is-available]: https://pytorch.org/docs/stable/generated/torch.cuda.is_available.html
