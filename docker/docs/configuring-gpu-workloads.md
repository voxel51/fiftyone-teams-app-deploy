<!-- markdownlint-disable no-inline-html line-length no-alt-text -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length no-alt-text -->

---

# Leveraging GPU Workloads

<!-- toc -->

- [Overview](#overview)
- [Utilizing GPUs For Delegated Operations](#utilizing-gpus-for-delegated-operations)
  - [Prerequisites](#prerequisites)
  - [Deploying GPU-enabled Delegated Operator Containers](#deploying-gpu-enabled-delegated-operator-containers)
  - [Validating GPU Access](#validating-gpu-access)

<!-- tocstop -->

## Overview

In many machine learning applications, it is desirable to utilize available
GPU hardware for intensive computations.
The FiftyOne Enterprise compose files allow users to schedule containers on
GPU-enabled nodes using the docker compose
[`deploy.resource.reservation.devices`][compose-deploy-resources]
settings for individual services.

The below will show an example deploying a GPU-enabled container via docker
compose by following their
[documentation][compose-gpu-how-to]
in a FiftyOne Enterprise context.

## Utilizing GPUs For Delegated Operations

### Prerequisites

This example assumes you have a docker compose node with GPU devices available.
This example also assumes you have followed the
[GPU prerequisites][compose-gpu-resources]
outlined by docker.

### Deploying GPU-enabled Delegated Operator Containers

In this example, we will leverage GPUs for
[delegated operators](./configuring-delegated-operators.md).

Under `.services.teams-do` in your `compose.delegated-operators.yaml`,
add the `deploy.resource.reservation.devices` configuration:

```yaml
  teams-do:
    extends:
      file: ../common-services.yaml
      service: teams-do-common
      deploy:
        resources:
          reservations:
            devices:
              - driver: nvidia
                count: 1
                capabilities: [gpu]
```

Now redeploy your stack via `docker compose up -d` and wait for the
`teams-do` containers to be deployed.

For advanced GPU configuration, including selecting specific GPU devices,
please refer to the
[docker documentation][compose-gpu-how-to].

### Validating GPU Access

You can validate that the container can correctly access the GPU drivers using
PyTorch's
[cuda.is_available method][pytorch-cuda-is-available]
from within the container.

```shell
$ docker compose exec teams-do \
    python -c 'import torch; print(torch.cuda.is_available())'
True
```

If `True` is printed, then you are ready to run computations on GPU hardware.

<!-- Reference Links -->

[compose-deploy-resources]: https://docs.docker.com/reference/compose-file/deploy/#resources
[compose-gpu-how-to]: https://docs.docker.com/compose/how-tos/gpu-support/
[compose-gpu-resources]: https://docs.docker.com/engine/containers/resource_constraints/#gpu
[pytorch-cuda-is-available]: https://pytorch.org/docs/stable/generated/torch.cuda.is_available.html
