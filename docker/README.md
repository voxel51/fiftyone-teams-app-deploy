# Deploying FiftyOne Teams App using Docker Compose

The `fiftyone-teams-app`, `fiftyone-teams-api`, and `fiftyone-app` images are avaialable via Docker Hub, with the appropriate credentials.  If you do not have Docker Hub credentials, please contact your support team for Docker Hub credentials.

## Initial Installation vs. Upgrades

`FIFTYONE_DATABASE_ADMIN` is set to `true` by default for FiftyOne Teams v1.1.0 installations and upgrades.  This is because FiftyOne Teams v1.1.0 is not backwards compatible with previous versions of the FiftyOne Teams database schema.

If you upgrade from previous versions of FiftyOne Teams your currently deployed FiftyOne Teams SDKs will no longer be able to connect to the database until you upgrade to `fiftyone` version `0.11.0`

- If you are performing an upgrade, please review our [Upgrade Process Recommendations](#upgrade-process-recommendations)


### v1.1.0 Upgrade Notes

#### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

Containers based on the `fiftyone-teams-api` and `fiftyone-app` images now _REQUIRE_ the inclusion of the `FIFTYONE_ENCRYPTION_KEY` variable.  This key is used to encrypt storage credentials in the MongoDB database.

The `FIFTYONE_ENCRYPTION_KEY` can be generated using the following python:

```
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
```

Voxel51 does not have access to this encryption key and cannot reproduce it.  If this key is lost you will need to schedule an outage window to drop the storage credentials collection, replace the encryption key, and add the storage credentials via the UI again.  Voxel51 strongly recommends storing this key in a safe place.

Storage credentials no longer need to be mounted into containers with appropriate environment variables being set; users with `Admin` permissions can use `/settings/cloud_storage_credentials` in the Web UI to add supported storage credentials.

FiftyOne Teams v1.1.0 continues to support the use of environment variables to set storage credentials in the application context but is providing an alternate configuration path for future functionality.

#### Environment Proxies

FiftyOne Teams now supports routing traffic through proxy servers; this can be configured by setting the following environment variables on all containers in the environment:

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

The `NO_PROXY_LIST` value must include the names of the compose services to allow FiftyOne Teams services to talk to each other without going through a proxy server.  By default these service names are `teams-api`, `teams-app`, `fiftyone-app`.

Examples of these settings are included in the `v1.1.0` [compose.yaml](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/compose.yaml) and [.env](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docker/.env) files.


### Upgrade Process Recommendations

The FiftyOne Teams 0.11.0 Client (database version `0.19.0`) is _NOT_ backwards-compatible with any FiftyOne Teams Database Version.  Upgrading the Web server will require upgrading `fiftyone` SDK versions. Voxel51 recommends the following upgrade process:

1. Upgrade to FiftyOne Teams v1.1.0 with `FIFTYONE_DATABASE_ADMIN=true` (this is the default in the `config.yaml` for this release).
1. Upgrade your `fiftyone` SDKs to version 0.11.0 (`pip install -U --index-url https://${TOKEN}@pypi.fiftyone.ai fiftyone==0.11.0`)
1. Use `fiftyone migrate --info` to ensure that all datasets are now at version `0.19.0`


## Deploying the FiftyOne Teams App container

In a directory that contains the `docker-compose.yml` and `.env` files included in this directory, on a system with docker-compose installed, edit the `.env` file to set the |parameters required for this deployment.

| Variable                       | Purpose                                                                                                                                                       | Required |
|--------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| `API_BIND_ADDRESS`             | The host address that `fiftyone-teams-api` should bind to; `127.0.0.1` is appropriate for this in most cases                                                  | Yes      |
| `API_BIND_PORT`                | The host port that `fiftyone-teams-api` should bind to; the default is `8000`                                                                                 | Yes      |
| `API_URL`                      | The URL that `fiftyone-teams-app` should use to communicate with `fiftyone-teams-api`; `teams-api` is the compose service name                                | Yes      |
| `APP_BIND_ADDRESS`             | The host address that `fiftyone-teams-app` should bind to; this should be an externally-facing IP in most cases                                               | Yes      |
| `APP_BIND_PORT`                | The host port that `fiftyone-teams-app` should bind tothe default is `3000`                                                                                   | Yes      |
| `APP_USE_HTTPS`                | Set this to true if your Application endpoint uses TLS; this should be 'true` in most cases'                                                                  | Yes      |
| `AUTH0_API_CLIENT_ID`          | The Auth0 API Client ID from Voxel51                                                                                                                          | Yes      |
| `AUTH0_API_CLIENT_SECRET`      | The Auth0 API Client Secret from Voxel51                                                                                                                      | Yes      |
| `AUTH0_AUDIENCE`               | The Auth0 Audience from Voxel51                                                                                                                               | Yes      |
| `AUTH0_BASE_URL`               | The URL where you plan to deploy your FiftyOne Teams application                                                                                              | Yes      |
| `AUTH0_CLIENT_ID`              | The Auth0 Application Client ID from Voxel51                                                                                                                  | Yes      |
| `AUTH0_CLIENT_SECRET`          | The Auth0 Application Client Secret from Voxel51                                                                                                              | Yes      |
| `AUTH0_DOMAIN`                 | The Auth0 Domain from Voxel51                                                                                                                                 | Yes      |
| `AUTH0_ISSUER_BASE_URL`        | The Auth0 Issuer URL from Voxel51                                                                                                                             | Yes      |
| `AUTH0_ORGANIZATION`           | The Auth0 Organization from Voxel51                                                                                                                           | Yes      |
| `AUTH0_SECRET`                 | A random string used to encrypt cookies; use something like `openssl rand -hex 32` to generate this string                                                    | Yes      |
| `FIFTYONE_BASE_DIR`            | This will be mounted as `/fiftyone` in the `fiftyone-teams-app` container and can be used to pass cloud storage credentials into the environment              | No       |
| `FIFTYONE_DEFAULT_APP_ADDRESS` | The host address that `fiftyone-app` should bind to; `127.0.0.1` is appropriate for this in most cases                                                        | Yes      |
| `FIFTYONE_DEFAULT_APP_PORT`    | The host port that `fiftyone-app` should bind to; the default is `5151`                                                                                       | Yes      |
| `FIFTYONE_ENCRYPTION_KEY`      | Used to encrypt storage credentials in MongoDB                                                                                                                | Yes      |
| `FIFTYONE_ENV`                 | GraphQL verbosity for the `fiftyone-teams-api` service; `production` will not log every GraphQL qury, any other value will                                    | No       |
| `FIFTYONE_TEAMS_PROXY_URL`     | The URL that `fiftyone-teams-app` will use to proxy requests to `fiftyone-app`                                                                                | Yes      |
| `GRAPHQL_DEFAULT_LIMIT`        | Default GraphQL limit for results                                                                                                                             | No       |
| `HTTPS_PROXY_URL`              | The URL for your enviornment https proxy                                                                                                                      | No       |
| `HTTP_PROXY_URL`               | The URL for your environment http proxy                                                                                                                       | No       |
| `NO_PROXY_LIST`                | The list of servers that should bypass the proxy; if a proxy is in use this must include the list of FiftyOne services (`teams-api, teams-app, fiftyone-app`) | No       |


You will need to edit the `docker-compose.yml` if you want to include cloud storage credentials for accessing samples; examples are included in the `docker-compose.yml` file.

In the same directory, run the following command:

`docker-compose up -d`

The FiftyOne Teams App is now exposed on port 3000; an SSL endpoint (Load Balancer or Nginx Proxy or something similar) will need to be configured to route traffic from the SSL endpoint to port 3000 on the host running the FiftyOne Teams App.

An example nginx site configuration that forwards http traffic to https, and https traffic for `your.server.name` to port 3000 might look like:

```
upstream teams-app {
  server localhost:3000;
}

server {
  server_name your.server.name;

  proxy_busy_buffers_size   512k;
  proxy_buffers   4 512k;
  proxy_buffer_size   256k;

  location / {
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_pass http://teams-app;
  }

    listen 443 ssl;
    ssl_certificate /path/to/your/certificate.pem;
    ssl_certificate_key /path/to/your/key.pem;
    ssl_dhparam /path/to/your/dhparams.pem;
}

server {
    if ($host = your.server.name) {
        return 301 https://$host$request_uri;
    }

  listen 80;
  server_name your.server.name;
    return 404;
}
```
