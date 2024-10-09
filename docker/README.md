<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

## FiftyOne Teams v2.0+ Requires a License File

FiftyOne Teams v2.0 introduces a new requirement for a license file.  This
license file should be obtained from your Customer Success Team before
upgrading to FiftyOne Teams 2.0 or beyond.

The license file now contains all of the Auth0 configuration that was
previously provided through environment variables; you may remove those secrets
from your `.env` and from any secrets created outside of the Voxel51
install process.

Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the license
file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne Teams docker
compose host.
e.g.:

```shell
. .env
mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
```

## Known Issues for FiftyOne Teams v1.6.0 and Above

### Invitations Disabled for Internal Authentication Mode

FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS), which
includes both
[`legacy` authentication mode][legacy-auth-mode]
and
[`internal` authentication mode][internal-auth-mode].

Inviting users to join your FiftyOne Teams instance is not currently supported
when `FIFTYONE_AUTH_MODE` is set to `internal`.

## Table of Contents

<!-- toc -->

- [Deploying FiftyOne Teams App with Docker Compose](#deploying-fiftyone-teams-app-with-docker-compose)
  - [Initial Installation vs. Upgrades](#initial-installation-vs-upgrades)
  - [FiftyOne Teams Features](#fiftyone-teams-features)
    - [Central Authentication Service](#central-authentication-service)
    - [Snapshot Archival](#snapshot-archival)
    - [FiftyOne Teams Authenticated API](#fiftyone-teams-authenticated-api)
    - [FiftyOne Teams Plugins](#fiftyone-teams-plugins)
      - [Builtin Plugins Only](#builtin-plugins-only)
      - [Shared Plugins](#shared-plugins)
      - [Dedicated Plugins](#dedicated-plugins)
      - [Delegated Operators](#delegated-operators)
    - [Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](#storage-credentials-and-fiftyone_encryption_key)
    - [Proxies](#proxies)
    - [Text Similarity](#text-similarity)
  - [Upgrading From Previous Versions](#upgrading-from-previous-versions)
    - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)
    - [From Before FiftyOne Teams Version 1.1.0](#from-before-fiftyone-teams-version-110)
    - [From FiftyOne Teams Version 1.1.0 and Before Version 1.6.0](#from-fiftyone-teams-version-110-and-before-version-160)
    - [From FiftyOne Teams Versions 1.6.0 to 1.7.1](#from-fiftyone-teams-versions-160-to-171)
    - [From FiftyOne Teams Version 2.0.0](#from-fiftyone-teams-version-200)
  - [Deploying FiftyOne Teams](#deploying-fiftyone-teams)
  - [FiftyOne Teams Environment Variables](#fiftyone-teams-environment-variables)

<!-- tocstop -->

---

# Deploying FiftyOne Teams App with Docker Compose

We publish the following FiftyOne Teams private images to Docker Hub:

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch`
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`
- `voxel51/fiftyone-teams-cas`

For Docker Hub credentials, please contact your Voxel51 support team.

---

## Initial Installation vs. Upgrades

When performing an initial installation, in `compose.override.yaml` set
`services.fiftyone-app.environment.FIFTYONE_DATABASE_ADMIN: true`.
When performing a FiftyOne Teams upgrade, set
`services.fiftyone-app.environment.FIFTYONE_DATABASE_ADMIN: false`.
See
[Upgrading From Previous Versions](#upgrading-from-previous-versions)

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
FiftyOne Teams version: 0.14.4

FiftyOne compatibility version: 0.22.3
Other compatible versions: >=0.19,<0.23

Database version: 0.21.2

dataset     version
----------  ---------
quickstart  0.22.0
$ fiftyone migrate --all
$ fiftyone migrate --info
FiftyOne Teams version: 0.14.4

FiftyOne compatibility version: 0.23.0
Other compatible versions: >=0.19,<0.23

Database version: 0.21.2

dataset     version
----------  ---------
quickstart  0.21.2
```

---

## FiftyOne Teams Features

### Central Authentication Service

FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
CAS requires additional configurations and consumes additional resources.
Please review these notes, and the
[Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
documentation before completing your upgrade.

Voxel51 recommends upgrading your deployment using
[`legacy` authentication mode][legacy-auth-mode]
and migrating to
[`internal` authentication mode][internal-auth-mode]
after confirming your initial upgrade was successful.

Please contact your Voxel51 customer success
representative for assistance in migrating to internal mode.

The CAS service requires changes to your `.env` files.
A brief summary of those changes include

- Add the `FIFTYONE_AUTH_SECRET` variable used in every service
- Add the following CAS Service configuration variables
  - `CAS_BASE_URL`
  - `CAS_BIND_ADDRESS`
  - `CAS_BIND_PORT`
  - `CAS_DATABASE_NAME`
  - `CAS_DEBUG`
  - `CAS_DEFAULT_USER_ROLE`

> **NOTE**: When multiple deployments use the same database instance,
> set `CAS_DATABASE_NAME` to a unique value for each deployment.

Please review these changes in the
[legacy-auth/env.template](legacy-auth/env.template)
and in the appropriate `legacy-auth/compose*` files.

To upgrade from versions prior to FiftyOne Teams v1.6

- Copy your `.env` file into the `legacy-auth` directory
- Copy your `compose.override.yaml` file into the `legacy-auth` directory
- `cd` into the `legacy-auth` directory
- Update your `.env` file, adding the variables listed above
  - For seed values, see
    [legacy-auth/env.template](legacy-auth/env.template)
- Update your `compose.override.yaml` with `teams-cas` changes (if necessary)
- Run `docker compose` commands from the `legacy-auth` directory
- When using path-based routing, configure a `/cas` route to value of the
  `CAS_BIND_PORT`

> **NOTE**: See
> [Upgrading From Previous Versions](#upgrading-from-previous-versions)

### Snapshot Archival

Since version v1.5, FiftyOne Teams supports
[archiving snapshots](https://docs.voxel51.com/teams/dataset_versioning.html#snapshot-archival)
to cold storage locations to prevent filling up the MongoDB database.
To enable this feature, set the `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
environment variable to the path of a chosen storage location.

Supported locations are network mounted filesystems and cloud storage folders.

- Network mounted filesystem
  - Set the environment variable `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH` to the
    mounted filesystem path in these containers
    - `fiftyone-api`
    - `teams-app`
  - Mount the filesystem to the `fiftyone-api` container
    (`teams-app` does not need this despite the variable set above).
    For an example, see
    [legacy-auth/compose.plugins.yaml](legacy-auth/compose.plugins.yaml).
- Cloud storage folder
  - Set the environment variable `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH` to a
    cloud storage path (for example
    `gs://my-voxel51-bucket/dev-deployment-snapshot-archives/`)
    in these containers
    - `fiftyone-api`
    - `teams-app`
  - Ensure the
    [cloud credentials](https://docs.voxel51.com/teams/installation.html#cloud-credentials)
    loaded in the `fiftyone-api` container have full edit capabilities to
    this bucket

See the
[configuration documentation](https://docs.voxel51.com/teams/dataset_versioning.html#dataset-versioning-configuration)
for other configuration values that control the behavior of automatic snapshot
archival.

### FiftyOne Teams Authenticated API

FiftyOne Teams v1.3 introduces the capability to connect FiftyOne Teams SDK
through the FiftyOne Teams API (instead of creating a direct connection to
MongoDB).

To enable the FiftyOne Teams Authenticated API you will need to
[expose the FiftyOne Teams API endpoint](docs/expose-teams-api.md)
and
[configure your SDK](https://docs.voxel51.com/teams/api_connection.html).

### FiftyOne Teams Plugins

FiftyOne Teams v1.3+ includes significant enhancements for
[Plugins](https://docs.voxel51.com/plugins/index.html)
to customize and extend the functionality of FiftyOne Teams in your environment.

There are three modes for plugins

1. Builtin Plugins Only
    - This is the default mode
    - Users may only run the builtin plugins shipped with Fiftyone Teams
    - Cannot run custom plugins
1. Shared Plugins
    - Users may run builtin and custom plugins
    - Plugins run in the existing `fiftyone-app` service
      - Plugins resource consumption may starve `fiftyone-app`,
        causing the app to be slow or crash
    - Requires creating a volume mounted to the services
      - `fiftyone-app` (read-only)
      - `teams-api` (read-write)
1. Dedicated Plugins
    - Users may run builtin and custom plugins
    - Plugins run in an additional `teams-plugins` service
    - Plugins run in a dedicated `teams-plugins` service
      - Plugins resource consumption does not affect `fiftyone-app`
    - Requires creating a volume mounted to the services
      - `teams-plugins` (read-only)
      - `teams-api` (read-write)

For multi-node deployments, please implement a storage
solution that provides access to the deployed plugins.

To use plugins with custom dependencies, build and use
[Custom Plugins Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).

To use the FiftyOne Teams UI to deploy plugins, navigate to
`https://<DEPLOY_URL>/settings/plugins`. Early-adopter plugins installed
manually must be redeployed using the FiftyOne Teams UI.

#### Builtin Plugins Only

Enabled by default. No additional configurations are required.

#### Shared Plugins

Plugins run in the `fiftyone-app` service.
To enable this mode, use the file
[legacy-auth/compose.plugins.yaml](legacy-auth/compose.plugins.yaml)
instead of
[legacy-auth/compose.yaml](legacy-auth/compose.yaml).
This compose file creates a new Docker Volume shared between FiftyOne Teams
services.

- Configure the services to access to the plugin volume
  - `fiftyone-app` requires `read`
  - `fiftyone-api` requires `read-write`
- Example `docker compose` command for this mode from the `legacy-auth`
directory

    ```shell
    docker compose \
      -f compose.plugins.yaml \
      -f compose.override.yaml \
      up -d
    ```

#### Dedicated Plugins

Plugins run in the `teams-plugins` service. To enable this mode, use the file
[legacy-auth/compose.dedicated-plugins.yaml](legacy-auth/compose.dedicated-plugins.yaml)
instead of
[legacy-auth/compose.yaml](legacy-auth/compose.yaml).
This compose file creates a new Docker Volume shared between FiftyOne Teams
services.

- Configure the services to access to the plugin volume
  - `teams-plugins` requires `read`
  - `fiftyone-api` requires `read-write`
- If you are
  [using a proxy](#proxies),
  add the `teams-plugins` service name to your environment variables
  - `no_proxy`
  - `NO_PROXY`
- Example `docker compose` command for this mode from the `legacy-auth`
  directory

    ```shell
    docker compose \
      -f compose.dedicated-plugins.yaml \
      -f compose.override.yaml \
      up --d
    ```

#### Delegated Operators

If you would like to execute
[delegated operations](https://docs.voxel51.com/teams/teams_plugins.html?highlight=delegated#teams-delegated-operations)
without the need to setup your own orchestrator, such as Airflow, you can launch worker
containers using [legacy-auth/compose.delegated-operators.yaml](legacy-auth/compose.delegated-operators.yaml)
in conjunction with any of the plugin configurations above.

- Example `docker compose` command for this mode from the `legacy-auth`
  directory

    ```shell
    docker compose \
      -f compose.dedicated-plugins.yaml \
      -f compose.delegated-operators.yaml \
      -f compose.override.yaml \
      up --d
    ```

### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

As of FiftyOne Teams 1.1, containers based on the `fiftyone-teams-api` and
`fiftyone-app` images must include the `FIFTYONE_ENCRYPTION_KEY` variable.
This key is used to encrypt storage credentials in the MongoDB database.

To generate a value for `FIFTYONE_ENCRYPTION_KEY`, run this
Python code and add the output to your `.env` file:

```python
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
```

Voxel51 does not have access to this encryption key and cannot reproduce it.
Please store this key in a safe place.
If the key is lost, you will need to

1. Schedule an outage window
    1. Drop the storage credentials collection
    1. Replace the encryption key
    1. Add the storage credentials via the UI again.

Storage credentials no longer need to be mounted into
containers with appropriate environment variables being set.
Users with `Admin` permissions may use the FiftyOne Teams UI
to manage storage credentials by navigating to
`https://<DEPOY_URL>/settings/cloud_storage_credentials`.

FiftyOne Teams version 1.3+ continues to support the use of environment
variables to set storage credentials in the application context and is
providing an alternate configuration path.

### Proxies

FiftyOne Teams supports routing traffic through proxy servers.
To configure this, set following environment variables in your
`compose.override.yaml`

1. All services

    ```yaml
    http_proxy: ${HTTP_PROXY_URL}
    https_proxy: ${HTTPS_PROXY_URL}
    no_proxy: ${NO_PROXY_LIST}
    HTTP_PROXY: ${HTTP_PROXY_URL}
    HTTPS_PROXY: ${HTTPS_PROXY_URL}
    NO_PROXY: ${NO_PROXY_LIST}
    ```

1. All services based on the `fiftyone-teams-app` and `fiftyone-teams-cas`
   images

    ```yaml
    GLOBAL_AGENT_HTTP_PROXY: ${HTTP_PROXY_URL}
    GLOBAL_AGENT_HTTPS_PROXY: ${HTTPS_PROXY_URL}
    GLOBAL_AGENT_NO_PROXY: ${NO_PROXY_LIST}
    ```

The environment variable `NO_PROXY_LIST` value should be a comma-separated list
of Docker Compose services that may communicate without going through a proxy
server. By default these service names are:

- `fiftyone-app`
- `teams-api`
- `teams-app`
- `teams-cas`
- `teams-plugins`

Examples of these settings are included in the FiftyOne Teams configuration files

- [common-services.yaml](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/common-services.yaml)
- [legacy-auth/env.template](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/legacy-auth/env.template)

By default, the Global Agent Proxy will log all outbound connections and
identify which connections are routed through the proxy.
To reduce the logging verbosity, add this environment variable to your
`teams-app` and `teams-cas` services.

```yaml
ROARR_LOG: false
```

### Text Similarity

FiftyOne Teams version 1.2 and higher supports using text similarity searches
for images that are indexed with a model that
[supports text queries](https://docs.voxel51.com/user_guide/brain.html#brain-similarity-text).
To use this feature, use a container image containing `torch` (PyTorch) instead
of the `fiftyone-app` image.
Use the Voxel51 provided image `fiftyone-app-torch` or build your own base
image including `torch`.

To override the default image, update `compose.override.yaml` with the value
for your image.
This will allow you to update your `compose.yaml` in future releases without
having to port this change forward. For example, `compose.override.yaml`
might look like:

```yaml
services:
  fiftyone-app:
    image: voxel51/fiftyone-app-torch:v2.2.0
```

For more information, see the docs for
[Docker Compose Extend](https://docs.docker.com/compose/extends/).

## Upgrading From Previous Versions

Voxel51 assumes you use the published Docker compose files to deploy your
FiftyOne Teams environment.
If you use custom deployment mechanisms, carefully review the changes in the
[Docker Compose Files](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/main/docker)
and update your deployment accordingly.

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this
upgrade.
You will need to either create a new Identity Provider (IdP) or modify your
existing configuration to migrate to a new Auth0 Tenant.

### From Before FiftyOne Teams Version 1.1.0

> **NOTE**: Upgrading from versions of FiftyOne Teams prior to v1.1.0
> requires upgrading the database and will interrupt all SDK connections.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_ your users to log in
> after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
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
1. Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Teams docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. Ensure your web server routes are updated to include routing
   `/cas/*` traffic to the `teams-cas` service.
   Example nginx configurations can be found
   [here](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/main/docker)
1. [Upgrade to FiftyOne Teams v2.2.0](#deploying-fiftyone-teams)
   with `FIFTYONE_DATABASE_ADMIN=true`
   (this is not the default for this release).
    > **NOTE**: FiftyOne SDK users will lose access to the FiftyOne
    > Teams Database at this step until they upgrade to `fiftyone==2.2.0`

1. Upgrade your FiftyOne SDKs to version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Teams version, navigate to
      `Account > Install FiftyOne`
1. Confirm that datasets have been migrated to version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

   - If not all datasets have been upgraded, have an admin run

      ```shell
      FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
      ```

### From FiftyOne Teams Version 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service
> (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through  environment variables; you may remove those secrets
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
    > [Central Authentication Service](#central-authentication-service)
1. Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the
   license file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne
   Teams docker compose host.

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

1. [Upgrade to FiftyOne Teams version 2.2.0](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE** Any FiftyOne SDK less than 2.2.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.2.0` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 0.25.1, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
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
   Teams docker compose host.

   ```shell
   . .env
   mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
   mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
   ```

1. [Upgrade to FiftyOne Teams version 2.2.0](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets
    > **NOTE** Any FiftyOne SDK less than 2.2.0
    > will lose connectivity at this point.
    > Upgrading to `fiftyone==2.2.0` is required.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 0.25.1, run

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Version 2.0.0

1. [Upgrade to FiftyOne Teams version 2.2.0](#deploying-fiftyone-teams)
1. Voxel51 recommends upgrading all FiftyOne Teams SDK users to FiftyOne Teams
   version 2.2.0, but it is not required
   - Login to the FiftyOne Teams UI
   - To obtain the CLI command to install the FiftyOne SDK associated with
     your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Voxel51 recommends that you upgrade all your datasets, but it is not
   required.  Users using the FiftyOne Teams 2.0.0 SDK will continue to operate
   uninterrupted during, and after, this migration

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at version 0.25.1, run

   ```shell
   fiftyone migrate --info
   ```

## Deploying FiftyOne Teams

1. Install
   [Docker Compose](https://docs.docker.com/compose/install/)
1. Choose to install using `legacy-auth` (recommended) or `internal-auth` by
   `cd`ing into either the `legacy-auth` or `internal-auth` subdirectory in this
   repository.
1. In the directory chosen above
    1. Rename the `env.template` file to `.env`
    1. Edit the `.env` file, setting all the customer provided required
    settings.
       See the
       [FiftyOne Teams Environment Variables](#fiftyone-teams-environment-variables)
       table.
    1. Create a `compose.override.yaml` with any configuration overrides for
       this deployment
        1. For the first installation, set

            ```yaml
            services:
              fiftyone-app-common:
                environment:
                  FIFTYONE_DATABASE_ADMIN: true
            ```

1. Make sure you have put your Voxel51-provided FiftyOne Teams license in the
   local directory identified by the `LOCAL_LICENSE_FILE_DIR` configured in
   your `.env` file.
1. Deploy FiftyOne Teams
    1. In the same directory, run

        ```shell
        docker-compose up -d
        ```

1. After the successful installation, and logging into Fiftyone Teams
    1. In `compose.override.yaml`, remove the `FIFTYONE_DATABASE_ADMIN` override

        ```yaml
        services:
          fiftyone-app-common:
            environment:
              # FIFTYONE_DATABASE_ADMIN: true
        ```

        > **NOTE**: This example shows commenting this line,
        > however you may remove the line.

        or set it to `false` like in

        ```yaml
        services:
          fiftyone-app-common:
            environment:
              FIFTYONE_DATABASE_ADMIN: false
        ```

The FiftyOne Teams API is exposed on port `8000`.
The FiftyOne Teams App is exposed on port `3000`.
The FiftyOne Teams CAS is exposed on port `3030`.

Configure an SSL endpoint (like a Load Balancer, Nginx Proxy, or similar)
to route traffic to the appropriate endpoints. An example Nginx configuration
for path-based routing can be found
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/example-nginx-path-routing.conf).
Example Nginx configurations for hostname-based routing can be found
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/example-nginx-site.conf)
for FiftyOne Teams App and FiftyOne Teams CAS services, and
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/example-nginx-api.conf)
for the FiftyOne Teams API service.

---

## FiftyOne Teams Environment Variables

| Variable                                     | Purpose                                                                                                                                                                                                                                                                        | Required                  |
|----------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------|
| `API_BIND_ADDRESS`                           | The host address that `fiftyone-teams-api` should bind to; `127.0.0.1` is appropriate for this in most cases                                                                                                                                                                   | Yes                       |
| `API_BIND_PORT`                              | The host port that `fiftyone-teams-api` should bind to; the default is `8000`                                                                                                                                                                                                  | Yes                       |
| `API_LOGGING_LEVEL`                          | Logging Level for `teams-api` service                                                                                                                                                                                                                                          | Yes                       |
| `API_URL`                                    | The URL that `fiftyone-teams-app` should use to communicate with `fiftyone-teams-api`; `teams-api` is the compose service name                                                                                                                                                 | Yes                       |
| `APP_BIND_ADDRESS`                           | The host address that `fiftyone-teams-app` should bind to; `127.0.0.1` is appropriate in most cases                                                                                                                                                                            | Yes                       |
| `APP_BIND_PORT`                              | The host port that `fiftyone-teams-app` should bind to the default is `3000`                                                                                                                                                                                                   | Yes                       |
| `APP_USE_HTTPS`                              | Set this to true if your Application endpoint uses TLS; this should be `true` in most cases'                                                                                                                                                                                   | Yes                       |
| `BASE_URL`                                   | The URL where you plan to deploy your FiftyOne Teams                                                                                                                                                                                                                           | `internal` auth mode only |
| `CAS_BASE_URL`                               | The URL that FiftyOne Teams Services should use to communicate with `teams-cas`; `teams-cas` is the compose service                                                                                                                                                            | Yes                       |
| `CAS_BIND_ADDRESS`                           | The host address that `teams-cas` should bind to; `127.0.0.1` is appropriate in most cases                                                                                                                                                                                     | Yes                       |
| `CAS_BIND_PORT`                              | The host port that `teams-cas` should bind to; the default is `3030`                                                                                                                                                                                                           | Yes                       |
| `CAS_DATABASE_NAME`                          | The MongoDB Database that the `teams-cas` service should use; the default is `fiftyone-cas`                                                                                                                                                                                    | Yes                       |
| `CAS_DEBUG`                                  | The logs that `teams-cas` should provide to stdout; see [debug](https://www.npmjs.com/package/debug) for documentation                                                                                                                                                         | Yes                       |
| `CAS_DEFAULT_USER_ROLE`                      | The default role when users initially log into the FiftyOne Teams application; the default is `GUEST`                                                                                                                                                                          | Yes                       |
| `CAS_MONGODB_URI`                            | The MongoDB Connection STring for CAS; this will default to `FIFTYONE_DATABASE_URI`                                                                                                                                                                                            | No                        |
| `FIFTYONE_APP_ALLOW_MEDIA_EXPORT`            | Set this to `"false"` if you want to disable media export options                                                                                                                                                                                                              | No                        |
| `FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED`            | Controls whether anonymous analytics are captured for the teams application. Set to false to opt-out of anonymous analytics.                                                                                                                                                                                                              | No                        |
| `FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION` | The recommended fiftyone SDK version. This will be displayed in install modal (i.e. `pip install ... fiftyone==0.11.0`)                                                                                                                                                        | No                        |
| `FIFTYONE_APP_THEME`                         | The default theme configuration for your FiftyOne Teams application as described [here](https://docs.voxel51.com/user_guide/config.html#configuring-the-app)                                                                                                                   | No                        |
| `FIFTYONE_API_URI`                           | The URI to be displayed in the `Install FiftyOne` Modal and `API Keys` configuration screens                                                                                                                                                                                   | No                        |
| `FIFTYONE_AUTH_SECRET`                       | The secret used for services to authenticate with `teams-cas`; also used to login to the SuperAdmin UI                                                                                                                                                                         | Yes                       |
| `FIFTYONE_DATABASE_NAME`                     | The MongoDB Database that `fiftyone-app`, `teams-api`, and `teams-app` use for FiftyOne Teams dataset metadata; the default is `fiftyone`                                                                                                                                      | Yes                       |
| `FIFTYONE_DATABASE_URI`                      | The MongoDB Connection String for FiftyOne Teams dataset metadata                                                                                                                                                                                                              | Yes                       |
| `FIFTYONE_DEFAULT_APP_ADDRESS`               | The host address that `fiftyone-app` should bind to; `127.0.0.1` is appropriate in most cases                                                                                                                                                                                  | Yes                       |
| `FIFTYONE_DEFAULT_APP_PORT`                  | The host port that `fiftyone-app` should bind to; the default is `5151`                                                                                                                                                                                                        | Yes                       |
| `FIFTYONE_ENCRYPTION_KEY`                    | Used to encrypt storage credentials in MongoDB                                                                                                                                                                                                                                 | Yes                       |
| `FIFTYONE_ENV`                               | GraphQL verbosity for the `fiftyone-teams-api` service; `production` will not log every GraphQL query, any other value will                                                                                                                                                    | No                        |
| `FIFTYONE_PLUGINS_DIR`                       | Persistent directory inside the containers for plugins to be mounted to. `teams-api` must have write access to this directory, all plugin nodes must have read access to this directory.                                                                                       | No                        |
| `FIFTYONE_SIGNED_URL_EXPIRATION`             | Set the time-to-live for signed URLs generated by the application in hours. Set in the `fiftyone-app` service.                                                                                                                                                                 | 24                        |
| `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`            | Full path to network-mounted file system or a cloud storage path to use for snapshot archive storage. The default `None` means archival is disabled.                                                                                                                           | No                        |
| `FIFTYONE_SNAPSHOTS_MAX_IN_DB`               | The max total number of Snapshots allowed at once. -1 for no limit. If this limit is exceeded then automatic archival is triggered if enabled, otherwise an error is raised.                                                                                                   | No                        |
| `FIFTYONE_SNAPSHOTS_MAX_PER_DATASET`         | The max number of Snapshots allowed per dataset. -1 for no limit. If this limit is exceeded then automatic archival is triggered if enabled, otherwise an error is raised.                                                                                                     | No                        |
| `FIFTYONE_SNAPSHOTS_MIN_LAST_LOADED_SEC`     | The minimum last-loaded age in seconds (as defined by `now-last_loaded_at`) a snapshot must meet to be considered for automatic archival. This limit is intended to help curtail automatic archival of a snapshot a user is actively working with. The default value is 1 day. | No                        |
| `FIFTYONE_TEAMS_PROXY_URL`                   | The URL that `fiftyone-teams-app` will use to proxy requests to `fiftyone-app`                                                                                                                                                                                                 | Yes                       |
| `GRAPHQL_DEFAULT_LIMIT`                      | Default GraphQL limit for results                                                                                                                                                                                                                                              | No                        |
| `HTTP_PROXY_URL`                             | The URL for your environment http proxy                                                                                                                                                                                                                                        | No                        |
| `HTTPS_PROXY_URL`                            | The URL for your environment https proxy                                                                                                                                                                                                                                       | No                        |
| `LOCAL_LICENSE_FILE_DIR`                    | Location of the directory that contains the FiftyOne Teams license file on the local server.                                                                                                                                                                                                               | Yes                          |
| `NO_PROXY_LIST`                              | The list of servers that should bypass the proxy; if a proxy is in use this must include the list of FiftyOne services (`fiftyone-app, teams-api,teams-app,teams-cas` must be included, `teams-plugins` should be included for dedicated plugins configurations)               | No                        |

<!-- Reference Links -->
[internal-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#internal-mode
[legacy-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#legacy-mode
