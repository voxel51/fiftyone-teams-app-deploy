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

# fiftyone-teams-app

FiftyOne Enterprise is the enterprise version of the open source
[FiftyOne](https://github.com/voxel51/fiftyone)
project.

The FiftyOne Enterprise Docker Compose files are the recommended way to
install and configure FiftyOne Enterprise on Docker.

This page assumes general knowledge of FiftyOne Enterprise and how to use it.
Please contact Voxel51 for more information regarding FiftyOne Enterprise.

## Important

### Version 2.0+ License File Requirement

FiftyOne Enterprise v2.0 introduces a new requirement for a license file.
This license file should be obtained from your Customer Success Team
before upgrading to FiftyOne Enterprise 2.0 or beyond.

Please refer to the
[upgrade documentation](./docs/upgrading.md#from-fiftyone-enterprise-versions-160-to-171)
for steps on how to add your license file.

### Version 2.2+ Delegated Operator Changes

FiftyOne Enterprise v2.2 introduces some changes to delegated operators.
Please refer to the
[upgrade documentation](./docs/upgrading.md#fiftyone-enterprise-v22-delegated-operator-changes)
for steps on how to upgrade your delegated operators.

### Version 2.5+ Delegated Operator Changes

FiftyOne Enterprise v2.5 introduces some changes to delegated operators.
Please refer to the
[upgrade documentation](./docs/upgrading.md#fiftyone-enterprise-v25-delegated-operator-changes)
for steps on how to upgrade your delegated operators.

## Table of Contents

<!-- toc -->

- [Requirements](#requirements)
- [Usage](#usage)
- [Initial Installation vs. Upgrades](#initial-installation-vs-upgrades)
- [Known Issues](#known-issues)
- [Advanced Configuration](#advanced-configuration)
  - [Builtin Delegated Operator Orchestrator](#builtin-delegated-operator-orchestrator)
  - [Central Authentication Service](#central-authentication-service)
  - [FiftyOne Enterprise Authenticated API](#fiftyone-enterprise-authenticated-api)
  - [GPU Enabled Workloads](#gpu-enabled-workloads)
  - [Plugins](#plugins)
  - [Proxies](#proxies)
  - [Snapshot Archival](#snapshot-archival)
  - [Static Banner Configuration](#static-banner-configuration)
  - [Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](#storage-credentials-and-fiftyone_encryption_key)
  - [Terms of Service, Privacy, and Imprint URLs](#terms-of-service-privacy-and-imprint-urls)
  - [Text Similarity](#text-similarity)
- [Validating](#validating)
- [FiftyOne Enterprise Environment Variables](#fiftyone-enterprise-environment-variables)

<!-- tocstop -->

## Requirements

[Docker Compose](https://docs.docker.com/compose/install/)
must be installed and configured on your machine.

## Usage

FiftyOne Enterprise v2.0 introduces a new requirement for a license file.  This
license file should be obtained from your Customer Success Team before
upgrading to FiftyOne Enterprise 2.0 or beyond.

The license file now contains all of the Auth0 configuration that was
previously provided through environment variables; you may remove those secrets
from your `.env` and from any secrets created outside of the Voxel51
install process.

Set the `LOCAL_LICENSE_FILE_DIR` value in your .env file and copy the license
file to the `LOCAL_LICENSE_FILE_DIR` directory on your FiftyOne Enterprise docker
compose host.
e.g.:

```shell
. .env
mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
```

> [!TIP]
> When rotating the license, to ensure that the new license values are
> picked up immediately, you may need to restart the `teams-cas` and
> `teams-api` services.

We publish the following FiftyOne Enterprise private images to Docker Hub:

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch`
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`
- `voxel51/fiftyone-teams-cas`
- `voxel51/fiftyone-teams-cv-full`

For Docker Hub credentials, please contact your Voxel51 support team.

To deploy FiftyOne Enterprise:

1. Choose to install using `legacy-auth` (recommended) or `internal-auth` by
   `cd`ing into either the `legacy-auth` or `internal-auth` subdirectory in this
   repository.
1. In the directory chosen above
    1. Rename the `env.template` file to `.env`
    1. Edit the `.env` file, setting all the customer provided required
    settings.
       See the
       [FiftyOne Enterprise Environment Variables](#fiftyone-enterprise-environment-variables)
       table.
    1. Create a `compose.override.yaml` with any configuration overrides for
       this deployment
        1. For the first installation, set

            ```yaml
            services:
              fiftyone-app:
                environment:
                  FIFTYONE_DATABASE_ADMIN: true
            ```

1. Make sure you have put your Voxel51-provided FiftyOne Enterprise license in the
   local directory identified by the `LOCAL_LICENSE_FILE_DIR` configured in
   your `.env` file.
1. Deploy FiftyOne Enterprise
    1. In the same directory, run

        ```shell
        docker compose up -d
        ```

1. After the successful installation, and logging into FiftyOne Enterprise
    1. In `compose.override.yaml`, remove the `FIFTYONE_DATABASE_ADMIN` override

        ```yaml
        services:
          fiftyone-app:
            environment:
              # FIFTYONE_DATABASE_ADMIN: true
        ```

        > **NOTE**: This example shows commenting this line,
        > however you may remove the line.

        or set it to `false` like in

        ```yaml
        services:
          fiftyone-app:
            environment:
              FIFTYONE_DATABASE_ADMIN: false
        ```

The FiftyOne Enterprise API is exposed on port `8000`.
The FiftyOne Enterprise App is exposed on port `3000`.
The FiftyOne Enterprise CAS is exposed on port `3030`.

Configure an SSL endpoint (like a Load Balancer, Nginx Proxy, or similar)
to route traffic to the appropriate endpoints. An example Nginx configuration
for path-based routing can be found
[here](./example-nginx-path-routing.conf).
Example Nginx configurations for hostname-based routing can be found
[here](./example-nginx-site.conf)
for FiftyOne Enterprise App and FiftyOne Enterprise CAS services, and
[here](./example-nginx-api.conf)
for the FiftyOne Enterprise API service.

## Initial Installation vs. Upgrades

When performing an initial installation, in `compose.override.yaml` set
`services.fiftyone-app.environment.FIFTYONE_DATABASE_ADMIN: true`.
When performing a FiftyOne Enterprise upgrade, set
`services.fiftyone-app.environment.FIFTYONE_DATABASE_ADMIN: false`.
See
[Upgrading From Previous Versions](./docs/upgrading.md)

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

When performing an upgrade, please review
[Upgrading From Previous Versions](./docs/upgrading.md)

## Known Issues

Please refer to the
[known-issues documentation](./docs/known-issues.md)
for common issues and their resolution.
For new issues, please submit a GitHub issue on the
[repository](https://github.com/voxel51/fiftyone-teams-app-deploy/issues).

## Advanced Configuration

### Builtin Delegated Operator Orchestrator

FiftyOne Enterprise v2.2 introduces a builtin orchestrator to run
[Delegated Operations](https://docs.voxel51.com/enterprise/enterprise_plugins.html#delegated-operations),
instead of (or in addition to) configuring your own orchestrator such as Airflow.

For configuring your delegated operators, see
[Configuring Delegated Operators](./docs/configuring-delegated-operators.md).

### Central Authentication Service

FiftyOne Enterprise v1.6 introduces the Central Authentication Service (CAS).
CAS requires additional configurations and consumes additional resources.
Please review these notes, and the
[Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
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

To upgrade from versions prior to FiftyOne Enterprise v1.6

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
> [Upgrading From Previous Versions](./docs/upgrading.md)

### FiftyOne Enterprise Authenticated API

FiftyOne Enterprise v1.3 introduces the capability to connect FiftyOne
Enterprise SDK through the FiftyOne Enterprise API (instead of creating a
direct connection to MongoDB).

To enable the FiftyOne Enterprise Authenticated API you will need to
[expose the FiftyOne Enterprise API endpoint](./docs/expose-teams-api.md)
and
[configure your SDK](https://docs.voxel51.com/enterprise/api_connection.html).

### GPU Enabled Workloads

FiftyOne services can be scheduled on GPU-enabled hardware for more efficient
computation.

To schedule pods on GPU-enabled hardware, see the
[configuring GPU workloads documentation](./docs/configuring-gpu-workloads.md).

### Plugins

FiftyOne Enterprise v1.3+ includes significant enhancements for
[Plugins](https://docs.voxel51.com/plugins/index.html)
to customize and extend the functionality of FiftyOne Enterprise in your environment.

There are three modes for plugins

1. Builtin Plugins Only
    - This is the default mode
    - Users may only run the builtin plugins shipped with FiftyOne Enterprise
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
[Custom Plugins Images](../docs/custom-plugins.md).

To use the FiftyOne Enterprise UI to deploy plugins, navigate to
`https://<DEPLOY_URL>/settings/plugins`. Early-adopter plugins installed
manually must be redeployed using the FiftyOne Enterprise UI.

For configuring your plugins, see
[Configuring Plugins](./docs/configuring-plugins.md).

### Proxies

FiftyOne Enterprise supports routing traffic through proxy servers.
Please refer to the
[proxy configuration documentation](./docs/configuring-proxies.md)
for information on how to configure proxies.

### Snapshot Archival

Since version v1.5, FiftyOne Enterprise supports
[archiving snapshots](https://docs.voxel51.com/enterprise/dataset_versioning.html#snapshot-archival)
to cold storage locations to prevent filling up the MongoDB database.
Supported locations are network mounted filesystems and cloud storage folders.

Please refer to the
[snapshot archival configuration documentation](./docs/configuring-snapshot-archival.md)
for configuring snapshot archival.

### Static Banner Configuration

FiftyOne Enterprise v2.6 introduces the ability to add a static banner to the
application.

Banner text is configured with `FIFTYONE_APP_BANNER_TEXT`.

Banner background color is configured with `FIFTYONE_APP_BANNER_COLOR`.

Banner text color is configured with:
`casSettings.env.FIFTYONE_APP_BANNER_TEXT_COLOR` and
`teamsAppSettings.env.FIFTYONE_APP_BANNER_TEXT_COLOR`

Examples:

```yaml
services:
  teams-app-common:
    environment:
      FIFTYONE_APP_BANNER_COLOR: `green | rgb(34,139,34") | '#f1f1f1'`
      FIFTYONE_APP_BANNER_TEXT_COLOR: `green | rgb(34,139,34") | '#f1f1f1'`
      FIFTYONE_APP_BANNER_TEXT: "Internal Deployment"
  teams-cas-common:
    environment:
      FIFTYONE_APP_BANNER_COLOR: `green | rgb(34,139,34") | '#f1f1f1'`
      FIFTYONE_APP_BANNER_TEXT_COLOR: `green | rgb(34,139,34") | '#f1f1f1'`
      FIFTYONE_APP_BANNER_TEXT: "Internal Deployment"
```

### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

As of FiftyOne Enterprise 1.1, containers based on the
`fiftyone-teams-cas`, `fiftyone-teams-api` and `fiftyone-app` images must
include the `FIFTYONE_ENCRYPTION_KEY` variable. This key is used to
encrypt storage credentials in the MongoDB database.

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
Users with `Admin` permissions may use the FiftyOne Enterprise UI
to manage storage credentials by navigating to
`https://<DEPOY_URL>/settings/cloud_storage_credentials`.

FiftyOne Enterprise version 1.3+ continues to support the use of environment
variables to set storage credentials in the application context and is
providing an alternate configuration path.

### Terms of Service, Privacy, and Imprint URLs

FiftyOne Enterprise v2.6 introduces the ability to override
the Terms of Service, Privacy, and Imprint (optional) links
if required in the App.

Configure the URLs by setting the following environment variables in
your `compose.override.yaml`.

Terms of Service URL is configured with `FIFTYONE_APP_TERMS_URL`.

Privacy URL is configured with `FIFTYONE_APP_PRIVACY_URL`.

Imprint/Impressum URL is configured with `FIFTYONE_APP_IMPRINT_URL`

Examples:

```yaml
services:
  teams-app-common:
    environment:
      FIFTYONE_APP_TERMS_URL: https://abc.com/tos
      FIFTYONE_APP_PRIVACY_URL: https://abc.com/privacy
      FIFTYONE_APP_IMPRINT_URL: https://abc.com/imprint
```

### Text Similarity

FiftyOne Enterprise version 1.2 and higher supports using text similarity searches
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
    image: voxel51/fiftyone-app-torch:v2.9.0
```

For more information, see the docs for
[Docker Compose Extend](https://docs.docker.com/compose/extends/).

---

## Validating

After deploying FiftyOne Enterprise and configuring authentication, please
follow
[validating your deployment](../docs/validating-deployment.md).

## FiftyOne Enterprise Environment Variables

| Variable                                     | Purpose                                                                                                                                                                                                                                                                        | Required                  |
|----------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------|
| `API_BIND_ADDRESS`                           | The host address that `fiftyone-teams-api` should bind to; `127.0.0.1` is appropriate for this in most cases                                                                                                                                                                   | Yes                       |
| `API_BIND_PORT`                              | The host port that `fiftyone-teams-api` should bind to; the default is `8000`                                                                                                                                                                                                  | Yes                       |
| `API_LOGGING_LEVEL`                          | Logging Level for `teams-api` service                                                                                                                                                                                                                                          | Yes                       |
| `API_URL`                                    | The URL that `fiftyone-teams-app` should use to communicate with `fiftyone-teams-api`; `teams-api` is the compose service name                                                                                                                                                 | Yes                       |
| `APP_BIND_ADDRESS`                           | The host address that `fiftyone-teams-app` should bind to; `127.0.0.1` is appropriate in most cases                                                                                                                                                                            | Yes                       |
| `APP_BIND_PORT`                              | The host port that `fiftyone-teams-app` should bind to the default is `3000`                                                                                                                                                                                                   | Yes                       |
| `APP_USE_HTTPS`                              | Set this to true if your Application endpoint uses TLS; this should be `true` in most cases'                                                                                                                                                                                   | Yes                       |
| `BASE_URL`                                   | The URL where you plan to deploy your FiftyOne Enterprise                                                                                                                                                                                                                      | `internal` auth mode only |
| `CAS_BASE_URL`                               | The URL that FiftyOne Enterprise Services should use to communicate with `teams-cas`; `teams-cas` is the compose service                                                                                                                                                       | Yes                       |
| `CAS_BIND_ADDRESS`                           | The host address that `teams-cas` should bind to; `127.0.0.1` is appropriate in most cases                                                                                                                                                                                     | Yes                       |
| `CAS_BIND_PORT`                              | The host port that `teams-cas` should bind to; the default is `3030`                                                                                                                                                                                                           | Yes                       |
| `CAS_DATABASE_NAME`                          | The MongoDB Database that the `teams-cas` service should use; the default is `fiftyone-cas`                                                                                                                                                                                    | Yes                       |
| `CAS_DEBUG`                                  | The logs that `teams-cas` should provide to stdout; see [debug](https://www.npmjs.com/package/debug) for documentation                                                                                                                                                         | Yes                       |
| `CAS_DEFAULT_USER_ROLE`                      | The default role when users initially log into the FiftyOne Enterprise application; the default is `GUEST`                                                                                                                                                                     | Yes                       |
| `CAS_MONGODB_URI`                            | The MongoDB Connection STring for CAS; this will default to `FIFTYONE_DATABASE_URI`                                                                                                                                                                                            | No                        |
| `FIFTYONE_APP_ALLOW_MEDIA_EXPORT`            | Set this to `"false"` if you want to disable media export options                                                                                                                                                                                                              | No                        |
| `FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED`   | Controls whether anonymous analytics are captured for the application. Set to false to opt-out of anonymous analytics.                                                                                                                                                         | No                        |
| `FIFTYONE_APP_BANNER_COLOR`                  | Global banner background color in App                                                                                                                                                                                                                                          | No                        |
| `FIFTYONE_APP_BANNER_TEXT_COLOR`             | Global banner text color in App                                                                                                                                                                                                                                                | No                        |
| `FIFTYONE_APP_BANNER_TEXT`                   | Global banner text in App                                                                                                                                                                                                                                                      | No                        |
| `FIFTYONE_APP_TERMS_URL`                     | Terms of Service URL used in App                                                                                                                                                                                                                                               | No                        |
| `FIFTYONE_APP_PRIVACY_URL`                   | Privacy URL used in App                                                                                                                                                                                                                                                        | No                        |
| `FIFTYONE_APP_IMPRINT_URL`                   | Imprint URL used in App                                                                                                                                                                                                                                                        | No                        |
| `FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION` | The recommended fiftyone SDK version. This will be displayed in install modal (i.e. `pip install ... fiftyone==0.11.0`)                                                                                                                                                        | No                        |
| `FIFTYONE_APP_THEME`                         | The default theme configuration for your FiftyOne Enterprise application as described [here](https://docs.voxel51.com/user_guide/config.html#configuring-the-app)                                                                                                              | No                        |
| `FIFTYONE_APP_DEFAULT_QUERY_PERFORMANCE`     | Controls whether Query Performance mode is enabled by default for every dataset for the application. Set to false to set default mode to off.                                                                                                                                  | No                        |
| `FIFTYONE_APP_ENABLE_QUERY_PERFORMANCE`      | Controls whether Query Performance mode is enabled for the application. Set to false to disable Query Performance mode for entire application.                                                                                                                                 | No                        |
| `FIFTYONE_API_URI`                           | The URI to be displayed in the `Install FiftyOne` Modal and `API Keys` configuration screens                                                                                                                                                                                   | No                        |
| `FIFTYONE_AUTH_SECRET`                       | The secret used for services to authenticate with `teams-cas`; also used to login to the SuperAdmin UI                                                                                                                                                                         | Yes                       |
| `FIFTYONE_DATABASE_NAME`                     | The MongoDB Database that `fiftyone-app`, `teams-api`, and `teams-app` use for FiftyOne Enterprise dataset metadata; the default is `fiftyone`                                                                                                                                 | Yes                       |
| `FIFTYONE_DATABASE_URI`                      | The MongoDB Connection String for FiftyOne Enterprise dataset metadata                                                                                                                                                                                                         | Yes                       |
| `FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH` | Full path to a network-mounted file system or a cloud storage path to use for storing logs generated by delegated operation runs, one file per job. The default `null` means log upload is disabled.                                                                           | No                        |
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
| `LOCAL_LICENSE_FILE_DIR`                     | Location of the directory that contains the FiftyOne Enterprise license file on the local server.                                                                                                                                                                              | Yes                       |
| `NO_PROXY_LIST`                              | The list of servers that should bypass the proxy; if a proxy is in use this must include the list of FiftyOne services (`fiftyone-app, teams-api,teams-app,teams-cas` must be included, `teams-plugins` should be included for dedicated plugins configurations)               | No                        |

<!-- Reference Links -->
[internal-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#internal-mode
[legacy-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#legacy-mode
