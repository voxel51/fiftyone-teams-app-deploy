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
[compose.agenticlabeler.yaml](../internal-auth/compose.agenticlabeler.yaml). The
service itself is declared in the deployment's builtin services list,
[builtin_services.yaml](../builtin_services.yaml), which teams-api mounts and
reconciles at startup.

## Requirements

- A GPU host. See
  [configuring GPU workloads](./configuring-gpu-workloads.md) for the NVIDIA
  driver, `nvidia-container-toolkit`, and `nvidia` runtime setup. vLLM has no
  CPU fallback.
- The `voxel51/fiftyone-teams-agentic-labeler` image. Contact your Voxel51
  support team for Docker Hub access.

## Run the worker

The worker is gated behind the `agentic-labeler` Compose profile. From your
auth-mode directory, add `compose.agenticlabeler.yaml` to your usual `-f` set
and set the profile:

```shell
docker compose --profile agentic-labeler \
  -f compose.dedicated-plugins.yaml \
  -f compose.delegated-operators.yaml \
  -f compose.agenticlabeler.yaml \
  -f compose.override.yaml \
  up -d
```

On upgrade, add the same file and profile to your existing `down` and `up`
commands (see [Upgrades](../README.md#upgrades)).

## Builtin services

The services are declared in [builtin_services.yaml](../builtin_services.yaml),
mounted into teams-api by `common-services.yaml` and reconciled at startup.
Both ship created stopped:

- `annotation-ai` (SAM2) targets the default `teams-do` worker
  (`delegation_target: builtin`) for convenience. SAM2 needs a GPU, so that
  worker must have GPU access, or retarget it to a GPU worker.
- `agentic-labeler` targets the dedicated GPU worker above
  (`delegation_target: agentic-labeler`).

This is the deployment's full builtin services list, deep-merged by `id`. Add,
remove, or retarget services by editing that file. Bump an entry's
`builtin_version` to re-apply a change to an environment that already stored it.

## Start a service

Both services are created stopped. In the FiftyOne Enterprise UI, go to
`Settings -> Services` and start the one you want. The Agentic Labeler model
loads into GPU memory and needs substantial host RAM. Size the host
accordingly. An undersized `memory` limit is OOM-killed during inference.

## `FIFTYONE_SERVICE_POD_IP`

The worker hosts the service in-process and publishes the address the
`teams-api` proxy uses to reach it. It auto-detects its own container IP at
runtime, which is correct on a standard single-network Compose host. On
multi-homed or non-default-network hosts, where auto-detect can pick the wrong
interface, set `FIFTYONE_SERVICE_POD_IP` on the `agentic-labeler` service to
the reachable address. This mirrors the Kubernetes path, where the resolver
reads the injected `POD_IP`.
