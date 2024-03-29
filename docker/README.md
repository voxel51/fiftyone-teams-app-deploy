<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

<!-- toc -->

- [Deploying FiftyOne Teams App with Docker Compose](#deploying-fiftyone-teams-app-with-docker-compose)
  - [Initial Installation vs. Upgrades](#initial-installation-vs-upgrades)
    - [FiftyOne Teams Upgrade Notes](#fiftyone-teams-upgrade-notes)
      - [Enabling Snapshot Archival](#enabling-snapshot-archival)
      - [Enabling FiftyOne Teams Authenticated API](#enabling-fiftyone-teams-authenticated-api)
      - [Enabling FiftyOne Teams Plugins](#enabling-fiftyone-teams-plugins)
      - [Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](#storage-credentials-and-fiftyone_encryption_key)
      - [Environment Proxies](#environment-proxies)
      - [Text Similarity](#text-similarity)
  - [Upgrade Process Recommendations](#upgrade-process-recommendations)
    - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)
    - [From Before FiftyOne Teams Version 1.1.0](#from-before-fiftyone-teams-version-110)
    - [From FiftyOne Teams Version 1.1.0 and later](#from-fiftyone-teams-version-110-and-later)
  - [Deploying FiftyOne Teams](#deploying-fiftyone-teams)
  - [FiftyOne Teams Environment Variables](#fiftyone-teams-environment-variables)

<!-- tocstop -->

---

# Deploying FiftyOne Teams App with Docker Compose

We publish container images to these Docker Hub repositories

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch`
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`

For Docker Hub credentials, please contact your Voxel51 support team.

---

## Initial Installation vs. Upgrades

When performing an initial installation, in `compose.yaml` set
`services.fiftyone-app.environment.FIFTYONE_DATABASE_ADMIN: true`.
When performing a FiftyOne Teams upgrade, set
`services.fiftyone-app.environment.FIFTYONE_DATABASE_ADMIN: false`.
See [Upgrade Process Recommendations](#upgrade-process-recommendations).

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

### FiftyOne Teams Upgrade Notes

#### Enabling Snapshot Archival

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
    [./compose.plugins.yaml](./compose.plugins.yaml).
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

See the [configuration documentation](https://docs.voxel51.com/teams/dataset_versioning.html#dataset-versioning-configuration)
for other configuration values that control the behavior of automatic snapshot archival.

#### Enabling FiftyOne Teams Authenticated API

FiftyOne Teams v1.3 introduces the capability to connect FiftyOne Teams SDK
through the FiftyOne Teams API (instead of creating a direct connection to MongoDB).

To enable the FiftyOne Teams Authenticated API you will need to
[expose the FiftyOne Teams API endpoint](docs/expose-teams-api.md)
and
[configure your SDK](https://docs.voxel51.com/teams/api_connection.html).

#### Enabling FiftyOne Teams Plugins

FiftyOne Teams v1.3+ includes significant enhancements for
[Plugins](https://docs.voxel51.com/plugins/index.html)
to customize and extend the functionality of FiftyOne Teams in your environment.

There are three modes for plugins

1. Builtin Plugins Only
    - No changes are required for this mode
1. Plugins run in the `fiftyone-app` deployment
    - To enable this mode, use the file
      [./compose.plugins.yaml](./compose.plugins.yaml)
      instead of
      [./compose.yaml](./compose.yaml)
    - Containers need the following access to plugin storage
      - `fiftyone-app` requires `read`
      - `fiftyone-api` requires `read-write`
    - Example `docker compose` command for this mode

        ```shell
        docker compose \
          -f compose.plugins.yaml \
          -f compose.override.yaml \
          up -d
        ```

1. Plugins run in a dedicated `teams-plugins` deployment
    - To enable this mode, use the file
      [./compose.dedicated-plugins.yaml](./compose.dedicated-plugins.yaml)
      instead of the
      [./compose.yaml](./compose.yaml)
    - Containers need the following access to plugin storage
      - `teams-plugins` requires `read`
      - `fiftyone-api` requires `read-write`
    - If you are [using a proxy](#environment-proxies), add the
      `teams-plugins` service name to your `no_proxy` and
      `NO_PROXY` environment variables.
    - Example `docker compose` command for this mode

        ```shell
        docker compose \
          -f compose.dedicated-plugins.yaml \
          -f compose.override.yaml \
          up -d
        ```

Both
[./compose.plugins.yaml](./compose.plugins.yaml)
and
[./compose.dedicated-plugins.yaml](./compose.dedicated-plugins.yaml)
create a new Docker Volume shared between FiftyOne Teams services.
For multi-node deployments, please implement a storage
solution allowing the access the deployed plugins.

Use the FiftyOne Teams UI to deploy plugins by navigating to `https://<DEPOY_URL>/settings/plugins`.
Early-adopter plugins installed manually must
be redeployed using the FiftyOne Teams UI.

#### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

As of FiftyOne Teams 1.1, containers based on the `fiftyone-teams-api` and
`fiftyone-app` images must include the `FIFTYONE_ENCRYPTION_KEY` variable.
This key is used to encrypt storage credentials in the MongoDB database.

To  generate `FIFTYONE_ENCRYPTION_KEY`, run this Python code

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
providing an alternate configuration path for future functionality.

#### Environment Proxies

FiftyOne Teams supports routing traffic through proxy servers.
To configure this, set following environment variables on

1. All containers

    ```yaml
    http_proxy: ${HTTP_PROXY_URL}
    https_proxy: ${HTTPS_PROXY_URL}
    no_proxy: ${NO_PROXY_LIST}
    HTTP_PROXY: ${HTTP_PROXY_URL}
    HTTPS_PROXY: ${HTTPS_PROXY_URL}
    NO_PROXY: ${NO_PROXY_LIST}
    ```

1. All containers based on the `fiftyone-teams-app` image

    ```yaml
    GLOBAL_AGENT_HTTP_PROXY: ${HTTP_PROXY_URL}
    GLOBAL_AGENT_HTTPS_PROXY: ${HTTPS_PROXY_URL}
    GLOBAL_AGENT_NO_PROXY: ${NO_PROXY_LIST}
    ```

The environment variable `NO_PROXY_LIST` value should be a comma-separated list
of Docker Compose services that may communicate without going through a proxy server.
By default these service names are

- `fiftyone-app`
- `teams-api`
- `teams-app`
- `teams-plugins`

Examples of these settings are included in the FiftyOne Teams configuration files

- [common-services.yaml](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/common-services.yaml)
- [env.template](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/env.template)

By default, the Global Agent Proxy will log all outbound connections
and identify which connections are routed through the proxy.
To reduce the logging verbosity, add this environment variable to your `teamsAppSettings.env`

```ini
ROARR_LOG: false
```

#### Text Similarity

FiftyOne Teams version 1.2 and higher supports using text
similarity searches for images that are indexed with a model that
[supports text queries](https://docs.voxel51.com/user_guide/brain.html#brain-similarity-text).
To use this feature, use a container image containing
`torch` (PyTorch) instead of the `fiftyone-app` image.
Use the Voxel51 provided image `fiftyone-app-torch` or
build your own base image including `torch`.

To override the default image, update
`compose.override.yaml` with the value for image.
This will allow you to update your `compose.yaml` in future
releases without having to port this change forward.
For example, `compose.override.yaml` might look like:

```yaml
services:
  fiftyone-app:
    image: voxel51/fiftyone-app-torch:v1.5.8
```

For more information, see the docs for
[Docker Compose Extend](https://docs.docker.com/compose/extends/).

## Upgrade Process Recommendations

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this upgrade.
To migrate to a new Auth0 Tenant, you will need to
create a new IdP or modify your existing configuration.

### From Before FiftyOne Teams Version 1.1.0

The FiftyOne 0.15.8 SDK (database version 0.23.7) is _NOT_ backwards-compatible
with FiftyOne Teams Database Versions prior to 0.19.0.
The FiftyOne 0.10.x SDK is not forwards compatible
with current FiftyOne Teams Database Versions.
If you are using a FiftyOne SDK older than 0.11.0, upgrading the
Web server will require upgrading all FiftyOne SDK installations.

Voxel51 recommends this upgrade process from
versions prior to FiftyOne Teams version 1.1.0:

1. Make sure your installation includes the required
   [FIFTYONE_ENCRYPTION_KEY](#fiftyone-teams-upgrade-notes)
   environment variable
1. Make sure you include the required `FIFTYONE_API_URI` environment variable
   (see
   [env.template](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/env.template#L17)
   for details)
1. [Upgrade to FiftyOne Teams version 1.5.8](#deploying-fiftyone-teams)
   with `FIFTYONE_DATABASE_ADMIN=true`
   (this is not the default in the `compose.yaml` for this release).
    - **NOTE:** FiftyOne SDK users will lose access to the
      FiftyOne Teams Database at this step until they upgrade to `fiftyone==0.15.8`
1. Upgrade your FiftyOne SDKs to version 0.15.8
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Check if datasets have been migrated to version 0.23.7.

    ```shell
    fiftyone migrate --info
    ```

   - If not all datasets have been upgraded, have an admin run

      ```shell
      FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
      ```

### From FiftyOne Teams Version 1.1.0 and later

The FiftyOne 0.15.8 SDK is backwards-compatible with
FiftyOne Teams Database Versions 0.19.0 and later.
You will not be able to connect to a FiftyOne Teams 1.5.8
database (version 0.23.7) with any FiftyOne SDK before 0.15.8.

Voxel51 always recommends using the latest version of the
FiftyOne SDK compatible with your FiftyOne Teams deployment.

Voxel51 recommends the following upgrade process for
upgrading from FiftyOne Teams version 1.1.0 or later:

1. Make sure you include the required `FIFTYONE_API_URI` environment variable
   (see
   [env.template](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/env.template#L17)
   for details)
1. Ensure all FiftyOne SDK users either
    - set `FIFTYONE_DATABASE_ADMIN=false`
    - `unset FIFTYONE_DATABASE_ADMIN`
        - This should generally be your default
1. [Upgrade to FiftyOne Teams version 1.5.8](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 0.15.8
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Have the admin run this to upgrade all datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

    - **NOTE** Any FiftyOne SDK less than 0.15.8 will lose database connectivity
      at this point. Upgrading to `fiftyone==0.15.8` is required

1. To ensure that all datasets are now at version 0.23.7, run

    ```shell
    fiftyone migrate --info
    ```

---

## Deploying FiftyOne Teams

1. Install docker-compose
1. From a directory containing the files `compose.yaml` and `env.template`
   files (included in this repository),
    1. Rename the `env.template` file to `.env`
    1. Edit the `.env` file, setting the parameters required for this deployment.
       [See table below](#fiftyone-teams-environment-variables).
1. In the same directory, run

    ```shell
    docker-compose up -d
    ```

1. Have the admin run to upgrade all datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 0.23.7, run

    ```shell
    fiftyone migrate --info
    ```

The FiftyOne Teams App is now exposed on port 3000.
An SSL endpoint (Load Balancer or Nginx Proxy or something similar)
will need to be configured to route traffic from the SSL endpoint
to port 3000 on the host running the FiftyOne Teams App.

An example nginx site configuration that forwards http traffic to
https, and https traffic for `your.server.name` to port 3000.
See
[./example-nginx-site.conf](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/example-nginx-site.conf).

---

## FiftyOne Teams Environment Variables

| Variable                                     | Purpose                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Required |
|----------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| `API_BIND_ADDRESS`                           | The host address that `fiftyone-teams-api` should bind to; `127.0.0.1` is appropriate for this in most cases                                                                                                                                                                                                                                                                                                                                                                                                                                    | Yes      |
| `API_BIND_PORT`                              | The host port that `fiftyone-teams-api` should bind to; the default is `8000`                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | Yes      |
| `API_URL`                                    | The URL that `fiftyone-teams-app` should use to communicate with `fiftyone-teams-api`; `teams-api` is the compose service name                                                                                                                                                                                                                                                                                                                                                                                                                  | Yes      |
| `APP_BIND_ADDRESS`                           | The host address that `fiftyone-teams-app` should bind to; this should be an externally-facing IP in most cases                                                                                                                                                                                                                                                                                                                                                                                                                                 | Yes      |
| `APP_BIND_PORT`                              | The host port that `fiftyone-teams-app` should bind to the default is `3000`                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Yes      |
| `APP_USE_HTTPS`                              | Set this to true if your Application endpoint uses TLS; this should be 'true` in most cases'                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Yes      |
| `AUTH0_API_CLIENT_ID`                        | The Auth0 API Client ID from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | Yes      |
| `AUTH0_API_CLIENT_SECRET`                    | The Auth0 API Client Secret from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Yes      |
| `AUTH0_AUDIENCE`                             | The Auth0 Audience from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Yes      |
| `AUTH0_BASE_URL`                             | The URL where you plan to deploy your FiftyOne Teams application                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | Yes      |
| `AUTH0_CLIENT_ID`                            | The Auth0 Application Client ID from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Yes      |
| `AUTH0_CLIENT_SECRET`                        | The Auth0 Application Client Secret from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | Yes      |
| `AUTH0_DOMAIN`                               | The Auth0 Domain from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | Yes      |
| `AUTH0_ISSUER_BASE_URL`                      | The Auth0 Issuer URL from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | Yes      |
| `AUTH0_ORGANIZATION`                         | The Auth0 Organization from Voxel51                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Yes      |
| `AUTH0_SECRET`                               | A random string used to encrypt cookies; use something like `openssl rand -hex 32` to generate this string                                                                                                                                                                                                                                                                                                                                                                                                                                      | Yes      |
| `FIFTYONE_APP_ALLOW_MEDIA_EXPORT`            | Set this to `"false"` if you want to disable media export options                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | No       |
| `FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION` | The recommended fiftyone SDK version. This will be displayed in install modal (i.e. `pip install ... fiftyone==0.11.0`)                                                                                                                                                                                                                                                                                                                                                                                                                         | No       |
| `FIFTYONE_APP_THEME`                         | The default theme configuration for your FiftyOne Teams application:<br>&ensp;- `dark`: Application will default to dark theme when user visits for the first time<br>&ensp;- `light`: Application will default to light theme when user visits for the first time<br>&ensp;- `always-dark`: Application will default to dark theme on each refresh (even if user changes theme to light within the app)<br>&ensp;- `always-light`: Application will default to light theme on each refresh (even if user changes theme to dark within the app) | No       | <!-- markdownlint-disable-line no-inline-html -->
| `FIFTYONE_BASE_DIR`                          | This will be mounted as `/fiftyone` in the `fiftyone-teams-app` container and can be used to pass cloud storage credentials into the environment                                                                                                                                                                                                                                                                                                                                                                                                | No       |
| `FIFTYONE_DEFAULT_APP_ADDRESS`               | The host address that `fiftyone-app` should bind to; `127.0.0.1` is appropriate for this in most cases                                                                                                                                                                                                                                                                                                                                                                                                                                          | Yes      |
| `FIFTYONE_DEFAULT_APP_PORT`                  | The host port that `fiftyone-app` should bind to; the default is `5151`                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Yes      |
| `FIFTYONE_ENCRYPTION_KEY`                    | Used to encrypt storage credentials in MongoDB                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Yes      |
| `FIFTYONE_ENV`                               | GraphQL verbosity for the `fiftyone-teams-api` service; `production` will not log every GraphQL query, any other value will                                                                                                                                                                                                                                                                                                                                                                                                                     | No       |
| `FIFTYONE_PLUGINS_DIR`                       | Persistent directory for plugins to be stored in. `teams-api` must have write access to this directory, all plugin nodes must have read access to this directory.                                                                                                                                                                                                                                                                                                                                                                               | No       |
| `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`            | Full path to network-mounted file system or a cloud storage path to use for snapshot archive storage. The default `None` means archival is disabled.                                                                                                                                                                                                                                                                                                                                                                                           | No       |
| `FIFTYONE_SNAPSHOTS_MAX_IN_DB`               | The max total number of Snapshots allowed at once. -1 for no limit. If this limit is exceeded then automatic archival is triggered if enabled, otherwise an error is raised.                                                                                                                                                                                                                                                                                                                                                                    | No       |
| `FIFTYONE_SNAPSHOTS_MAX_PER_DATASET`         | The max number of Snapshots allowed per dataset. -1 for no limit. If this limit is exceeded then automatic archival is triggered if enabled, otherwise an error is raised.                                                                                                                                                                                                                                                                                                                                                                      | No       |
| `FIFTYONE_SNAPSHOTS_MIN_LAST_LOADED_SEC`     | The minimum last-loaded age in seconds (as defined by `now-last_loaded_at`) a snapshot must meet to be considered for automatic archival. This limit is intended to help curtail automatic archival of a snapshot a user is actively working with. The default value is 1 day.                                                                                                                                                                                                                                                                  | No       |
| `FIFTYONE_TEAMS_PROXY_URL`                   | The URL that `fiftyone-teams-app` will use to proxy requests to `fiftyone-app`                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Yes      |
| `GRAPHQL_DEFAULT_LIMIT`                      | Default GraphQL limit for results                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | No       |
| `HTTP_PROXY_URL`                             | The URL for your environment http proxy                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | No       |
| `HTTPS_PROXY_URL`                            | The URL for your environment https proxy                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | No       |
| `NO_PROXY_LIST`                              | The list of servers that should bypass the proxy; if a proxy is in use this must include the list of FiftyOne services (`teams-api, teams-app, fiftyone-app`)                                                                                                                                                                                                                                                                                                                                                                                   | No       |
