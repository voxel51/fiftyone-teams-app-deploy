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

# Configuring Multimodal Datasets

<!-- toc -->

- [Overview](#overview)
- [Enabling the `VFF_MULTIMODAL` Feature Flag](#enabling-the-vff_multimodal-feature-flag)
  - [Services That Require The Flag](#services-that-require-the-flag)
- [Delegated Operator Scratch Space Requirements](#delegated-operator-scratch-space-requirements)
  - [Why Multimodal Needs Extra Scratch Space](#why-multimodal-needs-extra-scratch-space)
  - [Sizing Guidance](#sizing-guidance)
- [Redirecting Compaction Scratch Space With `FIFTYONE_COMPACTION_TEMP_LOCATION`](#redirecting-compaction-scratch-space-with-fiftyone_compaction_temp_location)
  - [Default Behavior](#default-behavior)
  - [Example Volume Configuration](#example-volume-configuration)
- [`fiftyone-app` Memory Sizing](#fiftyone-app-memory-sizing)
  - [Recommended Starting Point](#recommended-starting-point)
- [Pinning Projection Processing With `FIFTYONE_PROJECTION_DELEGATION_TARGET`](#pinning-projection-processing-with-fiftyone_projection_delegation_target)
  - [Behavior When Set](#behavior-when-set)
  - [Naming Your `teams-do` Worker](#naming-your-teams-do-worker)
  - [Example Configuration](#example-configuration)

<!-- tocstop -->

## Overview

FiftyOne Enterprise's multimodal datasets store large modalities associated
with each sample (e.g. sensor streams, point clouds, telemetry) as
Parquet-backed Iceberg tables rather than as fields directly on the
Mongo-backed sample document.
A background delegated-operator pipeline (`run_projections` followed by
`compact_projections`) continuously ingests new data and periodically
compacts it into larger, size-bounded files.

Running multimodal datasets requires:

1. The `VFF_MULTIMODAL` feature flag, set on every service that serves or
   processes multimodal data.
1. Enough scratch disk space on `teams-do` (delegated operator) containers
   for projection compaction to succeed — optionally redirected to a
   mounted volume via `FIFTYONE_COMPACTION_TEMP_LOCATION`.
1. Enough memory on `fiftyone-app` to serve multimodal grid queries, which
   run DuckDB in-process.
1. Optionally, `FIFTYONE_PROJECTION_DELEGATION_TARGET` to pin projection
   processing to a specific `teams-do` worker instead of relying on
   automatic selection.

## Enabling the `VFF_MULTIMODAL` Feature Flag

`VFF_MULTIMODAL` is a presence-only feature flag: setting it to any value
(e.g. `1`) enables the feature. Multimodal support is off by default.

### Services That Require The Flag

Add `VFF_MULTIMODAL` to the `environment:` block of each of the following
services:

| Compose service | Why it's needed |
| --- | --- |
| `teams-api` | Runs the periodic background task that queues projection delegated operations for datasets with pending data. |
| `fiftyone-app` | Serves the GraphQL/REST routes and grid queries that read multimodal Parquet data; without the flag, requests against a multimodal dataset are rejected. |
| `teams-do` (and any additional worker slots/GPU workers you run) | Runs `run_projections`/`compact_projections`, which write multimodal dataset metadata and raise an error if the flag isn't enabled. |
| `teams-app` (frontend) | Gates rendering of multimodal-specific UI components. |
| `teams-plugins` | Serves operator schema/input resolution for the projection pipeline; recommended for consistency even though projection execution itself only ever runs on `teams-do` workers. |

> [!NOTE]
> Set the flag on **every** service in the table above. A partial
> configuration (e.g. only `fiftyone-app`) leaves the UI able to query
> multimodal data while ingestion silently fails, or vice versa.

## Delegated Operator Scratch Space Requirements

### Why Multimodal Needs Extra Scratch Space

`compact_projections` reads all not-yet-compacted Parquet files for a
projection table, merges/sorts them, and writes back a consolidated file, up
to a configurable target size (1 GiB by default). For cloud warehouse
locations (`gs://`, `s3://`, `az://`), both the download of source files and
the write of the compacted output stage through the container's local
`/tmp` before being uploaded.

### Sizing Guidance

Unlike a Kubernetes deployment, the provided Docker Compose files for
`teams-do` don't set `read_only: true` and don't mount a separate,
size-limited volume at `/tmp` — it's just part of the container's normal
writable filesystem, which draws on whatever free disk space is actually
available on the Docker host. There's no artificial per-directory cap to
raise here; the practical requirement is simply:

- Ensure the host machine running your `teams-do` container(s) has enough
  free disk space to comfortably stage compaction's largest expected
  output file (1 GiB by default) plus the current backlog of
  not-yet-compacted source files for a projection table.
- If you customize your `teams-do` service to run with `read_only: true`,
  you must add your own writable mount at `/tmp` (e.g. a bind mount or a
  named volume) with enough capacity for the above — a read-only root
  filesystem with no `/tmp` mount will make compaction fail outright, not
  just run low on space.

## Redirecting Compaction Scratch Space With `FIFTYONE_COMPACTION_TEMP_LOCATION`

`compact_projections` stages all of its local scratch files — downloaded
and re-uploaded Parquet, plus a local copy of the Iceberg catalog metadata —
under a single directory. Set `FIFTYONE_COMPACTION_TEMP_LOCATION` in each
`teams-do` service's `environment:` block to point that directory at a
mounted volume instead of relying on the container's default writable
filesystem.

### Default Behavior

If unset, compaction stages files under the system temp directory (`/tmp`
inside the container), sized per "Delegated Operator Scratch Space
Requirements" above. The configured directory is created automatically if
it doesn't already exist.

### Example Volume Configuration

```yaml
services:
  teams-do-multimodal:
    image: voxel51/fiftyone-teams-cv-full:v2.22.0
    command: >
      /bin/sh -c "fiftyone delegated launch -t remote -m -n teams-do-multimodal"
    environment:
      VFF_MULTIMODAL: 1
      FIFTYONE_COMPACTION_TEMP_LOCATION: /mnt/compaction-scratch/compaction
      # ... plus the other teams-do environment variables
    volumes:
      - compaction-scratch:/mnt/compaction-scratch

volumes:
  compaction-scratch:
    driver: local
    driver_opts:
      type: nfs
      o: addr=<nfs-host>,rw
      device: ":/path/to/export"
```

> [!NOTE]
> The mounted volume must support concurrent writes from every `teams-do`
> worker that can run compaction. A plain local Docker volume only works if
> exactly one worker (on one host) ever runs the projection pipeline; use
> an NFS-backed volume (as shown above) or another shared filesystem for
> multi-worker or multi-host setups.

## `fiftyone-app` Memory Sizing

Serving a multimodal grid runs DuckDB inside `fiftyone-app` — it reads the
projection Parquet/Iceberg tables to compute the grid and its sidebar
filters. This is a **memory** requirement on `fiftyone-app`, separate from
the `teams-do` scratch space above.

`fiftyone-app` limits each DuckDB query to the memory available to the
container automatically — there is no DuckDB memory setting to configure.
`fiftyone-app` serves requests with multiple Hypercorn workers, and each
worker runs its own DuckDB connection, so the container's memory is divided
among them. The approximate per-query ceiling is:

```text
ceiling ≈ (fiftyone-app memory limit × 0.8) ÷ number of Hypercorn workers
```

Disk spill is disabled, so a query that needs more than its ceiling returns
an empty filter widget (logged as an out-of-memory) rather than crashing
the container.

The worker count defaults to **4** (an `ENV` baked into the `fiftyone-app`
image, so it's the runtime default). At that default a 4Gi container leaves
each query only ~0.8Gi. Whether that's enough depends entirely on your
datasets and projections; as a concrete example, 4 workers on a 500m CPU /
1.5Gi container — only ~0.3Gi per query — proved far too small for wide
projections. Set `HYPERCORN_WORKERS` in the service's `environment:` block
to lower the worker count and give each query more headroom. Fewer workers
also reduce the app's overall request concurrency, so balance it against
your traffic.

### Recommended Starting Point

```yaml
services:
  fiftyone-app:
    mem_limit: 4g
    cpus: "1"
    environment:
      VFF_MULTIMODAL: 1
      HYPERCORN_WORKERS: 2
```

This gives ≈1.6Gi per DuckDB query (`4Gi × 0.8 ÷ 2`). If your widest
projections (many columns) return empty sidebar filters, raise
`fiftyone-app` memory or lower `HYPERCORN_WORKERS` further.

## Pinning Projection Processing With `FIFTYONE_PROJECTION_DELEGATION_TARGET`

By default, `teams-api` automatically selects the lowest-compute active
orchestrator (excluding GPU orchestrators) to run each dataset's projection
pipeline. Set `FIFTYONE_PROJECTION_DELEGATION_TARGET` in `teams-api`'s
`environment:` block to pin all projection processing to one specific
`teams-do` worker instead — for example, to ensure it always lands on the
worker sized for compaction per the previous section.

### Behavior When Set

There is **no fallback to automatic selection**. If the configured value
does not match an active orchestrator capable of running the projection
operator, `teams-api` logs an error and skips queuing any projection
delegated operations that cycle — pending datasets simply won't be
processed until the value is corrected or removed.

### Naming Your `teams-do` Worker

The value must match the exact name a `teams-do` worker registers under,
which comes from the `-n <name>` argument in its `command:`. The default
`teams-do` slot in `compose.delegated-operators.yaml` does **not** pass
`-n`, so it registers under an unpredictable default name — if you want to
pin projection processing to a worker, add an explicit `-n <name>` to its
`command:` (as the `teams-do-2`/`teams-do-3`/`teams-do-gpu` slots already
do) rather than relying on the default slot.

### Example Configuration

```yaml
services:
  teams-do-multimodal:
    image: voxel51/fiftyone-teams-cv-full:v2.22.0
    command: >
      /bin/sh -c "fiftyone delegated launch -t remote -m -n teams-do-multimodal"
    environment:
      VFF_MULTIMODAL: 1
      # ... plus the other teams-do environment variables

  teams-api:
    environment:
      VFF_MULTIMODAL: 1
      FIFTYONE_PROJECTION_DELEGATION_TARGET: teams-do-multimodal
```
