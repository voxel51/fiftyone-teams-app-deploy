<!-- markdownlint-disable MD013 -->
# Configuring Activity Core

Activity Core is the basis for activity tracking across FiftyOne
Enterprise. It records what happens in a deployment — operator runs,
annotation and review decisions, sample and label mutations — and rolls
those events up into the data that the features built on it read:
annotation metrics, the Audit Log, the Jobs pages, and more to come.

It is a substrate, not a feature in its own right. Enabling it is what
makes those surfaces have anything to show.

<!-- toc -->

- [Overview](#overview)
- [Enabling Activity Core](#enabling-activity-core)
- [Queue durability](#queue-durability)
- [Using an external Redis](#using-an-external-redis)
- [Multi-organization deployments](#multi-organization-deployments)
- [Viewing the data](#viewing-the-data)
- [Resource impact](#resource-impact)

<!-- tocstop -->

## Overview

Two moving parts:

- **Producers** — `teams-api`, `fiftyone-app`, `teams-plugins`, and the
  delegated operators. They enqueue events onto a Redis-backed queue
  (`fiftyone.mq`) as work happens.
- **Workers** — the `activity-*-worker` `Deployment`s. They drain the
  queue and write the `activity_*` collections in the deployment's own
  FiftyOne database, then maintain the rollups.

Both halves are off by default. The chart renders no activity
environment on the producers and no worker `Deployment`s until you turn
it on.

## Enabling Activity Core

Two settings, and they are a pair:

```yaml
# values.yaml
activitySettings:
  enabled: true
fiftyoneMq:
  enabled: true
```

`activitySettings.enabled` renders `FIFTYONE_ACTIVITY_ENABLED=true` on
the producer workloads and creates the worker `Deployment`s.
`fiftyoneMq.enabled` renders the queue Redis and sets
`FIFTYONE_MQ_REDIS_URL` on everything that talks to it.

Enabling activity without the queue is rejected at render time — the
workers and producers would have nothing to connect to, and every event
would be dropped while the install looked healthy:

```text
Error: execution error at (fiftyone-teams-app/templates/api-deployment.yaml):
activitySettings.enabled is true but fiftyoneMq.enabled is false: ...
```

The reverse is allowed. The queue being reachable is not consent to
emit, which is why enablement is a flag rather than an inference from
`FIFTYONE_MQ_REDIS_URL`.

## Queue durability

The bundled queue Redis runs with `--appendonly yes --appendfsync
everysec` and `--maxmemory-policy noeviction`.

`fiftyoneMq.redis.persistence.enabled` defaults to `true`, backing the
append-only file with a `PersistentVolumeClaim` so queued events survive
a pod reschedule (node drain, eviction, `kubectl delete pod`, chart
upgrade). This requires a default `StorageClass`; set
`persistence.storageClass` or `persistence.existingClaim` if you do not
have one.

Disabling persistence puts the AOF on an `emptyDir`. It then survives a
container restart but not a pod reschedule:

```yaml
fiftyoneMq:
  redis:
    persistence:
      enabled: false
```

When enabling persistence alongside a non-empty
`fiftyoneMq.redis.podSecurityContext`, set `fsGroup` to the image's
redis GID so the mounted `/data` stays writable.

This bounds one loss window rather than making the pipeline durable end
to end: the in-process emit buffer is cleared when an enqueue fails, and
enqueuing is not atomic with the domain change the event describes.

## Using an external Redis

```yaml
fiftyoneMq:
  enabled: true
  redis:
    external:
      url: redis://my-redis.example.com:6379/0
```

Setting `external.url` suppresses the bundled Redis `Deployment`,
`Service`, and `PersistentVolumeClaim`; the workloads are wired to your
instance instead.

**The external instance must use `maxmemory-policy noeviction`.** The
queue stores jobs as ordinary Redis keys, so an evicting policy such as
`allkeys-lru` deletes queued work with no error on either side —
activity goes missing with nothing to indicate it failed. Verify:

```shell
redis-cli -u "${FIFTYONE_MQ_REDIS_URL}" \
  config get maxmemory-policy appendonly appendfsync
```

## Multi-organization deployments

`activitySettings.orgId` attributes the deployment-wide rollups to an
organization. Leave it empty for a single-organization deployment — the
workers discover the organization themselves.

Set it only when the deployment holds several organizations, where the
deployment-wide counts cannot be attributed to one of them. The workers
log which organization to name.

This affects only the deployment-wide rollups. Per-request events are
always stamped with the authenticated organization carried on the
request.

## Viewing the data

Enabling Activity Core turns on collection only. The surfaces that
display the data are internal debug views, off by default in every
deployment, and the chart never sets their flags. Opt in per environment
through `teamsAppSettings.env`:

```yaml
teamsAppSettings:
  env:
    # workflow Activity tab + label History panel
    VFF_WF_ACTIVITY: true
    # pre-release Metrics tab
    VFF_WF_METRIC: true
```

Set only the ones you want; each flag is independent of the other and of
`activitySettings.enabled`.

## Resource impact

Each worker `Deployment` requests `100m` CPU / `128Mi` memory and is
limited to `500m` CPU / `512Mi` memory by default
(`activitySettings.resources`).

The bundled queue Redis adds one more pod, with `maxmemory` defaulting
to `200mb` (`fiftyoneMq.redis.maxmemory`). Raise it if the queue backs
up under load.

Do not scale the rollup, snapshot, or prune workers above one replica.
Their schedulers are single-flight and hold no cross-pod lock, so a
second replica produces duplicate rollup and snapshot writes and lets
two prune passes compete over the same records.
