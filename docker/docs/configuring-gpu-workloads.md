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
- [Utilizing GPUs For Delegated Operations](#utilizing-gpus-for-delegated-operations)
  - [Prerequisites](#prerequisites)
  - [Deploying GPU-enabled Delegated Operator Containers](#deploying-gpu-enabled-delegated-operator-containers)
  - [Validating GPU Access](#validating-gpu-access)

<!-- tocstop -->

## Overview

Many machine learning applications utilize
GPU hardware for intensive computations.
The FiftyOne Enterprise docker compose files allow users to schedule containers on
GPU-enabled nodes using a service's
[`deploy.resource.reservation.devices`][compose-deploy-resources].

The below will show an example deploying a GPU-enabled container via docker
compose by following their
[documentation][compose-gpu-how-to]
in a FiftyOne Enterprise context.

## Utilizing GPUs For Delegated Operations

### Prerequisites

This example assumes you have a docker compose node with GPU devices available
and you have followed Docker's
[GPU prerequisites][compose-gpu-resources].

### Deploying GPU-enabled Delegated Operator Containers

We will configure the
[delegated operators](./configuring-delegated-operators.md)
with GPUs.

In your `compose.delegated-operators.yaml`, under `.services.teams-do`,
add the `deploy.resource.reservation.devices` configuration.

The below will deploy a CPU-based delegated operator (`teams-do`) as well
as a GPU-based delegated operator (`teams-do-with-gpu`):

```yaml
services:
  teams-do:
    extends:
      file: ../common-services.yaml
      service: teams-do-common

  teams-do-with-gpu:
    image: voxel51/fiftyone-teams-cv-full:v2.11.0
    command: >
      /bin/sh -c "fiftyone delegated launch -t remote  -n 'teams-do-with-gpu'"
    environment:
      API_URL: ${API_URL}
      FIFTYONE_DATABASE_ADMIN: false
      FIFTYONE_DATABASE_NAME: ${FIFTYONE_DATABASE_NAME}
      FIFTYONE_DATABASE_URI: ${FIFTYONE_DATABASE_URI}
      FIFTYONE_ENCRYPTION_KEY: ${FIFTYONE_ENCRYPTION_KEY}
      FIFTYONE_INTERNAL_SERVICE: true
      FIFTYONE_MEDIA_CACHE_SIZE_BYTES: -1
      FIFTYONE_PLUGINS_DIR: /opt/plugins
    restart: always
    volumes:
      - plugins-vol:/opt/plugins:ro
    deploy:
      replicas: ${FIFTYONE_DELEGATED_OPERATOR_WORKER_REPLICAS:-3}
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

Redeploy the stack via `docker compose up -d` and wait for the
`teams-do` containers to be deployed.

For advanced GPU configuration, including selecting specific GPU devices,
please refer to the
[docker documentation][compose-gpu-how-to].

### Validating GPU Access

You may validate that the container can access the GPU drivers using
PyTorch's
[cuda.is_available method][pytorch-cuda-is-available]
by execing into the container and running

```shell
$ docker compose exec teams-do-with-gpu \
    python -c 'import torch; print(torch.cuda.is_available())'
True
```

If `True` is printed, then computations may run on GPU hardware.

<!-- Reference Links -->

[compose-deploy-resources]: https://docs.docker.com/reference/compose-file/deploy/#resources
[compose-gpu-how-to]: https://docs.docker.com/compose/how-tos/gpu-support/
[compose-gpu-resources]: https://docs.docker.com/engine/containers/resource_constraints/#gpu
[pytorch-cuda-is-available]: https://pytorch.org/docs/stable/generated/torch.cuda.is_available.html
