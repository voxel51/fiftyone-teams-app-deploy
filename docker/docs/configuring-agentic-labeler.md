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
- Internet access to Hugging Face, or a pre-downloaded model
  (see [Model downloads](#model-downloads)).

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

## Start the service

After the `agentic-labeler` Docker Compose service is running,
the service also needs to be started within FiftyOne Enterprise.
In the FiftyOne Enterprise UI, go to *Settings* -> *Services*,
and start `agentic-labeler`.

The service's model loads into GPU memory
and needs substantial host memory.
Size the host accordingly.
An undersized `memory` limit will result in OOM-killing during inference.

Tune the `LABELER_*` values in
[builtin_services.yaml](../builtin_services.yaml) for your model and GPU.
When changing values, increment the `builtin_version` to trigger a refresh.

## Model downloads

The service does not ship with its model.
Each time the `agentic-labeler` container is created, the service downloads
the model (about 24 GB) from [Hugging Face](https://huggingface.co),
unless it is already stored in a volume
(see [Keep the model between restarts](#keep-the-model-between-restarts-recommended)).
Plan for the first start to take several minutes,
and leave at least 30 GB of free disk space on the host.

The default models are public.
You do not need a Hugging Face account to download them.

> [!NOTE]
> If you would like to use a different model,
> contact your Voxel51 Customer Success team with the request.

### Applying changes

The settings below go in your `compose.override.yaml`.
After changing them:

1. Run your usual `docker compose ... up -d` command.
1. In the FiftyOne Enterprise UI, go to *Settings* -> *Services*,
   then stop and start `agentic-labeler`.

### Network access

The `agentic-labeler` container needs outbound HTTPS access to
`huggingface.co` and `*.hf.co`.
If your network uses a proxy, set it on the `agentic-labeler` service.
`NO_PROXY` must include `teams-api` so the service can still reach the
FiftyOne API:

```yaml
services:
  agentic-labeler:
    environment:
      HTTPS_PROXY: http://proxy.example.com:3128  # Modify as needed
      https_proxy: http://proxy.example.com:3128  # Modify as needed
      NO_PROXY: teams-api,localhost,127.0.0.1
      no_proxy: teams-api,localhost,127.0.0.1
```

For more on proxies, see [configuring proxies](./configuring-proxies.md).
If the host cannot reach Hugging Face at all, see
[Offline and air-gapped hosts](#offline-and-air-gapped-hosts).

### Hugging Face token (recommended)

A Hugging Face token is not required, but we recommend one.
Without it, downloads can be slow enough for the service to time out
on start.

1. Create a read-only
   [Hugging Face access token](https://huggingface.co/docs/hub/security-tokens).
1. Add it to your `.env` file:

    ```shell
    HF_TOKEN=<your-token>
    ```

1. Pass it to the `agentic-labeler` service in `compose.override.yaml`:

    ```yaml
    services:
      agentic-labeler:
        environment:
          HF_TOKEN: ${HF_TOKEN}
    ```

Do not put the token in [builtin_services.yaml](../builtin_services.yaml).
Values in that file are not treated as secrets.

### Keep the model between restarts (recommended)

Store the model in a named volume so it is downloaded only once.
Add this to `compose.override.yaml`:

```yaml
services:
  agentic-labeler:
    volumes:
      - agentic-labeler-hf-cache:/home/voxel51/.cache/huggingface

volumes:
  agentic-labeler-hf-cache:
    name: agentic-labeler-hf-cache
```

### Offline and air-gapped hosts

If the host cannot reach Hugging Face, download the model on a host
that can, then copy it over.
The target host also needs the `voxel51/agentic-labeler` image.
Transfer it the same way as your other FiftyOne Enterprise images
(for example, with `docker save` and `docker load`).

1. On a host with internet access, download the model into a volume.
   Replace `<version>` with the image tag in your
   `compose.agenticlabeler.yaml`:

    ```shell
    export HF_TOKEN=<your-token>  # Optional
    docker run --rm \
      -e HF_TOKEN \
      -v agentic-labeler-hf-cache:/home/voxel51/.cache/huggingface \
      voxel51/agentic-labeler:<version> \
      python3 -c "from huggingface_hub import snapshot_download as d; \
    d('google/gemma-4-31B-it-qat-w4a16-ct'); d('google/gemma-4-31B-it-assistant')"
    ```

    The second model is a small helper model that speeds up labeling.

1. Export the volume to a file:

    ```shell
    docker run --rm \
      -v agentic-labeler-hf-cache:/cache \
      -v "$PWD":/backup \
      alpine tar cf /backup/hf-cache.tar -C /cache .
    ```

1. Copy `hf-cache.tar` to the target host and import it:

    ```shell
    docker volume create agentic-labeler-hf-cache
    docker run --rm \
      -v agentic-labeler-hf-cache:/cache \
      -v "$PWD":/backup \
      alpine sh -c "tar xf /backup/hf-cache.tar -C /cache && chown -R 1000:1000 /cache"
    ```

1. On the target host, add the volume and turn on offline mode in
   `compose.override.yaml`, then [apply the changes](#applying-changes):

    ```yaml
    services:
      agentic-labeler:
        environment:
          HF_HUB_OFFLINE: "1"
        volumes:
          - agentic-labeler-hf-cache:/home/voxel51/.cache/huggingface

    volumes:
      agentic-labeler-hf-cache:
        name: agentic-labeler-hf-cache
    ```

If you changed `LABELER_CONFIG_FILE` to a different model,
download the models named by `model_id` and `speculative_config.model`
in that config file instead.

## Troubleshooting

| Error or symptom | What to do |
| --- | --- |
| `Deployment planner timed out after 120s` | The service could not reach Hugging Face quickly enough. Check network and proxy access to `huggingface.co`, and set `HF_TOKEN`. If downloads work but are slow, raise the limit as shown below this table. |
| `Could not resolve the weights of target model` | The service cannot reach Hugging Face and the model is not downloaded. Check network and proxy access, or follow [Offline and air-gapped hosts](#offline-and-air-gapped-hosts). |
| Every start is as slow as the first | The model is not being kept between restarts. Add the [named volume](#keep-the-model-between-restarts-recommended). |
| `Permission denied` under `/home/voxel51/.cache` | The volume is not writable by the container user. Run `chown -R 1000:1000` on the volume contents. |

To raise the startup time limit (in seconds), add this to
`compose.override.yaml`, then [apply the changes](#applying-changes):

```yaml
services:
  agentic-labeler:
    environment:
      LABELER_LAUNCH_PLAN_TIMEOUT_S: "300"
```
