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
  - [Workloads That Require The Flag](#workloads-that-require-the-flag)
- [Delegated Operator Storage Requirements](#delegated-operator-storage-requirements)
  - [Minimum Recommended Sizing](#minimum-recommended-sizing)
  - [Example Storage Configuration](#example-storage-configuration)
- [Redirecting Compaction Scratch Space With `FIFTYONE_COMPACTION_TEMP_LOCATION`](#redirecting-compaction-scratch-space-with-fiftyone_compaction_temp_location)
  - [Example Volume Configuration](#example-volume-configuration)
- [`fiftyone-app` Memory Sizing](#fiftyone-app-memory-sizing)
  - [Recommended Starting Point](#recommended-starting-point)
- [Pinning Projection Processing With `FIFTYONE_PROJECTION_DELEGATION_TARGET`](#pinning-projection-processing-with-fiftyone_projection_delegation_target)
  - [Behavior When Set](#behavior-when-set)
  - [Example Delegation Target Configuration](#example-delegation-target-configuration)

<!-- tocstop -->

## Overview

FiftyOne Enterprise's multimodal datasets store large modalities associated
with each sample (e.g. sensor streams, point clouds, telemetry) as
Iceberg tables with Parquet files (rather than as fields directly on the
MongoDB sample document).
A background delegated-operator pipeline continuously ingests new data
and periodically compacts it into larger, size-bounded files.

Running multimodal datasets requires the following configuration beyond a
standard install:

1. Setting the `VFF_MULTIMODAL` environment variable (feature flag) on
   every workload that serves or processes multimodal data.
1. Providing sufficient ephemeral/scratch storage on delegated-operator
   workloads for projection compaction to succeed — optionally redirected
   to a mounted volume via the `FIFTYONE_COMPACTION_TEMP_LOCATION`
   environment variable.
1. Providing enough memory on `fiftyone-app` to serve multimodal grid
   queries, which run DuckDB in-process.
1. Optionally, setting `FIFTYONE_PROJECTION_DELEGATION_TARGET` to pin
   projection processing to a specific orchestrator instead of relying on
   automatic selection.

## Enabling the `VFF_MULTIMODAL` Feature Flag

`VFF_MULTIMODAL` is a presence-only feature flag: any value enables it
(including `0` or an empty string) — only an unset variable disables it.
Set `VFF_MULTIMODAL=1` to enable multimodal support, which is off by
default.

### Workloads That Require The Flag

```yaml
apiSettings:
  env:
    VFF_MULTIMODAL: 1

appSettings:
  env:
    VFF_MULTIMODAL: 1

delegatedOperatorDeployments:
  template:
    env:
      VFF_MULTIMODAL: 1

# and delegatedOperatorJobTemplates.template.env if you use on-demand
# (Kubernetes Job-based) delegated operators

teamsAppSettings:
  env:
    VFF_MULTIMODAL: 1

pluginsSettings:
  env:
    VFF_MULTIMODAL: 1
```

## Delegated Operator Storage Requirements

Compaction downloads and re-uploads Parquet files for each projection
table, staging them through the pod's local `/tmp`.

If your delegated-operator workloads run with
`securityContext.readOnlyRootFilesystem: true`
(recommended for Pod Security Admission `restricted` compliance), `/tmp`
is not writable at all unless you explicitly mount a writable volume there.
An `emptyDir` mounted at `/tmp` still counts against the pod's
`ephemeral-storage` limit; an explicit `sizeLimit` adds a second, per-volume
cap on that same usage — it does not grant capacity beyond the
`ephemeral-storage` budget. Compaction fails if it hits either cap, so size
both for the same target: whichever of the two is smaller is the effective
ceiling.

### Minimum Recommended Sizing

For any `delegatedOperatorDeployments`/`delegatedOperatorJobTemplates`
instance that will run the projection pipeline:

- `resources.limits`/`requests.ephemeral-storage`: at least `1.5Gi`
- A `tmpdir` `emptyDir` mounted at `/tmp` with `sizeLimit: 1Gi`

These are minimums for typical projection volumes, not a hard guarantee —
size up further if your projections accumulate a large backlog of
unconsolidated data between compaction runs (for example, if compaction has
been disabled or failing for a period of time).

### Example Storage Configuration

```yaml
delegatedOperatorDeployments:
  template:
    env:
      VFF_MULTIMODAL: 1
    resources:
      limits:
        ephemeral-storage: 1.5Gi
      requests:
        ephemeral-storage: 1.5Gi
    securityContext:
      readOnlyRootFilesystem: true
    volumeMounts:
      - name: tmpdir
        mountPath: /tmp
    volumes:
      - name: tmpdir
        emptyDir:
          sizeLimit: 1Gi
```

Apply the equivalent `env`, `resources`, `volumeMounts`, and `volumes`
settings under `delegatedOperatorJobTemplates.template` if you also use
on-demand (Kubernetes `Job`-based) delegated operators for projection
processing.

> [!NOTE]
> If your delegated-operator workloads do **not** set
> `readOnlyRootFilesystem: true`, `/tmp` is already part of the container's
> writable filesystem and draws directly from the pod's `ephemeral-storage`
> limit with no separate cap — you only need the explicit `tmpdir` volume
> and `sizeLimit` when the root filesystem is read-only.

## Redirecting Compaction Scratch Space With `FIFTYONE_COMPACTION_TEMP_LOCATION`

The projection pipeline temporarily stores files (downloaded and
re-uploaded Parquet, plus a local copy of the Iceberg catalog metadata)
under a single directory while compacting projection data. Set
`FIFTYONE_COMPACTION_TEMP_LOCATION` on
`delegatedOperatorDeployments.template.env` and/or
`delegatedOperatorJobTemplates.template.env` to point that directory at a
mounted volume instead of the `tmpdir` `emptyDir` described above.

If unset, compaction stages files under the system temp directory (`/tmp`
inside the container) — the `tmpdir` volume sized per "Delegated Operator
Storage Requirements" above. The configured directory is created
automatically if it doesn't already exist.

### Example Volume Configuration

```yaml
delegatedOperatorDeployments:
  template:
    env:
      FIFTYONE_COMPACTION_TEMP_LOCATION: /mnt/compaction-scratch/compaction
    volumeMounts:
      - name: compaction-scratch-vol
        mountPath: /mnt/compaction-scratch
    volumes:
      - name: compaction-scratch-vol
        persistentVolumeClaim:
          claimName: my-shared-compaction-scratch-pvc
```

Apply the equivalent `env`, `volumeMounts`, and `volumes` settings under
`delegatedOperatorJobTemplates.template` if you also use on-demand
(Kubernetes `Job`-based) delegated operators for projection processing.

> [!NOTE]
> The mounted volume must support concurrent writes from every
> delegated-operator replica that can run compaction (e.g. an NFS-backed
> `ReadWriteMany` PVC). A `ReadWriteOnce` volume only works if exactly one
> replica ever runs the projection pipeline.

## `fiftyone-app` Memory Sizing

Serving a multimodal grid runs DuckDB inside `fiftyone-app` — it reads the
projection Parquet/Iceberg tables to compute the grid and its sidebar
filters. DuckDB runs in memory and thus requires increasing the memory
allocated to the `fiftyone-app` pod.

`fiftyone-app` serves requests with multiple Hypercorn workers, and each
worker runs its own DuckDB connection, so the pod's memory is divided among
them. The approximate per-query ceiling is:

```text
ceiling ≈ (fiftyone-app memory limit × 0.8) ÷ number of Hypercorn workers
```

Disk spill is disabled, so a query that needs more than its ceiling returns
an empty filter widget (logged as an out-of-memory) rather than crashing
the pod.

The worker count defaults to **4** (hardcoded in the `fiftyone-app` image).
At that default a 4Gi pod leaves each query only ~0.8Gi. Whether that's
enough depends on your datasets and projections. As a concrete example, 4
workers on a 500m CPU / 1.5Gi pod — only ~0.3Gi per query — proved far too
small for wide projections. Set `HYPERCORN_WORKERS` in the app's
environment to lower the worker count and give each query more headroom.
Fewer workers also reduce the app's overall request concurrency, so
balance it against your traffic.

### Recommended Starting Point

```yaml
appSettings:
  env:
    VFF_MULTIMODAL: 1
    HYPERCORN_WORKERS: 2
  resources:
    limits:
      cpu: 1
      memory: 4Gi
    requests:
      cpu: 1
      memory: 4Gi
```

This gives ≈1.6Gi per DuckDB query (`4Gi × 0.8 ÷ 2`). If your widest
projections (many columns) return empty sidebar filters, raise
`fiftyone-app` memory or lower `HYPERCORN_WORKERS`. Keep `requests` equal
to `limits` so the scheduler reserves what DuckDB will actually use.

## Pinning Projection Processing With `FIFTYONE_PROJECTION_DELEGATION_TARGET`

By default, `teams-api` automatically selects the lowest-compute active
orchestrator (excluding GPU orchestrators) to run each dataset's projection
pipeline. Set `FIFTYONE_PROJECTION_DELEGATION_TARGET` on `apiSettings.env`
to pin all projection processing to one specific orchestrator instead —
for example, to ensure it always lands on the orchestrator instance sized
for compaction per the previous section.

The value must be the exact registered name of an always-on delegated
operator instance — i.e. the kebab-cased key under
`delegatedOperatorDeployments.deployments.<key>` (the same name each
instance registers with via `fiftyone delegated launch -n <name>`).
On-demand (`delegatedOperatorJobTemplates`) instances cannot be used as a
delegation target, since they don't run as a persistent, named orchestrator.

### Behavior When Set

There is **no fallback to automatic selection**. If the configured value
does not match an active orchestrator capable of running the projection
operator, `teams-api` logs an error and skips queuing any projection
delegated operations that cycle — pending datasets simply won't be
processed until the value is corrected or removed.

### Example Delegation Target Configuration

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDoCpuDefault:
      enabled: true
      # ... sized per "Delegated Operator Storage Requirements" above

apiSettings:
  env:
    FIFTYONE_PROJECTION_DELEGATION_TARGET: teams-do-cpu-default
```
