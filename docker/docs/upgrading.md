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

1. [Upgrade to FiftyOne Enterprise version 2.23.0](#upgrading-from-previous-versions)
1. Voxel51 recommends upgrading all FiftyOne Enterprise SDK users to FiftyOne Enterprise
   version 2.23.0
    1. Login to the FiftyOne Enterprise UI
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

#### FiftyOne Enterprise v2.22+ Multimodal Datasets

FiftyOne Enterprise v2.23.0 introduces multimodal dataset support: large
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
Settings → Metrics. Resource impact scales linearly with the worker
count: each additional slot adds ~0.2 CPU and ~1 GiB of memory
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

To utilize the prior image, update your `common-services.yaml` similar to the below:

```yaml
teams-do-common:
  image: voxel51/fiftyone-app:v2.23.0
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

> **NOTE**: Upgrading to FiftyOne Enterprise v2.23.0 _requires_ a license file.
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

1. [Upgrade to FiftyOne Enterprise version 2.23.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.23.0
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE**: Any FiftyOne SDK less than 2.23.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.23.0` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at the latest version, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Enterprise Version 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Enterprise v2.23.0 _requires_
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

> **NOTE**: Upgrading to FiftyOne Enterprise v2.23.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
> install process.

---

1. Copy your `compose.override.yaml` and `.env` files into the `legacy-auth`
   directory
1. `cd` into the `legacy-auth` directory
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

1. [Upgrade to FiftyOne Enterprise version 2.23.0](#upgrading-from-previous-versions)
1. Upgrade FiftyOne Enterprise SDK users to FiftyOne Enterprise version 2.23.0
    1. Login to the FiftyOne Enterprise UI
    1. To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Enterprise version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE**: Any FiftyOne SDK less than 2.23.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.23.0` is required.

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
> [Central Authentication Service](../README.md#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.23.0 _requires_ your users to
> log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Enterprise Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Enterprise v2.23.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Enterprise
> 2.0 or beyond.
>
> The license file contains all of the Auth0 configuration that was
> previously provided through environment variables. You may remove those secrets
> from your `.env` and from any secrets created outside of the Voxel51
> install process.

---

1. Copy your `compose.override.yaml` and `.env` files into the `legacy-auth`
   directory
1. `cd` into the `legacy-auth` directory
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
1. [Upgrade to FiftyOne Enterprise v2.23.0](#upgrading-from-previous-versions)
   with `FIFTYONE_DATABASE_ADMIN=true`
   (this is not the default for this release).
    > **NOTE**: FiftyOne SDK users will lose access to the FiftyOne
    > Enterprise Database at this step until they upgrade to `fiftyone==2.23.0`

1. Upgrade your FiftyOne SDKs to version 2.23.0
    1. Login to the FiftyOne Enterprise UI
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
