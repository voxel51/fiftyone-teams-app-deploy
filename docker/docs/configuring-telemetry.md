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

The telemetry overlay adds a lightweight per-service metrics collector
(sidecar pattern) plus a Redis backend. Once enabled, the Settings →
Metrics page in teams-app displays live CPU / memory / thread / file-descriptor
samples and tailed stdout logs for each observed service.

This overlay is optional and layered on top of the main compose file.

## Enable

Use the `compose.telemetry.yaml` overlay alongside your normal deployment:

```shell
docker compose \
  -f compose.yaml \
  -f compose.telemetry.yaml \
  up -d
```

Can be combined with the plugins or delegated-operators overlays. For
delegated operators, add the companion `compose.telemetry.delegated-operators.yaml`
overlay so the teams-do executor is also observed:

```shell
docker compose \
  -f compose.yaml \
  -f compose.dedicated-plugins.yaml \
  -f compose.delegated-operators.yaml \
  -f compose.telemetry.yaml \
  -f compose.telemetry.delegated-operators.yaml \
  up -d
```

## What it adds

- `telemetry-redis` — Redis 7 container that holds metric streams and log
  entries. Data is capped by a maxmemory policy (`allkeys-lru`) so disk usage
  stays bounded.
- `fiftyone-app-telemetry`, `teams-api-telemetry` — sidecar containers, one
  per observed service. Each joins the target's PID namespace via
  `pid: "service:<target>"` so it can read `/proc/<pid>/fd/1` and use
  psutil to sample CPU, memory, FDs, and thread counts.
- `teams-do-telemetry` (only with `compose.telemetry.delegated-operators.yaml`) —
  sidecar in `EXECUTOR_SIDECAR=true` mode that watches the executor for
  per-operation child processes and records per-op metrics back to the
  `delegated_ops` MongoDB document.
- `FIFTYONE_TELEMETRY_REDIS_URL` injected on `fiftyone-app`, `teams-api`,
  `teams-app`, and (when the DO overlay is used) `teams-do` so the in-app
  telemetry blueprint and SSE endpoints can read from Redis.

### Scaling teams-do with telemetry

docker-compose's `pid: "service:<name>"` only joins a single replica's
PID namespace. To keep the sidecar observation honest, the telemetry DO
overlay **forces `teams-do` replicas to 1**, overriding
`FIFTYONE_DELEGATED_OPERATOR_WORKER_REPLICAS`.

If you need more than one delegated-operator worker observed at the same
time, either:

1. Define additional explicit services in a compose override — e.g.
   `teams-do-1`, `teams-do-2` — each with a paired `teams-do-N-telemetry`
   sidecar using `pid: "service:teams-do-N"`.
2. Deploy via the helm chart, which automatically adds a telemetry sidecar
   to every pod in the delegated-operator deployment.

## Environment overrides

All knobs live in your `.env` — see `env.template` for the full list:

| Variable                       | Default                             | Purpose                                              |
| ------------------------------ | ----------------------------------- | ---------------------------------------------------- |
| `FIFTYONE_TELEMETRY_REDIS_URL` | `redis://telemetry-redis:6379`      | Override to point at an external Redis if desired    |
| `TELEMETRY_SIDECAR_IMAGE`      | `voxel51/telemetry-sidecar:latest`  | Pin sidecar to a specific tag for production         |
| `TELEMETRY_REDIS_IMAGE`        | `redis:7-alpine`                    | Alternate redis image                                |
| `TELEMETRY_REDIS_MAXMEMORY`    | `400mb`                             | Redis maxmemory budget                               |
| `TELEMETRY_NAMESPACE`          | `docker`                            | Namespace label attached to each registered pod      |
| `FIFTYONE_APP_TARGET_NAME`     | `hypercorn`                         | Substring used to locate the fiftyone-app process    |
| `TEAMS_API_TARGET_NAME`        | `fiftyone.teams.api`                | Substring used to locate the teams-api process       |
| `TEAMS_DO_TARGET_NAME`         | `fiftyone delegated`                | Substring used to locate the teams-do process        |

## Access control

The telemetry endpoints (`/telemetry/*` on teams-api; `/api/telemetry/stream`
and `/api/telemetry/logs` on teams-app) require an authenticated user with
the `ADMIN` role. Non-admin users and unauthenticated requests receive 401/403.

## Verify

```shell
docker compose exec telemetry-redis redis-cli HGETALL active_pods
docker compose exec telemetry-redis redis-cli XLEN metrics:docker:fiftyone-app
```

`XLEN` should increase over time. If it does not, check the sidecar logs:

```shell
docker compose logs fiftyone-app-telemetry teams-api-telemetry --tail 20
```
