<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img alt="Voxel51 Logo" src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px">
&nbsp;
<img alt="Voxel51 FiftyOne" src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# FiftyOne Enterprise: Docker Deployment Guide

FiftyOne Enterprise is the enterprise version of the open source
[FiftyOne](https://github.com/voxel51/fiftyone) project.

The FiftyOne Enterprise Docker Compose files are the recommended way to install
and configure FiftyOne Enterprise on Docker.

This guide walks you through the steps for installing FiftyOne Enterprise using
Docker Compose. It also includes advanced configuration, environment variables,
and upgrade considerations. This page assumes general knowledge of FiftyOne
Enterprise and how to use it. Please contact Voxel51 for more information
regarding FiftyOne Enterprise.

## Table of Contents

<!-- toc -->

- [:green_book: Prerequisites Skills and Knowledge](#green_book-prerequisites-skills-and-knowledge)
- [:white_check_mark: Technical Requirements](#white_check_mark-technical-requirements)
- [:clock10: Estimated Completion Time](#clock10-estimated-completion-time)
- [:floppy_disk: Sizing](#floppy_disk-sizing)
- [:wrench: Step 0: Set Up MongoDB Database](#wrench-step-0-set-up-mongodb-database)
- [:closed_lock_with_key: Step 1: Prepare License File](#closed_lock_with_key-step-1-prepare-license-file)
- [:gear: Step 2: Configure Environment](#gear-step-2-configure-environment)
  - [1. Copy the template `.env` file](#1-copy-the-template-env-file)
  - [2. Fill out required values in `.env`](#2-fill-out-required-values-in-env)
  - [3. Create a `compose.override.yaml` to override configuration](#3-create-a-composeoverrideyaml-to-override-configuration)
  - [:package: Official Docker Images](#package-official-docker-images)
- [:rocket: Step 3: Initial Deployment](#rocket-step-3-initial-deployment)
  - [1. Enable Database Admin mode](#1-enable-database-admin-mode)
  - [2. Launch the application](#2-launch-the-application)
- [:globe_with_meridians: Step 4: Configure SSL & Reverse Proxy (Nginx / Load Balancer)](#globe_with_meridians-step-4-configure-ssl--reverse-proxy-nginx--load-balancer)
  - [:compass: Routing Overview (Path-Based Proxy)](#compass-routing-overview-path-based-proxy)
  - [:open_file_folder: Nginx Configuration](#open_file_folder-nginx-configuration)
  - [:memo: Notes](#memo-notes)
- [:page_facing_up: Step 5: Configure Delegated Operation Logs](#page_facing_up-step-5-configure-delegated-operation-logs)
- [Step 6: Identity Provider (IdP) and Authentication (CAS) (Optional)](#step-6-identity-provider-idp-and-authentication-cas-optional)
  - [:information_source: IdP configuration](#information_source-idp-configuration)
  - [:hammer_and_wrench: Optional: CAS Customization Instructions](#hammer_and_wrench-optional-cas-customization-instructions)
- [Advanced DO Settings](#advanced-do-settings)
  - [:desktop_computer: GPU-Enabled Workloads](#desktop_computer-gpu-enabled-workloads)
  - [:bricks: Custom Plugin Images](#bricks-custom-plugin-images)
  - [:on: On-Demand Delegated Operator Executors](#on-on-demand-delegated-operator-executors)
- [Upgrades](#upgrades)
- [Known Issues](#known-issues)
- [Advanced Configuration](#advanced-configuration)
  - [Backup And Recovery](#backup-and-recovery)
  - [Secrets And Sensitive Data](#secrets-and-sensitive-data)
  - [Snapshot Archival](#snapshot-archival)
  - [Static Banner Configuration](#static-banner-configuration)
  - [Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](#storage-credentials-and-fiftyone_encryption_key)
  - [Terms of Service, Privacy, and Imprint URLs](#terms-of-service-privacy-and-imprint-urls)
- [Validating](#validating)
- [Health Checks And Monitoring](#health-checks-and-monitoring)
  - [Basic Health Assessment](#basic-health-assessment)
  - [Troubleshooting Unhealthy Containers](#troubleshooting-unhealthy-containers)
- [Environment Variables](#environment-variables)

<!-- tocstop -->

## :green_book: Prerequisites Skills and Knowledge

The following prerequisites skills & knowledge
are required for a successful and properly secured
deployment of FiftyOne Enterprise.

A knowledge of:

1. Docker compose.

1. MongoDB.

1. DNS and the ability to generate, modify, and delete DNS records.

1. TLS/SSL and the ability to generate TLS/SSL certificates.

## :white_check_mark: Technical Requirements

1. [Docker][docker-install]
   and
   [Docker Compose][docker-compose-install]
   installed.

1. License file from Voxel51.

1. [Docker Hub][docker-hub]
   credentials from Voxel51.

1. A MongoDB Database that meets FiftyOne's
  [version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).

1. A DNS record or records for ingress.

1. A TLS/SSL certificate or certificates for HTTPS ingress.

## :clock10: Estimated Completion Time

The estimated time to deploy FiftyOne Enterprise is approximately 2 hours.

## :floppy_disk: Sizing

Voxel51 recommends the following resource sizing:

- MongoDB: 4 CPU, 16GB RAM, 256GB Storage
- Docker Compose Server: 8 CPU, 32GB RAM, 256GB Storage

Voxel51 also recommends monitoring resource consumption across
the applications.
Resource usage varies dramatically with operations, use cases,
and dataset sizes.

## :wrench: Step 0: Set Up MongoDB Database

Before deploying FiftyOne Enterprise, you must have a running MongoDB database.
FiftyOne Enterprise supports:

- **MongoDB Atlas** (managed cloud)
- **MongoDB Community Edition** (self-hosted, open source)
- **MongoDB Enterprise** (self-hosted, commercial)

Ensure your MongoDB version meets FiftyOne's
[version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).

Once your database is running, record your MongoDB connection URI. You will
need it in Step 2 when configuring your `.env` file. The URI follows this
format:

```dotenv
FIFTYONE_DATABASE_URI="mongodb://username:password@mongodb-example.fiftyone.ai:27017/?authSource=admin"
```

## :closed_lock_with_key: Step 1: Prepare License File

> Required for **v2.0+**

1. Set `LOCAL_LICENSE_FILE_DIR` in your `.env`
2. Place the license file there and rename it to `license`
3. Ensure the license directory is volume-mounted into the containers
   (e.g., `/opt/data/licenses/`)
4. Set read permissions on the license file: `chmod 644 license`

```bash
# Set this to match the LOCAL_LICENSE_FILE_DIR value in your .env file
LOCAL_LICENSE_FILE_DIR="/path/to/your/licenses"
mkdir -p "${LOCAL_LICENSE_FILE_DIR}"
mv license.key "${LOCAL_LICENSE_FILE_DIR}/license"
chmod 644 "${LOCAL_LICENSE_FILE_DIR}/license"
```

> [!TIP]
> When rotating the license, to ensure that the new license values are
> picked up immediately, you may need to restart the `teams-cas` and `teams-api`
> services.

## :gear: Step 2: Configure Environment

Navigate into the `internal-auth` directory:

```bash
cd internal-auth
```

### 1. Copy the template `.env` file

```bash
cp env.template .env
```

### 2. Fill out required values in `.env`

At minimum, configure:

- `BASE_URL` / `AUTH0_BASE_URL` - Your WebApp URL
- `FIFTYONE_API_URI` - Your API URL (can be the same as webApp URL if using
  path-based routing)
- `FIFTYONE_DATABASE_URI` – Your MongoDB connection URI
- `FIFTYONE_ENCRYPTION_KEY` – Used to encrypt storage credentials

> :key: To generate a key:

```bash
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
```

- `LOCAL_LICENSE_FILE_DIR` – Path where the license is mounted
- `FIFTYONE_AUTH_SECRET` – Shared secret for CAS and App auth
- Any other variables noted in the `.env.template` or listed in the
  [Environment Variables](#environment-variables) section

### 3. Create a `compose.override.yaml` to override configuration

```yaml
services:
  fiftyone-app:
    environment:
      FIFTYONE_DATABASE_ADMIN: true # Only for first install
```

### :package: Official Docker Images

Voxel51 publishes the following private FiftyOne Enterprise images to Docker
Hub:

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch` ← for text similarity / PyTorch support
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`
- `voxel51/fiftyone-teams-cas`
- `voxel51/fiftyone-teams-cv-full` ← full CV/ML environment

> :closed_lock_with_key: For access, contact your Voxel51 support team to obtain
> Docker Hub credentials.

You can override the default image used by any service in
`compose.override.yaml`. For example:

```yaml
services:
  fiftyone-app:
    image: voxel51/fiftyone-app-torch:v2.18.1
```

> [!NOTE]
> Default images and versions for all services are defined in
> `common-services.yaml`. To override an image or version,
> set the value in `compose.override.yaml`. For example, instead of using
> the `fiftyone-app` image, you may instead the `fiftyone-app-gpt` image.
> Always include a version tag when overriding images (e.g., `:vX.Y.Z`).
> Omitting the tag will result in a **not found** error.

## :rocket: Step 3: Initial Deployment

### 1. Enable Database Admin mode

In `compose.override.yaml`, make sure:

```yaml
services:
  fiftyone-app:
    environment:
      FIFTYONE_DATABASE_ADMIN: false
```

> This allows the application to create and migrate the database schema.

### 2. Launch the application

FiftyOne Enterprise is deployed with **Dedicated Plugins** and **Delegated
Operators** enabled by default. Dedicated Plugins run in an isolated
`teams-plugins` container, and Delegated Operators offload plugin execution to
**worker containers**, enabling scalable and reliable long-running operations.

```shell
docker compose \
  -f compose.dedicated-plugins.yaml \
  -f compose.delegated-operators.yaml \
  -f compose.override.yaml \
  up -d
```

This will start the following containers:

- `fiftyone-app` (embedded API) → default port `5151`
- `fiftyone-teams-app` (UI) → default port `3000`
- `fiftyone-teams-api` (API) → default port `8000`
- `fiftyone-teams-cas` (Auth) → default port `3030`
- `fiftyone-teams-do-n` where n is the number of VPUs your in deployment

You can ensure that all your containers are up and healthy through:

```shell
docker compose ps
```

For HTTP-based health checks, run the following `curl` commands:

```shell
curl -Iv http://localhost:3030/cas/api
# Expected: HTTP/1.1 200 OK

curl -Iv http://localhost:8000/health
# Expected: HTTP/1.1 200 OK

curl -Iv http://localhost:3000/api/hello
# Expected: HTTP/1.1 200 OK
```

## :globe_with_meridians: Step 4: Configure SSL & Reverse Proxy (Nginx / Load Balancer)

Next, you will need to place a **reverse proxy or SSL endpoint** in front of
your FiftyOne Enterprise services. This can be a tool like:

- **Nginx**
- **HAProxy**
- **Cloud Load Balancer** (e.g., AWS ALB, GCP Load Balancer)

These proxies will:

- Route traffic to the correct services (`teams-app`, `teams-api`, `teams-cas`)
- Terminate HTTPS/SSL (if using TLS)
- Optionally apply authentication headers, logging, or load balancing

### :compass: Routing Overview (Path-Based Proxy)

| Path                 | Proxied To  | Description                          |
|----------------------| ----------- |--------------------------------------|
| `/`                  | `teams-app` | Main web UI                          |
| `/cas`               | `teams-cas` | Central Authentication Service (CAS) |
| `/cloud_credentials` | `teams-api` | Cloud credentials API endpoint       |
| `/graphql/v1`        | `teams-api` | GraphQL API endpoint                 |
| `/rpc`               | `teams-api` | RPC API endpoint                     |
| `/file`              | `teams-api` | File import handling                 |
| `/_pymongo`          | `teams-api` | MongoDB requests via SDK             |
| `/health`            | `teams-api` | Health check endpoint                |

### :open_file_folder: Nginx Configuration

All services are routed based on URL path using a single hostname. See the full
example configuration:
[`example-nginx-path-routing.conf`](./example-nginx-path-routing.conf)

### :memo: Notes

- FiftyOne Enterprise supports routing traffic through proxy servers. Please
  refer to the
  [proxy configuration documentation](./docs/configuring-proxies.md) for
  information on how to configure proxies.
- To validate your deployments api connection, see
  [Validating Your Deployment](../docs/validating-deployment.md)

## :page_facing_up: Step 5: Configure Delegated Operation Logs

Add the log path for delegated operation runs to your `compose.override.yaml`:

```yaml
services:
  teams-do:
    environment:
      FIFTYONE_DELEGATED_OPERATION_LOG_PATH: "gs://your-bucket/logs"
```

Logs are stored in the format:

```text
/mnt/shared/logs/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log
```

This is useful for auditing, debugging, or monitoring delegated operator
executions in shared storage or cloud buckets.

## Step 6: Identity Provider (IdP) and Authentication (CAS) (Optional)

### :information_source: IdP configuration

You can refer to the following docs to set up your Identity Provider with
FiftyOne.

- [Pluggable authentication docs](https://docs.voxel51.com/enterprise/pluggable_auth.html#pluggable-authentication)
  includes information on configuring CAS.
- To set up authentication for internal-auth mode: Refer to the
  [Getting Started with Internal Mode documentation](https://docs.voxel51.com/enterprise/pluggable_auth.html#getting-started-with-internal-mode).

FiftyOne Enterprise uses a Central Authentication Service (CAS) introduced in
v1.6. This enables centralized login, roles, and user management.

### :hammer_and_wrench: Optional: CAS Customization Instructions

1. Update required CAS variables in `.env`:
   - `FIFTYONE_AUTH_SECRET`
   - `CAS_BASE_URL`
   - `CAS_BIND_ADDRESS`
   - `CAS_BIND_PORT`
   - `CAS_DATABASE_NAME` (⚠️ Must be unique per deployment)
   - `CAS_DEBUG`
   - `CAS_DEFAULT_USER_ROLE`
1. Update `compose.override.yaml` with any needed `teams-cas` service changes.
1. Use `docker compose` from within the `legacy-auth`/`internal-auth` directory
   to bring up services.
1. Ensure your proxy (e.g., nginx) forwards `/cas` to the CAS service port.

## Advanced DO Settings

### :desktop_computer: GPU-Enabled Workloads

FiftyOne services like Delegated Operators can be scheduled on **GPU-enabled
hardware** for more efficient computation.

To setup containers with GPU resources, see the
[configuring GPU workloads documentation](./docs/configuring-gpu-workloads.md).

### :bricks: Custom Plugin Images

If your delegated operators or plugins require **custom dependencies**, build
and deploy **custom plugin images**. Base them on
`voxel51/fiftyone-teams-cv-full`, which includes a full CV/ML environment, and
extend with:

- Custom Python packages
- Internal SDKs or models

### :on: On-Demand Delegated Operator Executors

FiftyOne Enterprise v2.11 introduces support for on-demand delegated operator
executors for Databricks and Anyscale. Please refer to the
[configuration documentation](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/configuring-delegated-operators.md).

## Upgrades

The recommended upgrade path is:

1. Pull the latest compose files from this repo:

   ```shell
   git pull origin main
   ```

   This automatically populates the latest service versions.

1. Confirm `FIFTYONE_DATABASE_ADMIN` is set to `false` in your
   `compose.override.yaml`:

   ```yaml
   services:
     fiftyone-app:
       environment:
         FIFTYONE_DATABASE_ADMIN: false
   ```

   > This prevents automatic database migrations from running on startup and
   > breaking active SDK sessions.

1. If you are using [Custom Plugin Images](#bricks-custom-plugin-images),
   rebuild them using the updated base image version, then update their tags
   in `compose.override.yaml` to match the new release.

1. Bring the stack down and back up:

   ```shell
   docker compose \
     -f compose.dedicated-plugins.yaml \
     -f compose.delegated-operators.yaml \
     -f compose.override.yaml \
     down

   docker compose \
     -f compose.dedicated-plugins.yaml \
     -f compose.delegated-operators.yaml \
     -f compose.override.yaml \
     up -d
   ```

For full upgrade guidance, refer to [Upgrading](./docs/upgrading.md).

## Known Issues

For a list of common issues and their solutions, refer to the
[:page_facing_up: Known Issues documentation](./docs/known-issues.md).

If you encounter a new issue, please open a ticket on the
[:mailbox_with_mail: GitHub Issues page](https://github.com/voxel51/fiftyone-teams-app-deploy/issues).

## Advanced Configuration

### Backup And Recovery

FiftyOne Enterprise data is stored in MongoDB.
It is recommended to periodically backup the MongoDB instance
according to
[MongoDB best practices](https://www.mongodb.com/docs/manual/tutorial/backup-and-restore-tools/).

If needed, it is recommended to restore the most-recent backup
according to
[MongoDB best practices](https://www.mongodb.com/docs/manual/tutorial/backup-and-restore-tools/)

### Secrets And Sensitive Data

By default, database credentials, cookie secrets,
encryption keys, and authentication secrets are stored in environment variables.
This is configured by the following settings in `.env`:

```dotenv
# This should be a MongoDB Connection String for your database
FIFTYONE_DATABASE_URI="mongodb://username:password@mongodb-example.fiftyone.ai:27017/?authSource=admin"
# If you are using a different MongoDB Connection String for your CAS database,
#  set it here
# CAS_MONGODB_URI="mongodb://username:password@mongodb-cas-example.fiftyone.ai:27017/?authSource=admin"

# FIFTYONE_AUTH_SECRET is a random string used to authenticate to the CAS service
# This can be any string you care to use generated by any mechanism you prefer.
# You could use something like:
#  `cat /dev/urandom | LC_CTYPE=C tr -cd '[:graph:]' | head -c 32`
#  to generate this string.
# This is used for inter-service authentication and for the SuperUser to
#  authenticate at the CAS UI to configure the Central Authentication Service.
FIFTYONE_AUTH_SECRET=

# This key is required and is used to encrypt storage credentials in the MongoDB
#   do NOT lose this key!
# generate keys by executing the following in python:
#
# from cryptography.fernet import Fernet
# print(Fernet.generate_key().decode())
#
FIFTYONE_ENCRYPTION_KEY=

FIFTYONE_DEFAULT_APP_PORT=5151
```

FiftyOne Enterprise supports
[configuring secrets](https://docs.voxel51.com/enterprise/secrets.html#adding-secrets)
for the deployment to use.

Please see
[Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](#storage-credentials-and-fiftyone_encryption_key)
and
[adding secrets](https://docs.voxel51.com/enterprise/secrets.html#adding-secrets)
for questions regarding storage and encryption.

### Snapshot Archival

Since version v1.5, FiftyOne Enterprise supports
[archiving snapshots](https://docs.voxel51.com/enterprise/dataset_versioning.html#snapshot-archival)
to cold storage locations to prevent filling up the MongoDB database. Supported
locations are network mounted filesystems and cloud storage folders.

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

Containers based on the `fiftyone-teams-api` and `fiftyone-app`
images must include the `FIFTYONE_ENCRYPTION_KEY` variable.
This key is used to encrypt storage credentials in the MongoDB database.

To generate a value for `FIFTYONE_ENCRYPTION_KEY`, run this
Python code and add the output to your `.env` file.

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

Users with `Admin` permissions may use the FiftyOne Enterprise UI to manage storage
credentials by navigating to `https://<DEPOY_URL>/settings/cloud_storage_credentials`.

If added via the UI, storage credentials no longer need to be
mounted into containers or provided via environment variables.

FiftyOne Enterprise continues to support the use of environment variables to set
storage credentials in the application context and is providing an alternate
configuration path.

### Terms of Service, Privacy, and Imprint URLs

FiftyOne Enterprise v2.6 introduces the ability to override the Terms of
Service, Privacy, and Imprint (optional) links if required in the App.

Configure the URLs by setting the following environment variables in your
`compose.override.yaml`.

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

## Validating

After deploying FiftyOne Enterprise and configuring authentication, please
follow
[validating your deployment](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/validating-deployment.md).

## Health Checks And Monitoring

### Basic Health Assessment

Containers will report their status in the output of `docker compose ps`:

```shell
docker compose ps --format "table {{.Name}}\t{{.Service}}\t{{.RunningFor}}\t{{.Status}}"
```

Expected output for a healthy deployment:

```shell
NAME                      SERVICE         CREATED        STATUS
compose-fiftyone-app-1    fiftyone-app    15 hours ago   Up 15 hours
compose-fiftyone-app-2    fiftyone-app    15 hours ago   Up 15 hours
compose-fiftyone-app-3    fiftyone-app    15 hours ago   Up 15 hours
compose-teams-api-1       teams-api       15 hours ago   Up 15 hours
compose-teams-api-2       teams-api       15 hours ago   Up 15 hours
compose-teams-api-3       teams-api       15 hours ago   Up 15 hours
compose-teams-app-1       teams-app       15 hours ago   Up 15 hours
compose-teams-cas-1       teams-cas       15 hours ago   Up 15 hours
compose-teams-do-1        teams-do        15 hours ago   Up 15 hours
compose-teams-do-2        teams-do        15 hours ago   Up 15 hours
compose-teams-plugins-1   teams-plugins   15 hours ago   Up 15 hours
```

Note that number of containers and containers names may vary per deployment.

### Troubleshooting Unhealthy Containers

If containers show unhealthy states (e.g., `Restarting`, `Exited`):

1. **Get detailed container information**:

   ```shell
   docker inspect <container-name>
   ```

1. **Check application logs**:

   ```shell
   docker compose logs <service-name>
   ```

## Environment Variables

| Variable                                     | Purpose                                                                                                                                                                                                                                                                        | Required                  |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------- |
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
| `FIFTYONE_APP_THEME`                         | The default theme configuration for your FiftyOne Enterprise application as described [in our documentation](https://docs.voxel51.com/user_guide/config.html#configuring-the-app)                                                                                              | No                        |
| `FIFTYONE_APP_DEFAULT_QUERY_PERFORMANCE`     | Controls whether Query Performance mode is enabled by default for every dataset for the application. Set to false to set default mode to off.                                                                                                                                  | No                        |
| `FIFTYONE_APP_ENABLE_QUERY_PERFORMANCE`      | Controls whether Query Performance mode is enabled for the application. Set to false to disable Query Performance mode for entire application.                                                                                                                                 | No                        |
| `FIFTYONE_APP_DEPLOYMENT_CHARACTERISTICS`    | Deployment characteristics for the `teams-app`                                                                                                                                                                                                                                 | No                        |
| `FIFTYONE_API_URI`                           | The URI to be displayed in the `Install FiftyOne` Modal and `API Keys` configuration screens                                                                                                                                                                                   | No                        |
| `FIFTYONE_AUTH_SECRET`                       | The secret used for services to authenticate with `teams-cas`; also used to login to the SuperAdmin UI                                                                                                                                                                         | Yes                       |
| `FIFTYONE_DATABASE_NAME`                     | The MongoDB Database that `fiftyone-app`, `teams-api`, and `teams-app` use for FiftyOne Enterprise dataset metadata; the default is `fiftyone`                                                                                                                                 | Yes                       |
| `FIFTYONE_DATABASE_URI`                      | The MongoDB Connection String for FiftyOne Enterprise dataset metadata                                                                                                                                                                                                         | Yes                       |
| `FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH` | Full path to a network-mounted file system or a cloud storage path to use for storing logs generated by delegated operation runs, one file per job. The default `null` means log upload is disabled.                                                                           | No                        |
| `FIFTYONE_DEFAULT_APP_ADDRESS`               | The host address that `fiftyone-app` should bind to; `127.0.0.1` is appropriate in most cases                                                                                                                                                                                  | Yes                       |
| `FIFTYONE_DEFAULT_APP_PORT`                  | The host port that `fiftyone-app` should bind to; the default is `5151`                                                                                                                                                                                                        | Yes                       |
| `FIFTYONE_DO_REQUEUE_DELAY_MINUTES`          | The number of minutes an operation can sit in a queued state before being requeued. Can be any positive number. Default is `30`.                                                                                                                                               | No                        |
| `FIFTYONE_DO_REQUEUE_MAX_ATTEMPTS`           | The number of times an operation is requeued before being terminated. `-1` will never terminate. `0` will terminate without requeue. Default is `3`.                                                                                                                           | No                        |
| `FIFTYONE_DO_EXPIRATION_DAYS`                | Deprecated use `FIFTYONE_DO_LEGACY_EXPIRATION_MINUTES` instead. The amount of time in days that an unmonitored delegated operation can run before being automatically terminated. Default is 1 day.                                                                            | No                        |
| `FIFTYONE_DO_EXPIRATION_MINUTES`             | The amount of time in minutes that a monitored delegated operation can run without reporting its status before being automatically terminated. If your executor was launched with the `-m` argument or is push based, then it is monitored. Default is 30 minutes.             | No                        |
| `FIFTYONE_DO_LEGACY_EXPIRATION_MINUTES`      | The amount of time in minutes that an unmonitored delegated operation can run before being automatically terminated. Default is `FIFTYONE_DO_EXPIRATION_DAYS` converted to minutes. If this field is provided it will override `FIFTYONE_DO_EXPIRATION_DAYS`.                  | No                        |
| `FIFTYONE_ENCRYPTION_KEY`                    | Used to encrypt storage credentials in MongoDB                                                                                                                                                                                                                                 | Yes                       |
| `FIFTYONE_ENV`                               | GraphQL verbosity for the `fiftyone-teams-api` service; `production` will not log every GraphQL query, any other value will                                                                                                                                                    | No                        |
| `FIFTYONE_LOGGING_FORMAT`                    | The format to use for log messages; `json` or `text`. The default is `text`.                                                                                                                                                                                                   | No                        |
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
[docker-install]: https://docs.docker.com/engine/install/
[docker-compose-install]: https://docs.docker.com/compose/install/
[docker-hub]: https://hub.docker.com/
