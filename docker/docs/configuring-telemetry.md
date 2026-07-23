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

# Configuring FiftyOne Enterprise Telemetry

Telemetry adds a lightweight per-service metrics collector (sidecar
pattern) and a Redis backend.
In the app, the Settings → Metrics page in displays live CPU / memory /
thread / file-descriptor samples and tailed stdout logs for each
observed service.

**Telemetry is enabled by default.**
The base compose files contain the `telemetry-redis` service and
per-workload sidecars.

## Default deployment

Running

```shell
docker compose -f compose.yaml up -d
```

renders `fiftyone-app`, `teams-api`, `teams-app`, `teams-cas`,
`telemetry-redis`, `fiftyone-app-telemetry`, and `teams-api-telemetry`.

Optional overlays provide their own bundled sidecar.
The `compose.yaml`, `compose.plugins.yaml`, and
`compose.dedicated-plugins.yaml` are mutually exclusive base files —
pick one, then layer the `compose.delegated-operators.yaml` overlay on top.

The delegated-operators overlay defines three worker slots

1. `teams-do` (always on)
1. `teams-do-2`
    1. Enabled using the `do-2` and `do-3` Compose profiles
1. `teams-do-3`
    1. Enabled using the `do-3` Compose profile

The default (no profile) runs one delegated operator worker.
Set `COMPOSE_PROFILES=do-<N>` to add slots up to `<N>`.
`do-<N>` includes previous slots.
For example `do-3` runs three workers.

To run the dedicated-plugins base with two delegated-operator workers either:

1. Set `COMPOSE_PROFILES=do-2` in your `.env`
1. Set the environment variable while calling compose up

This renders the services

- FiftyOne Enterprise
  - `fiftyone-app`,
  - `teams-api`
  - `teams-app`
  - `teams-cas`
  - `teams-plugins`,
- Telemetry
  - `telemetry-redis`
  - `fiftyone-app-telemetry`
  - `teams-api-telemetry`,
  - `teams-plugins-telemetry`
  - `teams-do-telemetry`
  - `teams-do-2-telemetry`
- Delegated Operator
  - `teams-do`
  - `teams-do-2`

For GPU-enabled delegated operators, layer
`compose.delegated-operators.gpu.yaml` (which includes its own bundled
telemetry sidecar) and activate the `gpu` profile:

```shell
docker compose --profile gpu \
  -f compose.yaml \
  -f compose.delegated-operators.yaml \
  -f compose.delegated-operators.gpu.yaml \
  up -d
```

## What's included

- Services
  - `telemetry-redis`
    - Redis 7 container that holds metric streams and log entries.
    - Data is capped by the `allkeys-lru` maxmemory policy so disk usage
      stays bounded.
  - `teams-do-gpu`
    - only with `compose.delegated-operators.gpu.yaml` and `--profile gpu`
    - GPU-enabled delegated-operator worker registered as a distinct
      orchestrator (`-n teams-do-gpu`) plus its paired sidecar.
- Service Side Cars
  - `fiftyone-app-telemetry` and `teams-api-telemetry`
    - Each joins the target's PID namespace via `pid: "service:<target>"`
      so it can read `/proc/<pid>/fd/1` and use psutil to sample CPU,
      memory, file descriptors, and thread counts
    - For delegated operations run with the `SYS_PTRACE` capability so
      py-spy can attach to the target
  - `teams-plugins-telemetry`
    - Only with `compose.dedicated-plugins.yaml`
    - For the dedicated `teams-plugins` service
  - `teams-do-telemetry`
    - Only with `compose.delegated-operators.yaml`
    - In `EXECUTOR_SIDECAR=true` mode that watches the executor
      for per-operation child processes and records per-operation metrics
      back to the `delegated_ops` MongoDB document
    - Additional workers (`teams-do-2`, `teams-do-3`)
      - Opt-in via Compose profiles and each
        get their own paired sidecar
      - See
        [Scaling teams-do with telemetry](#scaling-teams-do-with-telemetry)
  - `teams-do-gpu-telemetry`
    - Only with `compose.delegated-operators.gpu.yaml` and `--profile gpu`
    - The sidecar reads GPU metrics via NVML and requires its own GPU
      reservation.
- Environment Variables
  - `FIFTYONE_TELEMETRY_REDIS_URL` environment variable is set on these services
    so the telemetry blueprint and server-sent events endpoints can read from Redis
    - `fiftyone-app`
    - `teams-api`
    - `teams-app`
    - `teams-plugins`
    - (when the DO overlay is used) `teams-do`

## Opting out

> [!IMPORTANT]
> Disabling the telemetry sidecar leaves the FiftyOne UI's
> delegated-operator log viewer empty.
> The log viewer depends on the sidecar to capture per-operation logs.

To run without telemetry

1. Add a `compose.override.yaml` that scales the
   telemetry services to zero replicas:

### If your environment disallows `SYS_PTRACE`

The delegated-operator sidecars add the `SYS_PTRACE` capability for additional
`py-spy` stack sampling, enabled by default.
If your host or Docker policy won't permit the capability, you can drop it and
keep the rest of telemetry using the [`!reset` tag][compose-merge]
(Docker Compose v2.24+) in a `compose.override.yaml`, with one entry per
delegated-operator sidecar you run:

```yaml
services:
  teams-do-telemetry:
    cap_add: !reset []
  # add these only when running the matching profile/overlay:
  # teams-do-2-telemetry: {cap_add: !reset []}   # COMPOSE_PROFILES=do-2|do-3
  # teams-do-3-telemetry: {cap_add: !reset []}   # COMPOSE_PROFILES=do-3
  # teams-do-gpu-telemetry: {cap_add: !reset []} # compose.delegated-operators.gpu.yaml
```

[compose-merge]: https://docs.docker.com/reference/compose-file/merge/#reset-value

### Scaling teams-do with telemetry

Docker Compose's `pid: "service:<name>"` only joins a single replica's PID namespace.
Thus single `teams-do` service scaled to `<N>` replicas
would leave `<N-1>` of them invisible to the sidecar.
Instead, `compose.delegated-operators.yaml` contains three worker
slots as its own Compose service, paired service sidecar, and
its own executor-socket volume:

| Slot | Service       | Sidecar                 | Activation               |
| ---- | ------------- | ----------------------- | ------------------------ |
| 1    | `teams-do`    | `teams-do-telemetry`    | always on (no profile)   |
| 2    | `teams-do-2`  | `teams-do-2-telemetry`  | profile `do-2` or `do-3` |
| 3    | `teams-do-3`  | `teams-do-3-telemetry`  | profile `do-3`           |

The default (no profile) runs one delegated operator worker. To add more,
activate the matching Compose profile — `do-<N>` runs `<N>` workers because
higher numbers include every prior slot:

```shell
# 1 worker (default):
docker compose -f compose.yaml \
  -f compose.delegated-operators.yaml up -d

# 2 workers:
COMPOSE_PROFILES=do-2 docker compose -f compose.yaml \
  -f compose.delegated-operators.yaml up -d

# 3 workers:
COMPOSE_PROFILES=do-3 docker compose -f compose.yaml \
  -f compose.delegated-operators.yaml up -d
```

Set `COMPOSE_PROFILES` in your `.env` to persist the choice, or pass
`--profile do-<N>` on the command line. Each worker registers under its
own orchestrator name (slot 2 as `teams-do-2`, slot 3 as `teams-do-3`)
so they surface distinctly in Settings → Metrics.

> [!IMPORTANT]
> v2.19 deprecates the `FIFTYONE_DELEGATED_OPERATOR_WORKER_REPLICAS`
> environment variable.
> Prior to v2.19, the `teams-do` replica count defaulted to `3`.
> To retain the prior replica count behavior,
> set `COMPOSE_PROFILES=do-3` in your `.env` file.
> [!NOTE]
> The cap of 3 is intentional as the value must not exceed your
> license's max concurrent delegated operators.
> For more than 3 workers, use the slot-3 blocks in
> `compose.delegated-operators.yaml` as templates.
> Duplicate them as `teams-do-4` / `teams-do-4-telemetry`,
> bumping the service name, `pid: "service:teams-do-N"`, `POD_NAME`,
> `-n teams-do-N`, and `telemetry-socket-N` volume on each copy.

### Sidecar lifecycle on workload restart

Each sidecar joins its workload's PID namespace at container-create
time.
If the workload is recreated (force-recreate, image upgrade, config
change) the namespace reference goes stale and the sidecar stays in
`Exited (137)` until manually recreated.

The compose files set `depends_on.<target>.restart: true` so the
sidecar is recreated in lockstep with the workload.
This requires Docker Compose v2.17 or newer.
If the workload crash-loops, the sidecar follows it.

## Environment overrides

All knobs live in your `.env` — see `env.template` for the full list:

| Variable                           | Default                                | Purpose                                                                                            |
| ---------------------------------- | -------------------------------------- | -------------------------------------------------------------------------------------------------- |
| `FIFTYONE_TELEMETRY_REDIS_URL`     | `redis://telemetry-redis:6379`         | Override to point at an external Redis if desired                                                  |
| `TELEMETRY_REDIS_IMAGE`            | `redis:7-alpine`                       | Alternate redis image                                                                              |
| `TELEMETRY_REDIS_MAXMEMORY`        | `400mb`                                | Redis maxmemory budget                                                                             |
| `TELEMETRY_NAMESPACE`              | `docker`                               | Namespace label attached to each registered pod                                                    |
| `FIFTYONE_APP_TARGET_NAME`         | `hypercorn`                            | Substring used to locate the fiftyone-app process                                                  |
| `TEAMS_API_TARGET_NAME`            | `fiftyone-teams-api`                   | Substring used to locate the teams-api process                                                     |
| `TEAMS_PLUGINS_TARGET_NAME`        | `hypercorn`                            | Substring used to locate the teams-plugins process                                                 |
| `TEAMS_DO_TARGET_NAME`             | `fiftyone delegated`                   | Substring used to locate the teams-do process                                                      |
| `NVIDIA_GPU_COUNT`                 | `1`                                    | GPU reservation for the GPU DO worker + sidecar                                                    |
| `NVIDIA_VISIBLE_DEVICES`           | `all`                                  | Pass-through to teams-do-gpu / sidecar                                                             |
| `NVIDIA_DRIVER_CAPABILITIES`       | `compute,utility`                      | Must include `utility` so NVML is available                                                        |

## Resource limits

Telemetry containers ship with conservative CPU and memory limits sized so the
sidecars do not starve the workloads they observe.
The values are declared under each service's `deploy.resources` block
in the Cpmpose files.
Compose v2 honors `cpus` and `memory` limits and reservations outside swarm
mode.

| Service                       | CPU limit | Memory limit | Notes                               |
| ----------------------------- | --------- | ------------ | ----------------------------------- |
| `telemetry-redis`             | `0.25`    | `512M`       | Reserves `0.10` CPU / `256M` memory |
| `*-telemetry` (any sidecar)   | `0.10`    | `512M`       | Reserved == limit                   |

To tune these for your hardware, override `deploy.resources` in a
`compose.override.yaml` (the override merges with the base entry).

## Verify

```shell
docker compose exec telemetry-redis redis-cli HGETALL active_targets
docker compose exec telemetry-redis redis-cli XLEN metrics:fiftyone-app
```

`XLEN` should increase over time.
If it does not, check the sidecar logs:

```shell
docker compose logs fiftyone-app-telemetry teams-api-telemetry --tail 20
```
