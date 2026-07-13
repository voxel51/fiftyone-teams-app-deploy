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
  - [Why Multimodal Needs Extra Scratch Space](#why-multimodal-needs-extra-scratch-space)
  - [Minimum Recommended Sizing](#minimum-recommended-sizing)
  - [Example Storage Configuration](#example-storage-configuration)
- [Pinning Projection Processing With `FIFTYONE_PROJECTION_DELEGATION_TARGET`](#pinning-projection-processing-with-fiftyone_projection_delegation_target)
  - [Behavior When Set](#behavior-when-set)
  - [Example Delegation Target Configuration](#example-delegation-target-configuration)

<!-- tocstop -->

## Overview

FiftyOne Enterprise's multimodal datasets store large, non-sample-centric
modalities (e.g. sensor streams, point clouds, telemetry) as Parquet-backed
Iceberg tables rather than as ordinary Mongo-backed samples.
A background delegated-operator pipeline (`run_projections` followed by
`compact_projections`) continuously ingests new data and periodically
compacts it into larger, size-bounded files.

Running multimodal datasets requires three pieces of configuration beyond a
standard install:

1. The `VFF_MULTIMODAL` feature flag, set on every workload that serves or
   processes multimodal data.
1. Enough ephemeral/scratch storage on delegated-operator workloads for
   projection compaction to succeed.
1. Optionally, `FIFTYONE_PROJECTION_DELEGATION_TARGET` to pin projection
   processing to a specific orchestrator instead of relying on automatic
   selection.

## Enabling the `VFF_MULTIMODAL` Feature Flag

`VFF_MULTIMODAL` is a presence-only feature flag: setting it to any value
(e.g. `1`) enables the feature.
Multimodal support is off by default.

### Workloads That Require The Flag

| Workload | Helm values path | Why it's needed |
| --- | --- | --- |
| `teams-api` | `apiSettings.env` | Runs the periodic background task that queues projection delegated operations for datasets with pending data. |
| `fiftyone-app` | `appSettings.env` | Serves the GraphQL/REST routes and grid queries that read multimodal Parquet data; without the flag, requests against a multimodal dataset are rejected. |
| Delegated operator workloads | `delegatedOperatorDeployments.template.env` and/or `delegatedOperatorJobTemplates.template.env` | Runs `run_projections`/`compact_projections`, which write multimodal dataset metadata and raise an error if the flag isn't enabled. |
| `teams-app` (frontend) | `teamsAppSettings.env` | Gates rendering of multimodal-specific UI components. |
| `teams-plugins` | `pluginsSettings.env` | Serves operator schema/input resolution for the projection pipeline; recommended for consistency even though projection execution itself only ever runs on delegated-operator workloads. |

> [!NOTE]
> Set the flag on **every** workload in the table above. A partial
> configuration (e.g. only `fiftyone-app`) leaves the UI able to query
> multimodal data while ingestion silently fails, or vice versa.

## Delegated Operator Storage Requirements

### Why Multimodal Needs Extra Scratch Space

`compact_projections` reads all not-yet-compacted Parquet files for a
projection table, merges/sorts them, and writes back a consolidated file, up
to a configurable target size (1 GiB by default). For cloud warehouse
locations (`gs://`, `s3://`, `az://`), both the download of source files and
the write of the compacted output stage through the pod's local `/tmp`
before being uploaded.

If your delegated-operator workloads run with
`securityContext.readOnlyRootFilesystem: true`
(recommended for Pod Security Admission `restricted` compliance), `/tmp`
is not writable at all unless you explicitly mount a writable volume there.
An `emptyDir` volume mounted at `/tmp` without its own `sizeLimit` draws
from the pod's overall `ephemeral-storage` resource budget; an `emptyDir`
with an explicit `sizeLimit` is capped separately, on top of that shared
budget. A `sizeLimit` that's too small for the amount of unconsolidated
data will cause compaction to be evicted or fail before it can complete,
even if the pod's aggregate `ephemeral-storage` limit is otherwise generous.

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
    teamsDoMultimodal:
      # ... sized per "Delegated Operator Storage Requirements" above

apiSettings:
  env:
    FIFTYONE_PROJECTION_DELEGATION_TARGET: teams-do-multimodal
```
