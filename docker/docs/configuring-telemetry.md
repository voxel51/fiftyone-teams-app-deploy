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
pattern) plus a Redis backend.
The Settings → Metrics page in teams-app displays live CPU / memory /
thread / file-descriptor samples and tailed stdout logs for each
observed service.

**Telemetry is enabled by default.**
The base compose files bundle the `telemetry-redis` service and
per-workload sidecars;
no opt-in flags or overlay files are needed.

## Default deployment

```shell
docker compose -f compose.yaml up -d
```

renders `fiftyone-app`, `teams-api`, `teams-app`, `teams-cas`,
`telemetry-redis`, `fiftyone-app-telemetry`, and `teams-api-telemetry`.

Optional overlays carry their own bundled sidecar.
`compose.yaml`, `compose.plugins.yaml`, and
`compose.dedicated-plugins.yaml` are mutually exclusive base files —
pick one, then layer the `compose.delegated-operators.yaml` overlay on
top.

The delegated-operators overlay defines three worker slots —
`teams-do` (always on) and `teams-do-2` / `teams-do-3` gated behind
cumulative Compose profiles (`do-2`, `do-3`). The default (no
profile) runs one observed worker; set `COMPOSE_PROFILES=do-N` to add
slots up to N. `do-N` includes every lower slot, so `do-3` runs three
workers. For example, to run the dedicated-plugins base with two
delegated-operator workers:

```shell
COMPOSE_PROFILES=do-2 docker compose \
  -f compose.dedicated-plugins.yaml \
  -f compose.delegated-operators.yaml \
  up -d
```

(or set `COMPOSE_PROFILES=do-2` in your `.env`).
This renders the dedicated-plugins base set (`fiftyone-app`,
`teams-api`, `teams-app`, `teams-cas`, `teams-plugins`,
`telemetry-redis`, `fiftyone-app-telemetry`, `teams-api-telemetry`,
`teams-plugins-telemetry`) plus `teams-do` / `teams-do-2` and their
paired sidecars from the overlay.

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

## What's bundled by default

- `telemetry-redis` — Redis 7 container that holds metric streams and
  log entries.
  Data is capped by the `allkeys-lru` maxmemory policy so disk usage
  stays bounded.
- `fiftyone-app-telemetry`, `teams-api-telemetry` — sidecar containers,
  one per observed service.
  Each joins the target's PID namespace via `pid: "service:<target>"`
  so it can read `/proc/<pid>/fd/1` and use psutil to sample CPU,
  memory, FDs, and thread counts.
  Sidecars for delegated operations run with the `SYS_PTRACE` capability so
  py-spy can attach to the target.
- `teams-plugins-telemetry` (only with `compose.dedicated-plugins.yaml`)
  — sidecar for the dedicated `teams-plugins` service.
- `teams-do-telemetry` (only with `compose.delegated-operators.yaml`)
  — sidecar in `EXECUTOR_SIDECAR=true` mode that watches the executor
  for per-operation child processes and records per-op metrics back
  to the `delegated_ops` MongoDB document. Additional workers
  (`teams-do-2`, `teams-do-3`) opt-in via Compose profiles and each
  get their own paired sidecar — see [Scaling teams-do with
  telemetry](#scaling-teams-do-with-telemetry).
- `teams-do-gpu` + `teams-do-gpu-telemetry` (only with
  `compose.delegated-operators.gpu.yaml` and `--profile gpu`) — a
  GPU-enabled delegated-operator worker registered as a distinct
  orchestrator (`-n teams-do-gpu`) plus its paired sidecar.
  The sidecar reads GPU metrics via NVML and requires its own GPU
  reservation.
- `FIFTYONE_TELEMETRY_REDIS_URL` is injected on `fiftyone-app`,
  `teams-api`, `teams-app`, `teams-plugins`, and (when the DO overlay
  is used) `teams-do` so the in-app telemetry blueprint and SSE
  endpoints can read from Redis.

## Opt out

> [!IMPORTANT]
> Disabling the telemetry sidecar leaves the FiftyOne UI's
> delegated-operator log viewer empty — it depends on the sidecar to
> capture per-operation logs.

To run without telemetry, add a `compose.override.yaml` that scales the
telemetry services to zero replicas:

```yaml
services:
  telemetry-redis:
    deploy:
      replicas: 0
  fiftyone-app-telemetry:
    deploy:
      replicas: 0
  teams-api-telemetry:
    deploy:
      replicas: 0
  # Only needed when running the corresponding overlay:
  teams-plugins-telemetry:
    deploy:
      replicas: 0
  teams-do-telemetry:
    deploy:
      replicas: 0
  # Additional delegated-operator slots only run with COMPOSE_PROFILES=do-2/do-3:
  teams-do-2-telemetry:
    deploy:
      replicas: 0
  teams-do-3-telemetry:
    deploy:
      replicas: 0
```

`docker compose -f compose.yaml -f compose.override.yaml up -d` starts
the base services without the telemetry collector.
The main containers still have `FIFTYONE_TELEMETRY_REDIS_URL` set, but
the in-app agent gracefully no-ops when Redis is unreachable.

### Scaling teams-do with telemetry

docker-compose's `pid: "service:<name>"` only joins a single replica's
PID namespace, so a single `teams-do` service scaled to N replicas
would leave N-1 of them invisible to the sidecar.
`compose.delegated-operators.yaml` instead declares three worker
slots, each as its own Compose service with its own paired sidecar and
its own executor-socket volume:

| Slot | Service       | Sidecar                 | Activation             |
| ---- | ------------- | ----------------------- | ---------------------- |
| 1    | `teams-do`    | `teams-do-telemetry`    | always on (no profile) |
| 2    | `teams-do-2`  | `teams-do-2-telemetry`  | profile `do-2`, `do-3` |
| 3    | `teams-do-3`  | `teams-do-3-telemetry`  | profile `do-3`         |

The default (no profile) runs one observed worker. To add more,
activate the matching Compose profile — `do-N` runs N workers because
higher numbers include every lower slot:

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
`--profile do-N` on the command line. Each worker registers under its
own orchestrator name (slot 2 as `teams-do-2`, slot 3 as `teams-do-3`)
so they surface separately in Settings → Metrics.

> [!IMPORTANT]
> This replaces the previous `FIFTYONE_DELEGATED_OPERATOR_WORKER_REPLICAS`
> setting, which is no longer honored — setting it has no effect.
> Pre-telemetry defaults rendered three `teams-do` replicas; the
> post-telemetry default renders one observed worker. If you relied
> on the previous default, set `COMPOSE_PROFILES=do-3` to restore the
> three-worker behavior (each worker now has its own sidecar).

Additionally,

> [!NOTE]
> The cap of 3 is intentional. The value must not exceed your
> license's max concurrent delegated operators. For more than 3
> workers, prefer the helm chart, which automatically attaches a
> telemetry sidecar to every pod in the delegated-operator
> deployment. If you must stay on docker compose, the slot-2 and
> slot-3 blocks in `compose.delegated-operators.yaml` are
> copy-paste templates: duplicate them as `teams-do-4` /
> `teams-do-4-telemetry` (and so on), bumping the service name,
> `pid: "service:teams-do-N"`, `POD_NAME`, `-n teams-do-N`, and
> `telemetry-socket-N` volume on each copy.

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

Telemetry containers ship with conservative CPU and memory limits that
mirror the helm chart's defaults — sized so the sidecars do not starve
the workloads they observe.
The values are declared under each service's `deploy.resources` block
in the compose files;
compose v2 honors `cpus` and `memory` limits/reservations outside swarm
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
