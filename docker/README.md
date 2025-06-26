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

# FiftyOne Enterprise: Docker Deployment Guide

FiftyOne Enterprise is the enterprise version of the open source
[FiftyOne](https://github.com/voxel51/fiftyone)
project.

The FiftyOne Enterprise Docker Compose files are the recommended way to
install and configure FiftyOne Enterprise on Docker.


This guide walks you through the steps for installing FiftyOne Enterprise 
using Docker Compose. It also includes advanced configuration, environment variables, 
and upgrade considerations. This page assumes general knowledge of FiftyOne Enterprise and how to use it.
Please contact Voxel51 for more information regarding FiftyOne Enterprise.

## Table of Contents

<!-- toc -->

- [Prerequisites](#prerequisites)
- [Step 1: Prepare License File](#step-1-prepare-license-file)
- [Step 2: Choose Authentication Mode](#step-2-choose-authentication-mode)
- [Step 3: Configure Environment](#step-3-configure-environment)
- [Step 4: Initial Deployment](#step-4-initial-deployment)
- [Step 5: Configure SSL & Reverse Proxy (Nginx / Load Balancer)](#step-5-ssl-reverse-proxy)
- [Step 6: Configuring FiftyOne Enterprise Plugins](#step-6-configuring-plugins)
- [Step 7: Configuring FiftyOne Enterprise Delegated Operators](#step-7-delegated-operators)
- [Step 8: Configuring Authentication (CAS)](#configuring-auth)
- [Upgrades](#upgrades)
- [Known Issues](#known-issues)
- [Advanced Configuration](#advanced-configuration)
- [Environment Variables](#environment-variables)

<!-- tocstop -->

## Prerequisites

- Docker and Docker Compose are installed
- License file from Voxel51
- Docker Hub credentials from Voxel51
- MongoDB instance available. 
  - FiftyOne Teams is compatible with MongoDB Community, Enterprise, or Atlas Editions.
  - If using MongoDB Community or Enterprise we recommend a minimum of 4vCPU and 16GB of RAM. Large datasets and 
  complex samples may require additional resources.
  - If using Atlas we recommend starting on at least a M40 cluster tier - you can then use utilization metrics to 
  make scaling decisions (up or down). Please note that we do not   support MongoDB Atlas Serverless instances 
  because we require Aggregations.


## Step 1: Prepare License File

> Required for **v2.0+**

1. Set `LOCAL_LICENSE_FILE_DIR` in your `.env`
2. Place the license file there and rename it to `license`

```bash
. .env
mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
```

> [!TIP]
> When rotating the license, to ensure that the new license values are
> picked up immediately, you may need to restart the `teams-cas` and
> `teams-api` services.

## Step 2: Choose Authentication Mode

FiftyOne Enterprise offers two authentication modes:

- `legacy-auth` → Choose this mode if using Auth0 for user authentication and authorization.
- `internal-auth` → Choose this mode when in an airgapped deployment - aigapped deployments 
will not require network egress to external services.

### 👉 Choose your mode

Navigate into the appropriate directory:

```bash
cd legacy-auth       # or internal-auth
```

## Step 3: Configure Environment

### 1. Copy the template `.env` file:

```
cp env.template .env
```

### 2. Fill out required values in `.env`

At minimum, configure:

- `BASE_URL` / `AUTH0_BASE_URL` - Your WebApp URL
- `FIFTYONE_API_URI` - Your API URL (can be the same as webApp URL if using path-based routing)
- `FIFTYONE_DATABASE_URI` – Your MongoDB connection URI
- `FIFTYONE_ENCRYPTION_KEY` – Used to encrypt storage credentials

> 🔑 To generate a key:

```
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
```

- `LOCAL_LICENSE_FILE_DIR` – Path where the license is mounted
- `FIFTYONE_AUTH_SECRET` – Shared secret for CAS and App auth
- Any other variables noted in the `.env.template` or listed in the [Environment Variables](#environment-variables) section

### 3. Create a `compose.override.yaml` to override configuration.

```
services:
  fiftyone-app:
    environment:
      FIFTYONE_DATABASE_ADMIN: true  # Only for first install
```

### 📦 Official Docker Images

Voxel51 publishes the following private FiftyOne Enterprise images to Docker Hub:

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch` ← for text similarity / PyTorch support
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`
- `voxel51/fiftyone-teams-cas`
- `voxel51/fiftyone-teams-cv-full` ← full CV/ML environment

> 🔐 For access, contact your Voxel51 support team to obtain Docker Hub credentials.

You can override the default image used by any service in `compose.override.yaml`. For example:

```
services:
  fiftyone-app:
    image: voxel51/fiftyone-app-torch:v2.10.0
```

## Step 4: Initial Deployment

### 1. Enable Database Admin mode

In `compose.override.yaml`, make sure:

```
services:
  fiftyone-app:
    environment:
      FIFTYONE_DATABASE_ADMIN: true
```

> This allows the application to create and migrate the database schema.

### 2. Launch the application

In the same directory:

```
docker compose up -d
```

This will start the following containers:

- `fiftyone-teams-app` (UI) → default port `3000`
- `fiftyone-teams-api` (API) → default port `8000`
- `fiftyone-teams-cas` (Auth) → default port `3030`


You can ensure that all your containers are up and healthy through:

```
docker compose ps
```

For HTTP-based health checks, run the following `curl` commands:

```
curl -Iv http://localhost:3030/cas/api
# Expected: HTTP/1.1 200 OK

curl -Iv http://localhost:8000/health
# Expected: HTTP/1.1 200 OK

curl -Iv http://localhost:3000/api/hello
# Expected: HTTP/1.1 200 OK
```

## Step 5: Configure SSL & Reverse Proxy (Nginx / Load Balancer)

Next, you will need to place a **reverse proxy or SSL endpoint** in front of your FiftyOne 
Enterprise services. This can be a tool like:

- **Nginx**
- **HAProxy**
- **Cloud Load Balancer** (e.g., AWS ALB, GCP Load Balancer)

These proxies will:

- Route traffic to the correct services (`teams-app`, `teams-api`, `teams-cas`)
- Terminate HTTPS/SSL (if using TLS)
- Optionally apply authentication headers, logging, or load balancing

### 🧭 Routing Overview (Path-Based Proxy)

| Path              | Proxied To       | Description                          |
|-------------------|------------------|--------------------------------------|
| `/`               | `teams-app`      | Main web UI                          |
| `/cas`            | `teams-cas`      | Central Authentication Service (CAS) |
| `/graphql/v1`     | `teams-api`      | GraphQL API endpoint                 |
| `/file`           | `teams-api`      | File import handling                 |
| `/_pymongo`       | `teams-api`      | MongoDB requests via SDK             |
| `/health`         | `teams-api`      | Health check endpoint                |

### 📁 Nginx Configuration Options

Voxel51 provides example Nginx configs for two routing strategies:

#### 🔹 1. **Path-Based Routing**
All services are routed based on URL path:

📄 Full configuration here: [`example-nginx-path-routing.conf`](./example-nginx-path-routing.conf)

#### 🔹 2. **Hostname-Based Routing**
teams-app and teams-api are routed using different subdomain or hostname:

- `fiftyone.your-company.com` → App
- `fiftyone-api.your-company.com` → API

📄 Full configuration here:
- [`example-nginx-site.conf`](./example-nginx-site.conf) (App + CAS)
- [`example-nginx-api.conf`](./example-nginx-api.conf) (API)


### 📌 Notes

- FiftyOne Enterprise supports routing traffic through proxy servers. Please refer to the 
[proxy configuration documentation](./docs/configuring-proxies.md) for information on how to configure proxies.
- To validate your deployments api connection, see [Validating Your Deployment](../docs/validating-deployment.md)


## Step 6: Configuring FiftyOne Enterprise Plugins

FiftyOne Enterprise supports three plugin modes: **Builtin**, **Shared**, and **Dedicated**. Each offers different 
trade-offs in isolation, flexibility, and resource management.

### 🔹 1. Builtin Plugins Only (Default)

This is the default configuration. It enables only the plugins shipped with the platform.

✅ No additional configuration needed.

### 🔹 2. Shared Plugins (Custom Plugins in `fiftyone-app`)

Custom plugins are run **within the same container** as the app (`fiftyone-app`). Use this if:

- You want to quickly prototype plugins
- You’re okay with shared resource usage between app and plugins

#### Enable shared plugin mode:

1. Use `compose.plugins.yaml` (instead of `compose.yaml`)
2. This mounts a shared volume for plugins across services

```
docker compose \
  -f compose.plugins.yaml \
  -f compose.override.yaml \
  up -d
```

> 📁 Plugins will run inside the `fiftyone-app` container.

### 🔹 3. RECOMMENDED: Dedicated Plugins (Isolated `teams-plugins` Service)

Custom plugins are run in a **separate `teams-plugins` container**, isolated from the app and API services.

Use this mode when:

- You want full isolation and stability
- You are running many or complex plugins
- You need to manage plugin memory or compute separately

#### Enable dedicated plugin mode:

1. Ensure your `.env` file includes the following:

```
FIFTYONE_TEAMS_PLUGIN_URL=http://teams-plugins:5151
```

2. Use `compose.dedicated-plugins.yaml` (instead of `compose.yaml`)

```
docker compose \
  -f compose.dedicated-plugins.yaml \
  -f compose.override.yaml \
  up -d
```

3. Optional: If using a [proxy server](./docs/configuring-proxies.md), ensure the plugin service is excluded from proxying.

> 🔧 This prevents traffic from being routed incorrectly through your proxy for internal plugin calls.

### 📌 Notes

- All plugin modes require persistent storage (volumes) for plugin files.
- For multi-node deployments, ensure that the volume is available on all nodes.
- To manage and deploy plugins via the UI, go to:  
  `https://<your-domain>/settings/plugins`

## Step 7: Configuring FiftyOne Enterprise Delegated Operators

Delegated Operators allow FiftyOne Enterprise to offload plugin execution to **worker containers**, enabling 
scalable and reliable long-running operations.

🧩 This feature is **compatible with all three plugin modes**: Builtin, Shared, and Dedicated.

### 🔧 Enabling Delegated Operator Mode

To launch worker containers, include `compose.delegated-operators.yaml` alongside your existing plugin mode.

#### Example: Enable on top of **Dedicated Plugins** mode

```
docker compose \
  -f compose.dedicated-plugins.yaml \
  -f compose.delegated-operators.yaml \
  -f compose.override.yaml \
  up -d
```

> 📁 This will start a `teams-delegated-operator` service and attach it to the shared plugin volume.

### 📄 Optional: Upload Run Logs

You can enable **log uploads** for delegated operation runs by setting:

```
FIFTYONE_DELEGATED_OPERATION_LOG_PATH=/mnt/shared/logs
```

Logs will be stored in the format:

```
/mnt/shared/logs/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log
```

This is useful for auditing, debugging, or monitoring delegated operator executions in shared storage or cloud buckets.

### 🖥️ GPU-Enabled Workloads

FiftyOne services like Delegated Operators can be scheduled on **GPU-enabled hardware** for more efficient computation.

To setup containers with GPU resources, see the  
[configuring GPU workloads documentation](./docs/configuring-gpu-workloads.md).

### 🧱 Custom Plugin Images

If your delegated operators or plugins require **custom dependencies**, build and deploy **custom plugin images**. 
You can base them on `voxel51/fiftyone-app`, and include:

- Custom Python packages
- ML libraries (e.g. PyTorch, OpenCV)
- Internal SDKs or models


## Step 8: Configuring Authentication (CAS)


FiftyOne Enterprise uses a Central Authentication Service (CAS) introduced in v1.6. This enables centralized 
login, roles, and user management.

### 🛠️ Optional: CAS Customization Instructions

1. Update required CAS variables in `.env`:
   - `FIFTYONE_AUTH_SECRET`
   - `CAS_BASE_URL`
   - `CAS_BIND_ADDRESS`
   - `CAS_BIND_PORT`
   - `CAS_DATABASE_NAME` (⚠️ Must be unique per deployment)
   - `CAS_DEBUG`
   - `CAS_DEFAULT_USER_ROLE`
1. Update `compose.override.yaml` with any needed `teams-cas` service changes.
1. Use `docker compose` from within the `legacy-auth`/`internal-auth` directory to bring up services.
1. Ensure your proxy (e.g., nginx) forwards `/cas` to the CAS service port.

### ℹ️ Notes

- [Pluggable authentication docs](https://docs.voxel51.com/enterprise/pluggable_auth.html#pluggable-authentication) 
includes information on configuring CAS.
- To set up authentication for internal-auth mode: Refer to the 
[Getting Started with Internal Mode documentation](https://docs.voxel51.com/enterprise/pluggable_auth.html#getting-started-with-internal-mode).


## Upgrades

When upgrading FiftyOne Enterprise, you must explicitly **prevent automatic database migrations** 
to avoid breaking active SDK sessions or deployments.

### 🚫 Disable Automatic Migrations

Before running your upgraded containers, set the following override:

"code"
services:
  fiftyone-app:
    environment:
      FIFTYONE_DATABASE_ADMIN: false
"code"

> This ensures that **no automatic migrations** will occur when the container starts.

The environment variable `FIFTYONE_DATABASE_ADMIN` acts as a safeguard to prevent the database from being modified automatically.

### 🛠️ What Happens If You Migrate with database admin False?

If `FIFTYONE_DATABASE_ADMIN=false` is set, and a migration attempt is made via CLI:

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

### 📚 Next Steps

After disabling `FIFTYONE_DATABASE_ADMIN`, refer to:

[Upgrading](./docs/upgrading.md)

for complete guidance on upgrading from previous versions

## Known Issues

For a list of common issues and their solutions, refer to the  
[📄 Known Issues documentation](./docs/known-issues.md).

If you encounter a new issue, please open a ticket on the  
[📬 GitHub Issues page](https://github.com/voxel51/fiftyone-teams-app-deploy/issues).


## Advanced Configuration

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


## Environment Variables

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
