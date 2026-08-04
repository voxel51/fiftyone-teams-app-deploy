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

# Configuring the Agentic Labeler Service

The Agentic Labeler is a builtin service (few-shot VLM inference via vLLM). It
runs on a dedicated GPU delegated-operator worker added by
[compose.agenticlabeler.yaml](../internal-auth/compose.agenticlabeler.yaml).

This page covers that worker. For how builtin services are declared, which
worker each one runs on, and accelerator sizing, see
[Configuring Service Orchestrators](../../docs/configuring-service-orchestrator.md).

## Requirements

- A GPU host. See
  [configuring GPU workloads](./configuring-gpu-workloads.md) for the NVIDIA
  driver, `nvidia-container-toolkit`, and `nvidia` runtime setup. vLLM has no
  CPU fallback.
- An accelerator meeting the
  [minimum for `agentic-labeler`](../../docs/configuring-service-orchestrator.md#accelerator-sizing).
- The `voxel51/fiftyone-teams-agentic-labeler` image. Contact your Voxel51
  support team for Docker Hub access.

## Run the worker

From your auth-mode directory, add `compose.agenticlabeler.yaml` to your usual
`-f` set:

```shell
docker compose \
  -f compose.dedicated-plugins.yaml \
  -f compose.delegated-operators.yaml \
  -f compose.agenticlabeler.yaml \
  -f compose.override.yaml \
  up -d
```

On upgrade, add the same file to your existing `down` and `up` commands (see
[Upgrades](../README.md#upgrades)).

## Start the service

The service is created stopped. In the FiftyOne Enterprise UI, go to
`Settings -> Services` and start `agentic-labeler`.

The model loads into GPU memory and needs substantial host RAM. Size the host
accordingly. An undersized `memory` limit is OOM-killed during inference.

Tune the `LABELER_*` values in
[builtin_services.yaml](../builtin_services.yaml) for your model and GPU. Bump
that entry's `builtin_version` so a changed value re-applies to an environment
that has already stored the service.
