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

The Agentic Labeler is a builtin service (few-shot VLM inference via vLLM).
It runs on a dedicated GPU delegated-operator worker added by
[internal-auth/compose.agenticlabeler.yaml](../internal-auth/compose.agenticlabeler.yaml)
and
[legacy-auth/compose.agenticlabeler.yaml](../legacy-auth/compose.agenticlabeler.yaml).

For how builtin services are declared, which worker
each one runs on, and accelerator sizing, see
[Configuring Service Orchestrators](../../docs/configuring-service-orchestrator.md).

## Requirements

- A GPU host.
  See
  [configuring GPU workloads](./configuring-gpu-workloads.md)
  for the NVIDIA driver, `nvidia-container-toolkit`, and `nvidia` runtime setup.
  vLLM has no CPU fallback.
- An accelerator meeting the
  [minimum for `agentic-labeler`](../../docs/configuring-service-orchestrator.md#accelerator-sizing).
- The `voxel51/agentic-labeler` image.
- Disk for the model cache.
  The default model is tens of GB; see
  [Persist the model cache](#persist-the-model-cache).

## Run the worker

From your auth mode directory,
add `-f compose.agenticlabeler.yaml` to your usual `-f` set:

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

## Persist the model cache

Starting the service downloads the model weights, which for the default
config are tens of GB and take 30 minutes or more.
Without a cache volume that download repeats on **every** service start.

`compose.agenticlabeler.yaml` ships named volumes for this, so a new
deployment needs no action.
A deployment created before these volumes shipped picks them up on the next
`down` and `up` with the current compose files.

### Mount at the declared paths, not the parent

The image declares the cache directories as Docker `VOLUME`s.
A declared `VOLUME` gets an **anonymous** volume unless something is mounted
at that exact path, and a mount on the parent does not suppress it — Docker
still creates the anonymous volume, it mounts over your named volume, and
the weights land in a volume that is discarded.
The download then repeats on every restart with no error and no symptom
other than the repeated download itself.

So mount each path explicitly:

```yaml
services:
  agentic-labeler:
    volumes:
      - labeler-hf-cache:/home/voxel51/.cache/huggingface
      - labeler-inductor-cache:/home/voxel51/.cache/torchinductor
      - labeler-triton-cache:/home/voxel51/.cache/triton

volumes:
  labeler-hf-cache:
  labeler-inductor-cache:
  labeler-triton-cache:
```

Mounting `/home/voxel51/.cache` alone does **not** work.

### Verify the cache is mounted

```shell
docker inspect <labeler-container> --format '{{json .Mounts}}'
```

Each cache path should show its named volume.
A 64-character hexadecimal volume name at `/home/voxel51/.cache/huggingface`
is an anonymous volume, which means the mount is being shadowed.

### Volume ownership

The service runs as uid 1000.
Docker seeds a **fresh** named volume from the image path including its
ownership, and the image chowns these directories to `1000:1000`, so a new
volume is writable with no extra step.

A volume that already exists and is root-owned is not re-seeded and needs a
one-time fix:

```shell
docker run --rm -v labeler-hf-cache:/cache alpine chown -R 1000:1000 /cache
```

## Start the service

After the `agentic-labeler` Docker Compose service is running,
the service also needs to be started within FiftyOne Enterprise.
In the FiftyOne Enterprise UI, go to *Settings* -> *Services*,
and start `agentic-labeler`.

The service's model loads into GPU memory
and needs substantial host memory.
Size the host accordingly.
An undersized `memory` limit will result in OOM-killing during inference.

## Tuning

### Where `LABELER_*` values are set

Set them in the `env:` block of the `builtin:agentic-labeler` entry in
[builtin_services.yaml](../builtin_services.yaml), on the host running
`teams-api`.
Increment `builtin_version` in the same entry, or the change is not applied.

The `env:` block is merged **over** the worker container's environment, so a
`LABELER_*` value set in the `environment:` block of `compose.override.yaml`
is silently overridden by whatever `builtin_services.yaml` declares.
Change the value in `builtin_services.yaml`.

Settings resolve in this order, highest first:

1. `LABELER_*` environment variables
2. The JSON config file named by `LABELER_CONFIG_FILE`
3. Built-in defaults

### Choosing a model

`LABELER_CONFIG_FILE` selects the model and its engine settings.
The image ships twelve reference configs under `/app/configs/`, in
non-thinking and thinking pairs, covering Gemma 4 (E2B through 31B, including
an int4 QAT build) and Qwen 3.6.
`builtin_services.yaml` pins `/app/configs/gemma4-31B-qat-maxvision.json`.

To run a different model, point `LABELER_CONFIG_FILE` at another shipped
config, or bind mount your own JSON.
On a single 24 GB card, `/app/configs/gemma_4_e2b.json` is the config that
fits; the default 31B config is not viable at that size (see
[accelerator sizing](../../docs/configuring-service-orchestrator.md#accelerator-sizing)).

The full list, and the fields each config accepts, are documented in the
`configs/README.md` shipped alongside them in the image:

```shell
docker exec <labeler-container> cat /app/configs/README.md
docker exec <labeler-container> ls /app/configs/
```

### Reference

Values not listed here come from the config file; anything in the config file
can be overridden with the matching `LABELER_`-prefixed variable.

| Variable | Default | Notes |
| --- | --- | --- |
| `LABELER_CONFIG_FILE` | `/app/configs/gemma4-31B-qat-maxvision.json` | Selects the model and its engine settings. |
| `LABELER_TENSOR_PARALLEL_SIZE` | unset (auto) | Shards one model across GPUs. Leave unset; a pin disables auto-sizing. See [Let the launcher size the deployment](#let-the-launcher-size-the-deployment). |
| `LABELER_DATA_PARALLEL_SIZE` | unset (auto) | Replica count. Leave unset. |
| `LABELER_MAX_MODEL_LEN` | from config file | Total context: prompt plus generation. See [Context length and the token budget](#context-length-and-the-token-budget). |
| `LABELER_MAX_NEW_TOKENS` | from config file | Generation ceiling. Must stay below `max_model_len`; move the two together. |
| `LABELER_ENFORCE_EAGER` | `false` | Disables CUDA graphs. Set `true` on cards of 24 GB or less if you pin a parallel width. |
| `LABELER_GPU_MEMORY_UTILIZATION` | `0.9` | Fraction of VRAM vLLM may use. Lowering it shrinks the KV cache and can make the engine unstartable — it is not a safety margin. |
| `LABELER_MIN_KV_HEADROOM_GB` | `8.0` | Per-GPU VRAM the planner reserves for activations and CUDA graph capture. |
| `LABELER_RUN_CONCURRENCY` | `16` | In-flight requests per run. Lower it for more KV headroom. |

### Let the launcher size the deployment

The service entrypoint is a launcher that inspects the host's GPUs, computes
the model's weight footprint from the checkpoint, and chooses a
tensor-parallel and data-parallel width that leaves `min_kv_headroom_gb` per
GPU for activations and CUDA graph capture.
If the model cannot fit at any width it falls back to a smaller config where
one is declared.

Pinning `LABELER_TENSOR_PARALLEL_SIZE` or `LABELER_DATA_PARALLEL_SIZE`
overrides that: pins are honored verbatim, and only the GPU count is
validated.
The capacity analysis — including the headroom reservation — is skipped
entirely.

If you must pin a width, also set `LABELER_ENFORCE_EAGER=true` on cards of
24 GB or less.
vLLM captures CUDA graphs *after* it allocates the KV cache, and that capture
memory is not fully accounted for in its own profiling, so a deployment that
starts cleanly can still run out of memory during capture.

### Context length and the token budget

Images dominate the prompt.
The shipped Gemma 4 maxvision configs encode each image at up to 1,120 tokens
(`mm_processor_kwargs.max_soft_tokens`) and accept up to 16 images per prompt
(`limit_mm_per_prompt`).
A few-shot request carries every example image plus the target, so a request
with five positive and five negative examples is roughly 12k tokens of image
data before any text.

`max_model_len` must cover all of it:

```text
max_model_len  >=  (examples + 1) x max_soft_tokens  +  prompt text  +  max_new_tokens
```

A request that exceeds it fails at submission:

```text
ValueError: The decoder prompt (length 7064) is longer than the maximum
model length of 6144.
```

There is a ceiling as well.
`max_model_len` cannot exceed the KV cache, which vLLM sizes from whatever
VRAM is left after the weights and reports at startup:

```shell
docker logs <labeler-container> 2>&1 | grep "GPU KV cache size"
```

If `max_model_len` is larger than that token count the engine refuses to
start.
Raising `max_model_len` does not itself consume more VRAM.

## Running `teams-api` and the worker on separate hosts

The `teams-api` service and the GPU worker do not have to be on the same
host.
Four things change when they are not.

**`builtin_services.yaml` is read only by `teams-api`.**
It must be the copy on the host running `teams-api`.
A copy on the GPU host has no effect.

**`entrypoint.container.port` is the address the proxy dials.**
The `teams-api` `/service` proxy connects to this port on the worker, so on a
split-host deployment it must be the port **published on the GPU host**.

**`entrypoint.container.healthcheck.port` is an in-container port.**
The health probe runs *inside* the worker and dials `127.0.0.1` on this port,
so it must be the port the service listens on in the container — `8001` by
default — regardless of what is published on the host.

Setting `healthcheck.port` to a published host port is a common mistake on
split-host deployments.
Nothing is listening on that port inside the container, the service is never
marked healthy, and the Services page reports `STARTING` indefinitely with no
error in any log.

**`FIFTYONE_AUTH_SECRET` must match.**
The value in `.env` on the GPU host must equal the value in `.env` on the
`teams-api` host.

If the worker is on a multi-homed host, or one where the default network
interface is not the reachable one, set `FIFTYONE_SERVICE_POD_IP` on the
worker to the address `teams-api` should dial.
See
[`FIFTYONE_SERVICE_POD_IP`](../../docs/configuring-service-orchestrator.md#fiftyone_service_pod_ip).

## Verify and troubleshoot

Confirm the worker registered under the name the service targets.
It must match the service's `delegation_target`:

```shell
docker logs <worker-container> 2>&1 | grep "Registering executor"
```

Confirm the health endpoint answers inside the container.
This is the exact check the health probe makes:

```shell
docker exec <labeler-container> \
  curl -so /dev/null -w '%{http_code}\n' http://127.0.0.1:8001/healthz
```

Confirm the cache is mounted and not shadowed:

```shell
docker inspect <labeler-container> --format '{{json .Mounts}}'
```

Read the startup sizing decisions:

```shell
docker logs <labeler-container> 2>&1 \
  | grep -E "KV cache|Model loading|downloading weights"
```

`GPU KV cache size` is the ceiling to size `max_model_len` against.
`Time spent downloading weights` should appear on the first start and be
**absent** on later starts; if it reappears every time, the cache volume is
not taking effect.

| Symptom | Likely cause |
| --- | --- |
| Services page stuck on `STARTING`, no errors | `healthcheck.port` is not the in-container port |
| Model re-downloads on every start | Cache volume shadowed by a `VOLUME` declaration, or mounted on the parent path |
| `decoder prompt ... longer than the maximum model length` | `max_model_len` too small for the number of example images |
| CUDA out of memory during startup | A pinned parallel width skipped the capacity analysis; unset the pin, or set `LABELER_ENFORCE_EAGER=true` |
| Engine refuses to start, KV cache smaller than `max_model_len` | `max_model_len` above the ceiling, or `gpu_memory_utilization` lowered |
| Changes to `builtin_services.yaml` have no effect | `builtin_version` not incremented, or the file edited on the wrong host |
