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

# Upgrading FiftyOne Enterprise

<!-- toc -->

- [Upgrading From Previous Versions](#upgrading-from-previous-versions)
  - [A Note On Database Migrations](#a-note-on-database-migrations)
  - [From FiftyOne Enterprise Version 2.0.0 and Later](#from-fiftyone-enterprise-version-200-and-later)
    - [FiftyOne Enterprise v2.26+ Compose Files Moved to `docker/`](#fiftyone-enterprise-v226-compose-files-moved-to-docker)
    - [FiftyOne Enterprise v2.26+ Invitations and Service Accounts](#fiftyone-enterprise-v226-invitations-and-service-accounts)
    - [FiftyOne Enterprise v2.25+ Activity Analytics](#fiftyone-enterprise-v225-activity-analytics)
    - [FiftyOne Enterprise v2.23+ Service Orchestrators](#fiftyone-enterprise-v223-service-orchestrators)
    - [FiftyOne Enterprise v2.22+ Multimodal Datasets](#fiftyone-enterprise-v222-multimodal-datasets)
    - [FiftyOne Enterprise v2.19+ Telemetry Sidecars](#fiftyone-enterprise-v219-telemetry-sidecars)
      - [Host Requirements](#host-requirements)
      - [Opting out of Telemetry](#opting-out-of-telemetry)
      - [Scaling delegated-operator workers](#scaling-delegated-operator-workers)
    - [FiftyOne Enterprise v2.16+ Additional API Routes](#fiftyone-enterprise-v216-additional-api-routes)
    - [FiftyOne Enterprise v2.15+ Additional API Routes](#fiftyone-enterprise-v215-additional-api-routes)
    - [FiftyOne Enterprise v2.7+ Delegated Operator Changes](#fiftyone-enterprise-v27-delegated-operator-changes)
    - [FiftyOne Enterprise v2.5+ Delegated Operator Changes](#fiftyone-enterprise-v25-delegated-operator-changes)
    - [FiftyOne Enterprise v2.2+ Delegated Operator Changes](#fiftyone-enterprise-v22-delegated-operator-changes)
    - [Delegated Operation Capacity](#delegated-operation-capacity)
    - [Existing Orchestrators](#existing-orchestrators)
  - [From FiftyOne Enterprise Versions 1.6.0 to 1.7.1](#from-fiftyone-enterprise-versions-160-to-171)
  - [From FiftyOne Enterprise Version 1.1.0 and Before Version 1.6.0](#from-fiftyone-enterprise-version-110-and-before-version-160)
  - [From Before FiftyOne Enterprise Version 1.1.0](#from-before-fiftyone-enterprise-version-110)
  - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)

<!-- tocstop -->

## Upgrading From Previous Versions

Voxel51 assumes you use the published Docker compose files to deploy your
FiftyOne Enterprise environment.
If you use custom deployment mechanisms, carefully review the changes in the
[Docker Compose Files](../)
and update your deployment accordingly.

### A Note On Database Migrations

The environment variable `FIFTYONE_DATABASE_ADMIN`
controls whether the database may be migrated.
This is a safety check to prevent automatic database
upgrades that will break other users' SDK connections.
When false (or unset), either an error will occur

```shell
$ fiftyone migrate --all
Traceback (most recent call last):
...
OSError: Cannot migrate database from v0.22.0 to v0.22.3 when database_admin=False.
```

or no action will be taken:

```shell
$ fiftyone migrate --info
FiftyOne Enterprise version: 0.14.4
FiftyOne compatibility version: 0.22.3
Other compatible versions: >=0.19,<0.23
Database version: 0.21.2
dataset     version
----------  ---------
quickstart  0.22.0
$ fiftyone migrate --all
$ fiftyone migrate --info
FiftyOne Enterprise version: 0.14.4
FiftyOne compatibility version: 0.23.0
Other compatible versions: >=0.19,<0.23
Database version: 0.21.2
dataset     version
----------  ---------
quickstart  0.21.2
```

### From FiftyOne Enterprise Version 2.0.0 and Later

1. [Upgrade to FiftyOne Enterprise version 2.26.0](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Enterprise SDK users to FiftyOne Enterprise
   version 2.26.0
    1. Log in to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Voxel51 recommends that you upgrade all your datasets.

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at the latest version, run

   ```shell
   fiftyone migrate --info
   ```

#### FiftyOne Enterprise v2.26+ Compose Files Moved to `docker/`

FiftyOne Enterprise v2.26.0 removes the `legacy` authentication mode
and the `FIFTYONE_AUTH_MODE` environment variable.
It also removes the `internal-auth` and `legacy-auth` directories
and `common-services.yaml`.
The compose files now live in the `docker` directory:

- `compose.yaml` defines every base service.
  Run every `docker compose` command with `-f compose.yaml` first.
- `compose.plugins.yaml`, `compose.dedicated-plugins.yaml`,
  and the other files add to `compose.yaml`.
  They no longer replace it.
- `compose.yaml` names the Compose project `voxel51`.
  Container and volume names start with `voxel51`
  instead of the directory name.
- `compose.activity.yaml` is removed.
  Activity Analytics is part of `compose.yaml`.

Because the project name changes,
Compose creates a new, empty plugins volume.
`migrate-volumes.sh` copies your plugins from the old volume into the new one.

If you already set your own project name,
with a top-level `name:` in `compose.override.yaml`,
`COMPOSE_PROJECT_NAME` in `.env`,
or `-p` on the command line,
Compose keeps that name.
It reuses your existing volumes,
so your plugins and other data stay where they are,
and `migrate-volumes.sh` has nothing to copy.
Compose doesn't remember `-p` between commands.
If you use it,
pass the same `-p` to every command in the steps below,
including `down`, `config`, `pull`, `up`,
and the options after `--` for `migrate-volumes.sh`.

> [!WARNING]
> Named volumes that you added in `compose.override.yaml`,
> such as a MongoDB data volume,
> also start empty under the new project name.
> To keep using an old volume,
> set its full name in `compose.override.yaml`
> and mark it `external`,
> so Compose uses it as it is.
> For example,
> for a volume named `mongodb-data` that you deployed from `internal-auth`:
>
> ```yaml
> volumes:
>   mongodb-data:
>     name: internal-auth_mongodb-data
>     external: true
> ```
>
> Run `docker volume ls` to find your old volume names.

Other things named after the Compose project also change with it:

- **Default network:**
  Compose creates `voxel51_default`
  in place of `internal-auth_default` or `legacy-auth_default`,
  usually on a new subnet.
  Containers outside the stack that attach to the old network,
  such as a reverse proxy
  or another Compose project that lists it as `external`,
  lose their connection,
  and firewall rules written for the old subnet stop matching.
  To keep the old network name,
  set it in `compose.override.yaml`:

  ```yaml
  networks:
    default:
      name: internal-auth_default
  ```

  Disconnect those outside containers before you stop the old stack,
  so Compose can remove the old network.
  Reconnect them after the new stack starts.
- **Container names:**
  `internal-auth-teams-api-1` becomes `voxel51-teams-api-1`.
  Update scripts, cron jobs, and monitoring that use container names.
- **Project label:**
  the `com.docker.compose.project` label changes to `voxel51`.
  Update log and monitoring filters that use it,
  and commands that pass `-p`.
- **Directory:**
  commands run from `docker/`.
  Update systemd units, deploy scripts, and runbooks
  that change into `docker/internal-auth` or `docker/legacy-auth`.

Settings in `compose.override.yaml`,
such as network options or extra services,
carry over because you copy the file.
Published host ports don't change.

To upgrade:

1. Before you get the v2.26 files,
   stop the stack from your old directory
   (`internal-auth` or `legacy-auth`).
   `--remove-orphans` also stops services from your other `-f` files.
   Don't add `-v`,
   so your volumes stay.

   ```shell
   docker compose down --remove-orphans
   ```

   If you turned on Activity Analytics in v2.25 with `compose.activity.yaml`,
   events still in its queue when the old stack stops aren't carried over.
   The new project starts with an empty `fiftyone-mq-redis-data` volume.
   To lose as few events as possible,
   stop the old stack when no one is using the App or running operations.
1. Get the v2.26 files.
1. Copy your `compose.override.yaml` and `.env` files
   into the `docker` directory,
   then `cd` into it.
   Don't copy these Compose files from your old directory,
   because the v2.26 files in the `docker` directory replace them:

    - `compose.yaml`
    - `compose.plugins.yaml`
    - `compose.dedicated-plugins.yaml`
    - `compose.delegated-operators.yaml`
    - `compose.delegated-operators.gpu.yaml`
    - `compose.agenticlabeler.yaml`
    - `compose.activity.yaml`

1. In the `.env` and `compose.override.yaml` files,
   update relative paths
   (such as `LOCAL_LICENSE_FILE_DIR` and bind mount sources)
   so they work from the `docker` directory.
   Compose resolves them from the `docker` directory now,
   and it creates an empty directory
   for a bind mount source that's missing.
1. If you deployed from `legacy-auth`, in the `.env` file:
    1. Rename `AUTH0_BASE_URL` to `BASE_URL`
    1. Remove `AUTH0_SECRET`. It's no longer used.
1. Remove `FIFTYONE_AUTH_MODE` from the `.env` and `compose.override.yaml` files
   if it's set.
1. In `compose.override.yaml`,
   remove any `extends:` block that points at `common-services.yaml`.
   Keep the settings you override under the service name.
   For example, change

   ```yaml
   services:
     teams-api:
       extends:
         file: ../common-services.yaml
         service: teams-api-common
       environment:
         MY_SETTING: my-value
   ```

   to

   ```yaml
   services:
     teams-api:
       environment:
         MY_SETTING: my-value
   ```

1. Update your `docker compose` command.
   Start it with `-f compose.yaml`,
   keep your other `-f` files,
   and remove `-f compose.activity.yaml`.
   For example, a deployment with dedicated plugins and delegated operators
   changes
   `-f compose.dedicated-plugins.yaml -f compose.delegated-operators.yaml -f compose.override.yaml`
   to

   ```shell
   docker compose \
     -f compose.yaml \
     -f compose.dedicated-plugins.yaml \
     -f compose.delegated-operators.yaml \
     -f compose.override.yaml \
     config
   ```

   Check that the output has the services and settings you expect.

1. Pull the v2.26 images with the same `-f` options and `pull`.
   On air-gapped hosts, load the images instead.
1. Copy your plugins into the new project.
   Pass your `-f` files after `--`.
   Check the plan first with `--dry-run`,
   then run it without:

   ```shell
   ./migrate-volumes.sh \
     --dry-run \
     -- \
     -f compose.yaml \
     -f compose.dedicated-plugins.yaml \
     -f compose.delegated-operators.yaml \
     -f compose.override.yaml
   ```

   The script finds the old plugins volume by its Compose labels.
   It skips the copy when:
    - the new project's plugins volume already has files,
      for example because you kept your project name
    - your `-f` files don't include a plugins file
    - no `internal-auth` or `legacy-auth` project has a plugins volume

   It picks a source on its own only from the `internal-auth` or `legacy-auth` project.
   If both have one,
   or your old deployment used a different project name,
   name it with `--from-project`.
   The skip message lists other projects on the host that have a plugins volume.
1. Start the stack with the same `-f` options and `up -d`.
1. Once the new stack works, remove the old plugins volume.
   The script prints the command.
   Keep your other old volumes
   until you've checked the data in the new stack.

#### FiftyOne Enterprise v2.26+ Invitations and Service Accounts

FiftyOne Enterprise v2.26.0 changes how `teams-cas` stores invitations
and service accounts.

**Upgrading needs no extra step.**
The first time a v2.26 `teams-cas` starts,
it updates the stored data before it serves requests:

- Each pending invitation becomes a pending user.
  That user holds a seat of the invited role
  until someone accepts or revokes the invitation,
  or the invitation expires.
- Service accounts move into their own collection,
  and their group memberships move with them.

`teams-cas` can safely repeat each step,
so several `teams-cas` replicas can start at the same time.
If the update fails,
`teams-cas` exits and tries again the next time it starts.

**What changes for administrators:**

- Pending invitations count toward the license's seat limits.
  When an organization reaches its limit for a role,
  `teams-cas` refuses new invitations for that role,
  the same way it refuses new users.
  An expired invitation frees its seat.
- Invitations are on by default
  for organizations that never changed the setting.

**Rolling back to v2.25 needs one extra step,**
because a v2.25 `teams-cas` can't read the data in its v2.26 form.
Until you do that step,
v2.25 doesn't list service accounts
and counts too many seats for pending invitations.

1. From the `docker` directory,
   stop v2.26 with the `-f` options you use for `up`.
   Don't add `-v`, so your volumes stay.
   For example:

   ```shell
   docker compose \
     -f compose.yaml \
     -f compose.plugins.yaml \
     -f compose.override.yaml \
     down
   ```

   Then run the same command with `ps --all` in place of `down`,
   and check that it lists no containers.
   If a v2.26 `teams-cas` starts,
   it converts the data to the v2.26 form again.
1. Go back to the v2.25.2 files.
   In a git clone of this repository, run:

   ```shell
   git checkout fiftyone-teams-app-2.25.2
   ```

   Otherwise, download and unpack
   [the v2.25.2 files](https://github.com/voxel51/fiftyone-teams-app-deploy/archive/refs/tags/fiftyone-teams-app-2.25.2.tar.gz).
1. Start v2.25.2 from the directory you used before the upgrade,
   `docker/internal-auth` or `docker/legacy-auth`,
   with the `-f` options you used for v2.25.2.
   That directory still has the `.env` and `compose.override.yaml`
   from before the upgrade.
   - Compose uses that directory's project name again,
     so v2.25.2 uses the volumes it had before the upgrade.
     Changes made after the upgrade,
     such as plugins you installed,
     are only in the `voxel51_*` volumes.
   - If you set your own project name,
     v2.25.2 uses the same volumes as v2.26.
1. Get a `fiftyone-migrator` release with the `rollback` command
   from Voxel51 support.
   Run it against the deployment's databases,
   with a dry run first:

   ```shell
   export CAS_DATABASE_URI=<the teams-cas MongoDB connection string>
   export CAS_DATABASE_NAME=<the teams-cas database name>
   export FIFTYONE_DATABASE_URI=<the FiftyOne MongoDB connection string>
   export FIFTYONE_DATABASE_NAME=<the FiftyOne database name>

   # Shows what would be undone
   fiftyone-migrator rollback 2.25.1 --include cas --dry-run

   fiftyone-migrator rollback 2.25.1 --include cas
   ```

   The command returns invitations, service accounts, and their groups
   to the form v2.25 uses.
   You can run it again safely.
   It works whether or not anyone ran `fiftyone-migrator`
   on the deployment before.

Upgrading to v2.26 again later needs no extra step.

#### FiftyOne Enterprise v2.25+ Activity Analytics

FiftyOne Enterprise v2.25.0 introduces Activity Analytics, which records
operator runs, annotation and review decisions, and sample and label
mutations for the Audit Log, the Jobs pages, and the annotation Metrics
tab.

**Activity Analytics is on by default.** It shipped opt-in in v2.25.0,
behind the `compose.activity.yaml` overlay. In later versions an upgrade
with no changes:

- starts two new services from the base compose file:
  `fiftyone-mq-redis` (the queue, persisted on the new
  `fiftyone-mq-redis-data` volume) and `activity-worker`
- sets `FIFTYONE_ACTIVITY_ENABLED=true` on `fiftyone-app`, `teams-api`,
  `teams-plugins`, and the `teams-do*` workers

FiftyOne Enterprise v2.26 removes `compose.activity.yaml`.
Remove it from your `-f` set.

Activity data accumulates from the upgrade onward; earlier activity is not
backfilled.

To stay opted out, see
[Turning Activity Analytics off](./configuring-activity-analytics.md#turning-activity-analytics-off).

#### FiftyOne Enterprise v2.23+ Service Orchestrators

FiftyOne Enterprise v2.23.0 introduces service orchestrators: delegated-operator
workers that host long-lived services (always-on model servers)
rather than tasks that exit.
Two builtin services (requiring GPU(s)) are provided:
`annotation-ai` (SAM2) and `agentic-labeler` (few-shot VLM labeling).

This upgrade adds a `builtin_services.yaml` bind mount to `teams-api` in
`common-services.yaml`.
`teams-api` reconciles the `builtin_services.yaml` service list entries at startup.
The two builtin services will appear in the
FiftyOne Enterprise UI under *Settings* -> *Services*.
Both are registered in a stopped state and neither starts on its own.

The two services target different workers by default:

- `annotation-ai`
  - Targets the default `teams-do` worker (`delegation_target: builtin`).
  - Needs GPU access before the service can start,
    or retarget the service to a worker that has one.
- `agentic-labeler`
  - Targets the dedicated GPU worker added by `compose.agenticlabeler.yaml`,
    which is not part of the default `-f` set.

> **NOTE**: In `v2.23.0`, `compose.agenticlabeler.yaml` errantly contained
> `voxel51/fiftyone-teams-agentic-labeler`.
> This image doesn't exist and an error occurred when pulled.
> In `v2.23.1`, image is fixed (set to `voxel51/agentic-labeler`).
> If you copied that file or pinned the image in an override, correct the name.

Edit
[builtin_services.yaml](../builtin_services.yaml)
to add, remove, or retarget services.
To update a builtin service's definition, you must
increment the service's `builtin_version`.
Otherwise the cached definition will not be updated.

See the
[Configuring Service Orchestrators](../../docs/configuring-service-orchestrator.md)
documentation for the builtin services list and accelerator sizing,
[Configuring the Agentic Labeler Service](./configuring-agentic-labeler.md)
for the dedicated GPU worker, and
[Leveraging GPU Workloads](./configuring-gpu-workloads.md)
for host GPU setup.

#### FiftyOne Enterprise v2.22+ Multimodal Datasets

FiftyOne Enterprise v2.22.0 introduces multimodal dataset support: large
modalities associated with each sample, stored as Parquet-backed Iceberg
tables rather than as fields on the sample document, ingested/compacted via
a background delegated-operator pipeline.

Multimodal datasets require:

- The `VFF_MULTIMODAL` feature flag, set on `teams-api`, `fiftyone-app`,
  `teams-do` workers, `teams-app`, and `teams-plugins`.
- Sufficient host disk space for `teams-do` containers to stage projection
  compaction files in `/tmp`.
- Enough memory on `fiftyone-app` to serve multimodal grid queries, which
  run DuckDB in-process; size it and `HYPERCORN_WORKERS` accordingly.
- Optionally, `FIFTYONE_PROJECTION_DELEGATION_TARGET` on `teams-api` to pin
  projection processing to a specific `teams-do` worker.

See the
[Configuring Multimodal Datasets](./configuring-multimodal.md)
documentation for full details, required services, and example
configuration.

#### FiftyOne Enterprise v2.19+ Telemetry Sidecars

FiftyOne Enterprise v2.19.0 adds observability features viewable by
admins directly in the FiftyOne UI.
These are powered by a `telemetry-sidecar` service paired with each
`fiftyone-app`, `teams-api`, `teams-plugins`, and `teams-do*` service,
plus a `telemetry-redis` service that buffers the streamed metrics and
logs.

**Resource impact:**
By default the telemetry requires an additional `0.55` CPU and `1.5Gi` of
resources used by:

- `telemetry-sidecar` container
  - `teams-api`: 0.1 CPU and 512 Mi memory
  - `fiftyone-app`: 0.2 CPU and 512 Mi memory
- Redis container
  - 0.25 CPU and 512Mi memory

The additional required resources depends of replica count for each
deployment.
The optional deployments may also increase this amount.

##### Host Requirements

1. **Docker Compose v2.17+** for `depends_on.<svc>.restart: true`
   semantics.
   Older versions will see stale PID namespaces after a target
   container is recreated.
1. **Delegated-operator workers are scaled via [Compose profiles](https://docs.docker.com/compose/how-tos/profiles/)**
   (`do-2`, `do-3`) rather than the deprecated
   `FIFTYONE_DELEGATED_OPERATOR_WORKER_REPLICAS` env var, because
   Compose's `pid: "service:<name>"` only joins a single replica.
   Slot 1 (`teams-do`) is always on, and slots 2-3 each get their own
   paired sidecar. See
   [Scaling delegated-operator workers](#scaling-delegated-operator-workers)
   below.
1. **`teams-do` requires the **`SYS_PTRACE`** capability to allow the
   telemetry agent to observe the target process.

**External Redis:**
To use an existing Redis instance instead of the bundled one by setting
`FIFTYONE_TELEMETRY_REDIS_URL` in your `.env` to a fully-qualified URL
(e.g. `redis://my-managed-redis.example.com:6379`) and scaling
`telemetry-redis` to `replicas: 0` as below.

##### Opting out of Telemetry

Telemetry is enabled by default.
To disable it, add a `compose.override.yaml` that scales the
`telemetry-redis` and `*-telemetry` services to `replicas: 0`.
See
[`configuring-telemetry.md`](configuring-telemetry.md#opting-out)
for the full override snippet.

> [!IMPORTANT]
> The sidecar powers the FiftyOne UI's delegated-operator log viewer.
> Disabling telemetry will leave that log viewer empty.

##### Scaling delegated-operator workers

> [!WARNING]
> **Breaking change.**
> `FIFTYONE_DELEGATED_OPERATOR_WORKER_REPLICAS` is deprecated
> in docker compose deployments — setting it has no effect.
> The default rendered worker count drops from **3** (pre-2.19.0) to
> **1** when you layer `compose.delegated-operators.yaml`.

`compose.delegated-operators.yaml` now declares three worker slots,
each paired with its own telemetry sidecar:

| Slot | Service       | Activation             |
| ---- | ------------- | ---------------------- |
| 1    | `teams-do`    | always on (no profile) |
| 2    | `teams-do-2`  | profile `do-2`, `do-3` |
| 3    | `teams-do-3`  | profile `do-3`         |

Set `COMPOSE_PROFILES=do-<N>` to add slots 2-3; profiles are nested so
`do-3` includes slot 2.

**To preserve the previous default of three workers**, set the
following in your `.env`:

```shell
COMPOSE_PROFILES=do-3
```

If
you previously set `FIFTYONE_DELEGATED_OPERATOR_WORKER_REPLICAS`,
remove it from your `.env` and pick the matching `do-<N>` profile.
For more than three, the slot-2/3 service blocks can be duplicated
as `teams-do-4` / etc.
(bump the service name, `pid`, `POD_NAME`, `-n`, and socket-volume
name on each copy) — see the per-slot inline comments in
`compose.delegated-operators.yaml`.

Each worker now registers under its own orchestrator name (slot 2 as
`teams-do-2`, slot 3 as `teams-do-3`) so they surface separately in
Settings → Monitoring → System Health.
Resource impact scales linearly with the worker count:
each additional slot adds ~0.2 CPU and ~1 GiB of memory
(worker reservation + paired sidecar).

See
[`docker/docs/configuring-telemetry.md`](configuring-telemetry.md#scaling-teams-do-with-telemetry)
for the full slot/profile reference.

#### FiftyOne Enterprise v2.16+ Additional API Routes

FiftyOne Enterprise v2.16.0 adds the `/cloud_credentials` endpoint to the `teams-api`.
If using path-based routing, please update your Nginx configuration to
include this endpoint:

```nginx
server {
  server_name your.server.name;

  # existing configuration

  location /cloud_credentials {
    proxy_pass http://teams-api;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
  }

  # existing configuration
}
```

Please see the
[ingress documentation](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/docs/expose-teams-api.md)
for full details.

#### FiftyOne Enterprise v2.15+ Additional API Routes

FiftyOne Enterprise v2.15.0 adds the `/rpc` endpoints to the `teams-api`.
If using path-based routing, please update your Nginx configuration to
include these endpoints:

```nginx
server {
  server_name your.server.name;

  # existing configuration

  location /rpc {
    proxy_pass http://teams-api;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
  }

  # existing configuration
}
```

Please see the
[ingress documentation](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/docs/expose-teams-api.md)
for full details.

#### FiftyOne Enterprise v2.7+ Delegated Operator Changes

FiftyOne Enterprise v2.7.0 changes the `FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH`
environment variable to `FIFTYONE_DELEGATED_OPERATION_LOG_PATH`.
Please note that this change is backwards compatible, but should
be changed in your manifests moving forward.

#### FiftyOne Enterprise v2.5+ Delegated Operator Changes

FiftyOne Enterprise v2.5.0 changes the base image of the built-in delegated
operators (`teams-do`) from `voxel51/fiftyone-app` to `voxel51/fiftyone-teams-cv-full`.
The `voxel51/fiftyone-teams-cv-full` image includes all of the dependencies
required to run complex workflows out of the box.

If you built your own image with custom dependencies,
you will likely want to remake those images based off
of this new `voxel51/fiftyone-teams-cv-full` image.

Please note: this image is approximately 2GB larger than its predecessor
and, as such, might take longer to pull and start.

To utilize the prior image,
update your `compose.override.yaml` similar to the below:

```yaml
services:
  teams-do:
    image: voxel51/fiftyone-app:v2.26.0
```

#### FiftyOne Enterprise v2.2+ Delegated Operator Changes

FiftyOne Enterprise v2.2 introduces some changes to delegated operators, detailed
below.

#### Delegated Operation Capacity

By default, all deployments are provisioned with capacity to support up to three
delegated operations simultaneously. You will need to configure the
[builtin orchestrator](../README.md#builtin-delegated-operator-orchestrator)
or an external
orchestrator, with enough workers, to be able to utilize this full capacity.
If your team finds the usage is greater than this, please reach out to your
Voxel51 support team for guidance and to increase this limit!

#### Existing Orchestrators

> [!NOTE]
> If you are currently utilizing an external orchestrator for delegated
> operations, such as Airflow or Flyte, you may have an outdated execution
> definition that could negatively affect the experience. Please reach out to
> Voxel51 support team for guidance on updating this code.

Additionally,

> [!WARNING]
> If you cannot update the orchestrator DAG/workflow code, you must set
> `FIFTYONE_ALLOW_LEGACY_ORCHESTRATORS=true` in order for the delegated
> operation system to function properly.

### From FiftyOne Enterprise Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Enterprise v2.26.0 *requires* a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
> install process.

---

1. Ensure all FiftyOne SDK users either
    - Set the `FIFTYONE_DATABASE_ADMIN` to `false`

      ```shell
      FIFTYONE_DATABASE_ADMIN=false
      ```

    - Unset the environment variable `FIFTYONE_DATABASE_ADMIN`
      (this should generally be your default)

        ```shell
        unset FIFTYONE_DATABASE_ADMIN
        ```

1. Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Enterprise docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. [Upgrade to FiftyOne Enterprise version 2.26.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.26.0
    1. Log in to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE**: Any FiftyOne SDK less than 2.26.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.26.0` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at the latest version, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Enterprise Version 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Enterprise v2.26.0 *requires*
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Enterprise Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Enterprise v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.26.0 *requires* a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
> install process.

---

1. Copy your `compose.override.yaml` and `.env` files
   into the `docker` directory
1. `cd` into the `docker` directory
1. In the `.env` file, set `BASE_URL` to your WebApp URL
1. In the `.env` file, set the required environment variables
    - `FIFTYONE_API_URI`
    - `FIFTYONE_AUTH_SECRET`
    - `CAS_BASE_URL`
    - `CAS_BIND_ADDRESS`
    - `CAS_BIND_PORT`
    - `CAS_DATABASE_NAME`
    - `CAS_DEBUG`
    - `CAS_DEFAULT_USER_ROLE`

    > **NOTE**: For the `CAS_*` variables, consider using
    > the seed values from the `.env.template` file.
    > See
    > [Central Authentication Service](https://helm.fiftyone.ai/#central-authentication-service)
1. In your `.env` file, set the `LOCAL_LICENSE_FILE_DIR` variable value. Copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Enterprise docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. Ensure all FiftyOne SDK users either
    - Set the `FIFTYONE_DATABASE_ADMIN` to `false`

      ```shell
      FIFTYONE_DATABASE_ADMIN=false
      ```

    - Unset the environment variable `FIFTYONE_DATABASE_ADMIN`
      (this should generally be your default)

        ```shell
        unset FIFTYONE_DATABASE_ADMIN
        ```

1. [Upgrade to FiftyOne Enterprise version 2.26.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.26.0
    1. Log in to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE**: Any FiftyOne SDK less than 2.26.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.26.0` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at the latest version, run

    ```shell
    fiftyone migrate --info
    ```

### From Before FiftyOne Enterprise Version 1.1.0

> **NOTE**: Upgrading from versions of FiftyOne Enterprise prior to v1.1.0
> requires upgrading the database and will interrupt all SDK connections.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Enterprise v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](../README.md#step-6-identity-provider-idp-and-authentication-cas)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.26.0 *requires* your users to
> log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Enterprise Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.26.0 *requires* a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
> install process.

---

1. Copy your `compose.override.yaml` and `.env` files
   into the `docker` directory
1. `cd` into the `docker` directory
1. In the `.env` file, set `BASE_URL` to your WebApp URL
1. In your `.env` file, set the required environment variables:
    - `FIFTYONE_ENCRYPTION_KEY`
    - `FIFTYONE_API_URI`
    - `FIFTYONE_AUTH_SECRET`
1. In your `.env` file, set the `LOCAL_LICENSE_FILE_DIR` variable value. Copy
   the license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Enterprise docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. Update your web server routes to include routing
   `/cas/*` traffic to the `teams-cas` service.
   Please see our [example nginx configurations](../) for more information.
1. [Upgrade to FiftyOne Enterprise v2.26.0](#upgrading-from-previous-versions)
   with `FIFTYONE_DATABASE_ADMIN=true`
   (this is not the default for this release).
    > **NOTE**: FiftyOne SDK users will lose access to the FiftyOne
    > Enterprise Database at this step until they upgrade to `fiftyone==2.26.0`

1. Upgrade your FiftyOne SDKs to version 2.26.0
    1. Log in to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Enterprise version, navigate to
      `Account > Install FiftyOne`
1. Confirm that datasets have been migrated to the latest version

    ```shell
    fiftyone migrate --info
    ```

   - If not all datasets have been upgraded, have an admin run

      ```shell
      FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
      ```

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this
upgrade.
You will need to either create a new Identity Provider (IdP) or modify your
existing configuration to migrate to a new Auth0 Tenant.
