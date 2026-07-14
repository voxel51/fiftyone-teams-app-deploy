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
  - [Sizing Guidance](#sizing-guidance)
- [Redirecting Compaction Scratch Space With `FIFTYONE_COMPACTION_TEMP_LOCATION`](#redirecting-compaction-scratch-space-with-fiftyone_compaction_temp_location)
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
Iceberg tables with Parquet files (rather than as fields directly on the
MongoDB sample document).
A background delegated-operator pipeline continuously ingests new data
and periodically compacts it into larger, size-bounded files.

Running multimodal datasets requires:

1. Setting the `VFF_MULTIMODAL` environment variable (feature flag) on every service that serves or
   processes multimodal data.
1. Providing sufficient disk space on `teams-do` (delegated operator) containers
   for projection compaction to succeed — optionally redirected to a
   mounted volume via the `FIFTYONE_COMPACTION_TEMP_LOCATION` environment variable.
1. Providing enough memory on `fiftyone-app` to serve multimodal grid queries, which
   run DuckDB in-process.
1. Optionally, setting `FIFTYONE_PROJECTION_DELEGATION_TARGET` to pin projection
   processing to a specific `teams-do` worker instead of relying on
   automatic selection.

## Enabling the `VFF_MULTIMODAL` Feature Flag

`VFF_MULTIMODAL` is a presence-only feature flag: any value enables it
(including `0` or an empty string) — only an unset variable disables it.
Set `VFF_MULTIMODAL=1` to enable multimodal support, which is off by
default.

### Services That Require The Flag

```yaml
services:
  teams-api:
    environment:
      VFF_MULTIMODAL: 1

  fiftyone-app:
    environment:
      VFF_MULTIMODAL: 1

  teams-do: # and any additional teams-do-2/teams-do-3/teams-do-gpu slots you run
    environment:
      VFF_MULTIMODAL: 1

  teams-app:
    environment:
      VFF_MULTIMODAL: 1

  teams-plugins:
    environment:
      VFF_MULTIMODAL: 1
```

## Delegated Operator Scratch Space Requirements

Compaction downloads and re-uploads Parquet files for each projection
table, staging them through the container's local `/tmp`.

### Sizing Guidance

In Docker Compose, the container's writable filesystem draws on whatever
free disk space is available on the Docker host. There's no artificial
per-directory cap to raise; the practical requirement is simply:

- Ensure the host machine running your `teams-do` container(s) has enough
  free disk space to stage the compaction's largest expected
  output file (1 GiB by default) plus the current backlog of
  not-yet-compacted source files for a projection table.
- If you customize your `teams-do` service to run with `read_only: true`,
  you must add your own writable mount at `/tmp` (e.g. a bind mount or a
  named volume) with enough capacity for the above — a read-only root
  filesystem with no `/tmp` mount will make compaction fail outright, not
  just run low on space.

## Redirecting Compaction Scratch Space With `FIFTYONE_COMPACTION_TEMP_LOCATION`

The projection pipeline temporarily stores files (downloaded and
re-uploaded Parquet, plus a local copy of the Iceberg catalog metadata)
under a single directory while compacting projection data. Set
`FIFTYONE_COMPACTION_TEMP_LOCATION` in each `teams-do` service's
`environment:` block to point that directory at a mounted volume instead
of relying on the container's default writable filesystem.

If unset, compaction stages files under the system temp directory (`/tmp`
inside the container), sized per "Delegated Operator Scratch Space
Requirements" above. The configured directory is created automatically if
it doesn't already exist.

### Example Volume Configuration

Add to (or create) a `compose.override.yaml`:

```yaml
services:
  teams-do:
    environment:
      FIFTYONE_COMPACTION_TEMP_LOCATION: /mnt/compaction-scratch/compaction
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
filters. DuckDB runs in memory and thus requires increasing the memory
allocated to the `fiftyone-app` service.

`fiftyone-app` serves requests with multiple Hypercorn workers, and each
worker runs its own DuckDB connection, so the container's memory is divided
among them. The approximate per-query ceiling is:

```text
ceiling ≈ (fiftyone-app memory limit × 0.8) ÷ number of Hypercorn workers
```

Disk spill is disabled, so a query that needs more than its ceiling returns
an empty filter widget (logged as an out-of-memory) rather than crashing
the container.

The worker count defaults to **4** (hardcoded in the `fiftyone-app` image).
At that default a 4G container leaves each query only ~0.8G. Whether
that's enough depends on your datasets and projections. As a concrete
example, 4 workers on a 500m CPU / 1.5G memory container — only ~0.3G
memory per query, which proved far too small for wide projections. Set
`HYPERCORN_WORKERS` in the service's `environment:` block to lower the
worker count and give each query more headroom. Fewer workers also reduce
the app's overall request concurrency, so balance it against your traffic.

### Recommended Starting Point

```yaml
services:
  fiftyone-app:
    deploy:
      resources:
        limits:
          cpus: "1"
          memory: 4G
        reservations:
          cpus: "1"
          memory: 4G
    environment:
      VFF_MULTIMODAL: 1
      HYPERCORN_WORKERS: 2
```

This gives ≈1.6G per DuckDB query (`4G × 0.8 ÷ 2`). If your widest
projections (many columns) return empty sidebar filters, raise
`fiftyone-app` memory or lower `HYPERCORN_WORKERS`.

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
which comes from the `-n <name>` argument in its `command:`. The
`teams-do-2`, `teams-do-3`, and `teams-do-gpu` slots in
`compose.delegated-operators.yaml`/`compose.delegated-operators.gpu.yaml`
already pass a fixed `-n <name>`, so any of them can be used as a
delegation target as-is. The default `teams-do` slot does not pass `-n`
and registers under an unpredictable name — add `-n teams-do` to its
`command:` if you want to target it directly instead.

### Example Configuration

```yaml
services:
  teams-api:
    environment:
      VFF_MULTIMODAL: 1
      FIFTYONE_PROJECTION_DELEGATION_TARGET: teams-do-2
```
