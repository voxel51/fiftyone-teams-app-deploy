<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>

---

# Deploying FiftyOne Teams App using Docker Compose

The `fiftyone-teams-app`, `fiftyone-teams-api`, and `fiftyone-app` images are available via Docker Hub, with the appropriate credentials. If you do not have Docker Hub credentials for the `voxel51` repositories, please contact your support team for Docker Hub credentials.

---

## Initial Installation vs. Upgrades

`FIFTYONE_DATABASE_ADMIN` is set to `false` by default for FiftyOne Teams version 1.3.0 installations and upgrades. This is because FiftyOne Teams version 1.3.0 is backwards compatible with FiftyOne Teams database schema version 0.19 (Teams Version 1.1) and newer.

- If you are performing an initial install, you will either want to connect to your MongoDB database with the 0.13.0 SDK before performing the FiftyOne Teams installation, or you will want to set `FIFTYONE_DATABASE_ADMIN: true` in the `environment` section of the `fiftyone-app` service definition.

- If you are performing an upgrade, please review our [Upgrade Process Recommendations](#upgrade-process-recommendations)

---

### FiftyOne Teams Upgrade Notes

#### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

As of FiftyOne Teams 1.1, containers based on the `fiftyone-teams-api` and `fiftyone-app` images now _REQUIRE_ the inclusion of the `FIFTYONE_ENCRYPTION_KEY` variable. This key is used to encrypt storage credentials in the MongoDB database.

The `FIFTYONE_ENCRYPTION_KEY` can be generated using the following python:

```
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
```

Voxel51 does not have access to this encryption key and cannot reproduce it. If this key is lost you will need to schedule an outage window to drop the storage credentials collection, replace the encryption key, and add the storage credentials via the UI again. Voxel51 strongly recommends storing this key in a safe place.

Storage credentials no longer need to be mounted into containers with appropriate environment variables being set; users with `Admin` permissions can use `/settings/cloud_storage_credentials` in the Web UI to add supported storage credentials.

FiftyOne Teams version 1.3 continues to support the use of environment variables to set storage credentials in the application context but is providing an alternate configuration path for future functionality.

#### Environment Proxies

FiftyOne Teams version 1.1 and higher support routing traffic through proxy servers; this can be configured by setting the following environment variables on all containers in the environment:

```
http_proxy: ${HTTP_PROXY_URL}
https_proxy: ${HTTPS_PROXY_URL}
no_proxy: ${NO_PROXY_LIST}
HTTP_PROXY: ${HTTP_PROXY_URL}
HTTPS_PROXY: ${HTTPS_PROXY_URL}
NO_PROXY: ${NO_PROXY_LIST}
```

You must also set the following environment variables on containers based on the `fiftyone-teams-app` image:

```
GLOBAL_AGENT_HTTP_PROXY: ${HTTP_PROXY_URL}
GLOBAL_AGENT_HTTPS_PROXY: ${HTTPS_PROXY_URL}
GLOBAL_AGENT_NO_PROXY: ${NO_PROXY_LIST}
```

The `NO_PROXY_LIST` value must include the names of the compose services to allow FiftyOne Teams services to talk to each other without going through a proxy server. By default these service names are `teams-api`, `teams-app`, `fiftyone-app`.

Examples of these settings are included in the FiftyOne Teams version 1.1.1 [compose.yaml](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/compose.yaml) and [.env](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/.env) files.

By default the Global Agent Proxy will log all outbound connections and identify which connections are routed through the proxy. You can reduce the verbosity of the logging output by adding the following environment variable to your `teamsAppSettings.env`:

```
ROARR_LOG: false
```

#### Text Similarity

FiftyOne Teams version 1.2 and higher supports using text similarity searches for images that are indexed with a model that [supports text queries](https://docs.voxel51.com/user_guide/brain.html#brain-similarity-text). If you choose to make use of this feature, you must use the `fiftyone-app-torch` image provided by Voxel51 instead of the `fiftyone-app` image, or build your own base image including torch.

Voxel51 recommends using a `compose.override.yaml` to [override the image selection](https://docs.docker.com/compose/extends/); this will allow you to update your `compose.yaml` in future releases without having to port this change forward. An example `compose.override.yaml` for this situation might look like:

```
version: '3.8'
services:
  fiftyone-app:
    image: voxel51/fiftyone-app-torch:v1.3.0
```

## Upgrade Process Recommendations

### Upgrade Process Recommendations From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this upgrade. You will need to either create a new IdP or modify your existing configuration in order to migrate to a new Auth0 Tenant.

### Upgrade Process Recommendations From Before FiftyOne Teams Version 1.1.0

The FiftyOne 0.13.0 SDK (database version 0.21.0) is _NOT_ backwards-compatible with FiftyOne Teams Database Versions prior to 0.19.0, and the FiftyOne 0.10.x SDK is not forwards compatible with current FiftyOne Teams Database Versions. If you are using a FiftyOne SDK older than 0.11.0, upgrading the Web server will require upgrading all FiftyOne SDK installations.

Voxel51 recommends the following upgrade process for upgrading from versions prior to FiftyOne Teams version 1.1.0:

1. Make sure your installation includes the required [FIFTYONE_ENCRYPTION_KEY](#fiftyone-teams-upgrade-notes) environment variable
1. If you are using a proxy server, make sure you have configured the appropriate [proxy environment variables](#environment-proxies)

1. [Upgrade to FiftyOne Teams version 1.3.0](#deploying-fiftyone-teams) with `FIFTYONE_DATABASE_ADMIN=true` (this is not the default in the `config.yaml` for this release).<br>
   **NOTE:** FiftyOne SDK users will lose access to the FiftyOne Teams Database at this step until they upgrade to `fiftyone==0.13.0`
1. Upgrade your FiftyOne SDKs to version 0.13.0<br>
   The command line for installing the FiftyOne SDK associated with your FiftyOne Teams version is available in the FiftyOne Teams UI under `Account > Install FiftyOne` after a user has logged in.
1. Use `fiftyone migrate --info` to make sure that all datasets have been migrated to version 0.21.0.
   - If not all datasets have been upgraded, an admin can run `FIFTYONE_DATABASE_ADMIN=true fiftyone migreat --all` in their local environment


### Upgrade Process Recommendations From FiftyOne Teams Version 1.1.0 and later

The FiftyOne 0.13.0 SDK (database version 0.21.0) is backwards-compatible with FiftyOne Teams Database Versions 0.19.0 and later, but the FiftyOne 0.11.0 SDK is _NOT_ forward compatible with FiftyOne Teams Database Version 0.21.0.

Voxel51 always recommends using the latest version of the FiftyOne SDK compatible with your FiftyOne Teams deployment.

Voxel51 recommends the following upgrade process for upgrading from FiftyOne Teams version 1.1.0 or later:

1. Ensure all FiftyOne SDK users set `FIFTYONE_DATABASE_ADMIN=false` or `unset FIFTYONE_DATABASE_ADMIN` (this should generally be your default)
1. [Upgrade to FiftyOne Teams version 1.3.0](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 0.13.0<br>
   The command line for installing the FiftyOne SDK associated with your FiftyOne Teams version is available in the FiftyOne Teams UI under `Account > Install FiftyOne` after a user has logged in.
1. Have the admin run `FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all` to upgrade all datasets<br>
   **NOTE** Any FiftyOne SDK less than 0.13.0 will lose database connectivity at this point; upgrading to `fiftyone==0.13.0` is required
1. Use `fiftyone migrate --info` to ensure that all datasets are now at version 0.21.0

---

## Deploying FiftyOne Teams

In a directory that contains the `compose.yml` and `.env` files included in this repository, on a system with docker-compose installed, edit the `.env` file to set the parameters required for this deployment ([see table below](#fiftyone-teams-environment-variables)).

In the same directory, run the following command:

`docker-compose up -d`

The FiftyOne Teams App is now exposed on port 3000; an SSL endpoint (Load Balancer or Nginx Proxy or something similar) will need to be configured to route traffic from the SSL endpoint to port 3000 on the host running the FiftyOne Teams App.

An example nginx site configuration that forwards http traffic to https, and https traffic for `your.server.name` to port 3000, [has been included in this repository](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/example-nginx-site.conf).

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
| `FIFTYONE_APP_THEME`                         | The default theme configuration for your FiftyOne Teams application:<br>&ensp;- `dark`: Application will default to dark theme when user visits for the first time<br>&ensp;- `light`: Application will default to light theme when user visits for the first time<br>&ensp;- `always-dark`: Application will default to dark theme on each refresh (even if user changes theme to light within the app)<br>&ensp;- `always-light`: Application will default to light theme on each refresh (even if user changes theme to dark within the app) | No       |
| `FIFTYONE_BASE_DIR`                          | This will be mounted as `/fiftyone` in the `fiftyone-teams-app` container and can be used to pass cloud storage credentials into the environment                                                                                                                                                                                                                                                                                                                                                                                                | No       |
| `FIFTYONE_DEFAULT_APP_ADDRESS`               | The host address that `fiftyone-app` should bind to; `127.0.0.1` is appropriate for this in most cases                                                                                                                                                                                                                                                                                                                                                                                                                                          | Yes      |
| `FIFTYONE_DEFAULT_APP_PORT`                  | The host port that `fiftyone-app` should bind to; the default is `5151`                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Yes      |
| `FIFTYONE_ENCRYPTION_KEY`                    | Used to encrypt storage credentials in MongoDB                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Yes      |
| `FIFTYONE_ENV`                               | GraphQL verbosity for the `fiftyone-teams-api` service; `production` will not log every GraphQL query, any other value will                                                                                                                                                                                                                                                                                                                                                                                                                     | No       |
| `FIFTYONE_PLUGINS_DIR`                       | Persistent directory for plugins to be stored in. `teams-api` must have write access to this directory, all plugin nodes must have read access to this directory.                                                                                                                                                                                                                                                                                                                                                                               | No       |
| `FIFTYONE_TEAMS_PROXY_URL`                   | The URL that `fiftyone-teams-app` will use to proxy requests to `fiftyone-app`                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | Yes      |
| `GRAPHQL_DEFAULT_LIMIT`                      | Default GraphQL limit for results                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | No       |
| `HTTP_PROXY_URL`                             | The URL for your environment http proxy                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | No       |
| `HTTPS_PROXY_URL`                            | The URL for your environment https proxy                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | No       |
| `NO_PROXY_LIST`                              | The list of servers that should bypass the proxy; if a proxy is in use this must include the list of FiftyOne services (`teams-api, teams-app, fiftyone-app`)                                                                                                                                                                                                                                                                                                                                                                                   | No       |
