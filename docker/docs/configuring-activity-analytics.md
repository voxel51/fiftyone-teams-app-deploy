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

# Configuring Activity Analytics

Activity Analytics records what happens in a deployment and aggregates it
for the Audit Log and Jobs pages in the app.

**Activity Analytics is opt-in.**
The `fiftyone-mq-redis` and `activity-worker` services live in the
`compose.activity.yaml` overlay, not in the base compose files. A plain
`docker compose up` does not start them, and no activity is recorded.
`FIFTYONE_ACTIVITY_ENABLED` defaults to `false` besides, so the emit
seams in the services that do start are no-ops. This matches the Helm
chart, where `activitySettings.enabled` and `fiftyoneMq.enabled` both
default to `false` and the chart renders no activity env at all until
they are on.

## How it works

Activity events travel over a Redis-backed queue (`fiftyone.mq`):

1. `teams-api` and `fiftyone-app` enqueue events. `teams-api` hosts the
   domain-event bridge, and `fiftyone-app` emits operator and annotation
   activity. Both emit only while `FIFTYONE_ACTIVITY_ENABLED` is true, and
   read `FIFTYONE_MQ_REDIS_URL` to find the queue once it is.
1. `fiftyone-mq-redis` holds the queue. It is a stock `redis:7` that
   persists the queue with AOF onto the `fiftyone-mq-redis-data` volume,
   so a restart replays queued jobs instead of losing them. See
   [Queue durability](#queue-durability) for what that does and does not
   cover.
1. `activity-worker` consumes the queue and writes the `activity_*`
   collections. They are co-located in the deployment's own FiftyOne
   database, so they are covered by your existing MongoDB backups.

The emit path is best-effort and fire-and-forget. If the queue is
unavailable, events are dropped and the emitting operation still
succeeds. To stop events flowing, leave `FIFTYONE_ACTIVITY_ENABLED`
false — that short-circuits before a client is built, rather than
relying on a connection failing.

One `activity-worker` container runs four workers via the combined
`fiftyone-activity-worker` entrypoint:

| Worker   | Responsibility                                         |
| -------- | ------------------------------------------------------ |
| ingest   | Drains the queue into the raw `activity_*` collections |
| rollup   | Aggregates raw events into the rollups the app reads   |
| snapshot | Periodically records dataset and deployment state      |
| prune    | Enforces the retention window and the storage size cap |

## Enabling Activity Analytics

From your auth-mode directory, add `compose.activity.yaml` to your usual
`-f` set:

```shell
docker compose \
  -f compose.yaml \
  -f compose.activity.yaml \
  -f compose.override.yaml \
  up -d
```

That renders `fiftyone-mq-redis` and `activity-worker` alongside the
other services. `activity-worker` declares
`depends_on: fiftyone-mq-redis`.

The overlay layers onto any base file, so substitute
`compose.plugins.yaml` or `compose.dedicated-plugins.yaml` for
`compose.yaml` if that is what you deploy. It also composes with the
delegated-operator and GPU overlays:

```shell
docker compose \
  -f compose.dedicated-plugins.yaml \
  -f compose.delegated-operators.yaml \
  -f compose.activity.yaml \
  -f compose.override.yaml \
  up -d
```

### `FIFTYONE_ACTIVITY_ENABLED` is the gate

`FIFTYONE_ACTIVITY_ENABLED` decides whether a service emits at all. It
defaults to `false`, and `emit`, `flush`, and the operator mutation
capture are no-ops while it is — checked before any queue client is
constructed. `FIFTYONE_MQ_REDIS_URL` says only *where* to reach the
queue once enabled; it is not the switch, because it carries a default
of its own and so cannot distinguish "unset" from "deliberately pointed
at localhost".

For the base, `compose.plugins.yaml`, and `compose.dedicated-plugins.yaml`
layerings, adding `compose.activity.yaml` to the `-f` set flips the flag
to `true` for `fiftyone-app` and `teams-api`. No `.env` change is needed
for those two.

**Set `FIFTYONE_ACTIVITY_ENABLED=true` in `.env` if you run a
dedicated-plugins or delegated-operator stack.** The overlay cannot flip
`teams-plugins` or the `teams-do*` workers: those services are absent
from some `-f` sets, and naming a service in an overlay creates it rather
than annotating it, so the overlay would start containers a base stack
never asked for. They read the flag from `.env` instead, and without it
they stay off while `fiftyone-app` and `teams-api` are on. That gap
matters precisely where those services do the work — under the
dedicated-plugins layering the workflows plugin (the annotation and
review event producer) executes in `teams-plugins` rather than in
`fiftyone-app`, and under the delegated-operator overlays operator runs
execute in `teams-do`.

Setting it in `.env` is the belt-and-braces option for every layering: an
explicit value there wins over the overlay in both directions. You are
already editing `.env` to set `FIFTYONE_ACTIVITY_ORG_ID`, without which
the tenant-scoped read path returns nothing.

Verify what a given `-f` set resolves to before bringing it up:

```shell
docker compose -f compose.yaml -f compose.activity.yaml config \
  | grep FIFTYONE_ACTIVITY_ENABLED
```

Include the same `-f` set on every subsequent `docker compose` command
for the deployment. Omitting `compose.activity.yaml` on a later
`up -d` removes the two services from the project.

## Environment variables

Set these in your `.env` file. See the Activity Analytics section of
`env.template` for the same list with defaults.

| Variable                              | Default                            | Description                                                                                                                          |
| ------------------------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `FIFTYONE_ACTIVITY_ENABLED`           | `false`                            | The gate: nothing emits while false. `compose.activity.yaml` sets it for `fiftyone-app` and `teams-api`; set it here for the others. |
| `FIFTYONE_ACTIVITY_ORG_ID`            | empty                              | Organization id that activity events are scoped by. Set it for single-org deployments.                                               |
| `FIFTYONE_MQ_REDIS_URL`               | `redis://fiftyone-mq-redis:6379/0` | Queue connection string. Point it at an external Redis to replace the bundled service.                                               |
| `FIFTYONE_ACTIVITY_RETENTION_DAYS`    | `365`                              | Retention window for raw events. Rollups are kept indefinitely. `0` disables expiry.                                                 |
| `FIFTYONE_ACTIVITY_MAX_STORAGE_BYTES` | `10737418240`                      | Size cap on raw events. The prune worker removes oldest-first when exceeded. `0` disables.                                           |
| `FIFTYONE_MQ_REDIS_MAXMEMORY`         | `200mb`                            | `maxmemory` for the bundled queue Redis. Raise it if the queue backs up under load.                                                  |

`activity-worker` derives its MongoDB connection from
`FIFTYONE_DATABASE_URI` and `FIFTYONE_DATABASE_NAME`. You do not
configure it separately.

## Queue durability

The `fiftyone-mq-redis` service starts with `--appendonly yes` and
`--appendfsync everysec`, and mounts the named volume
`fiftyone-mq-redis-data` at `/data`. Redis appends every write to an
append-only file (AOF) there and flushes it once per second.

**What this covers.** Restarting or recreating the container, and
restarting the Docker host, no longer empty the queue. Redis replays the
AOF on startup and the `activity-worker` picks up where it left off.
Loss on an unclean stop is bounded to roughly the last second of writes.

**What this does not cover.** A persisted queue is not an end-to-end
durability guarantee, and there are still ways to lose events:

- The emit path buffers events in the emitting process. That buffer is
  cleared when an enqueue fails, so events dropped there never reach
  Redis and are not recoverable from the AOF.
- Enqueuing is not atomic with the change it describes. An operation can
  commit to MongoDB while its activity event fails to enqueue, so the
  queue is not a complete record of what happened.
- Removing the volume removes the queue. `docker compose down -v` and
  `docker volume rm` delete `fiftyone-mq-redis-data` along with any jobs
  that had not been consumed yet.

Treat Activity Analytics data in MongoDB as the durable copy — it is
covered by your existing MongoDB backups. The queue is the transport, and
AOF narrows one window in that transport rather than making it lossless.
Writing events to MongoDB directly and keeping only the work queue in
Redis is tracked as `FOEPD-4411`.

Confirm persistence is on:

```shell
docker compose exec fiftyone-mq-redis \
  redis-cli config get appendonly appendfsync
```

## The queue Redis requires `noeviction`

The bundled `fiftyone-mq-redis` service starts with
`--maxmemory-policy noeviction`. Keep it that way.

The queue is a BullMQ queue, and BullMQ stores queued jobs as ordinary
Redis keys. Under an evicting policy such as `allkeys-lru`, Redis
reclaims memory by deleting those keys, and queued jobs disappear
without an error on either side. Activity data goes missing with no
signal that anything failed.

With `noeviction`, a full queue rejects writes instead. The emit path
is best-effort, so a rejected write degrades to a dropped event and the
emitting operation is unaffected.

Note that the bundled `telemetry-redis` service does use `allkeys-lru`.
That is correct for telemetry, which stores expendable metric samples.
Do not copy that setting onto the queue Redis, and do not point
`FIFTYONE_MQ_REDIS_URL` at `telemetry-redis`.

If you supply an external Redis through `FIFTYONE_MQ_REDIS_URL`, verify
its policy and its persistence — the compose file only configures the
bundled service:

```shell
redis-cli -u "${FIFTYONE_MQ_REDIS_URL}" \
  config get maxmemory-policy appendonly appendfsync
```

## Keep `activity-worker` at one replica

Do not scale `activity-worker` above one replica.

The rollup, snapshot, and prune schedulers inside the container are
single-flight and hold no cross-container lock. A second replica runs
its own copy of each schedule, which produces duplicate rollup and
snapshot writes and lets two prune passes compete over the same
records.

The ingest worker is the only part of the pipeline that would benefit
from horizontal scale, and it is not separable from the others in this
image. A single container keeps up with the emit volume of a normal
deployment.

## Verifying

The commands below need `compose.activity.yaml` in the `-f` set, the
same as the `up -d` above. Without it compose does not know the two
services exist.

Confirm both services are running:

```shell
docker compose -f compose.yaml -f compose.activity.yaml \
  ps fiftyone-mq-redis activity-worker
```

Confirm the queue Redis has the expected policy:

```shell
docker compose exec fiftyone-mq-redis \
  redis-cli config get maxmemory-policy
```

Check the worker's logs for ingest activity:

```shell
docker compose logs activity-worker
```

In the app, the Audit Log page populates as events are ingested and
rolled up. An empty Audit Log with a healthy worker usually means
`FIFTYONE_ACTIVITY_ORG_ID` does not match the organization you are
viewing.
