# Deploying FiftyOne Teams App using Docker Compose

The fiftyone-teams-app, fiftyone-teams-api, and fiftyone-app containers are avaialable via Docker Hub, with the appropriate credentials.  If you do not have Docker Hub credentials, please contact your support team for Docker Hub credentials.

## Initial Installation vs. Upgrades

`FIFTYONE_DATABASE_ADMIN` is set to `false` by default.  This is in order to make sure that upgrades do not break existing client installs.

- If you are performing a new install, consider setting `FIFTYONE_DATABASE_ADMIN` to `true`
- If you are performing an upgrade, please review our [Upgrade Process Recommendations](#upgrade-process-recommendations)

### Upgrade Process Recommendations

The FiftyOne Teams 0.10.0 Client (database version `0.18.0`) is backwards-compatible with the FiftyOne Teams 0.8.8 Database (version `0.16.6`) and all the versions between. Voxel51 recommends the following upgrade process:

1. Ensure all Python clients set `FIFTYONE_DATABASE_ADMIN=false` (this should generally be your default)
1. Upgrade FiftyOne Teams Python clients to FiftyOne Teams v0.10.0
1. Upgrade your FiftyOne Teams deploy to v1.0.0
1. Have an admin set `FIFTYONE_DATABASE_ADMIN=true` in their local Python client
1. Have the admin run `fiftyone migrate --all` to upgrade all datasets
1. Use `fiftyone migrate --info` to ensure that all datasets are now at version `0.18.0``

## Deploying the FiftyOne Teams App container

In a directory that contains the `docker-compose.yml` and `.env` files included in this directory, on a system with docker-compose installed, edit the `.env` file to set the |parameters required for this deployment.

| Variable                       | Purpose                                                                                                                                          |
|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| `AUTH0_API_CLIENT_ID`          | The Auth0 API Client ID from Voxel51                                                                                                             |
| `AUTH0_API_CLIENT_SECRET`      | The Auth0 API Client Secret from Voxel51                                                                                                         |
| `AUTH0_AUDIENCE`               | The Auth0 Audience from Voxel51                                                                                                                  |
| `AUTH0_CLIENT_ID`              | The Auth0 Application Client ID from Voxel51                                                                                                     |
| `AUTH0_CLIENT_SECRET`          | The Auth0 Application Client Secret from Voxel51                                                                                                 |
| `AUTH0_DOMAIN`                 | The Auth0 Domain from Voxel51                                                                                                                    |
| `AUTH0_ISSUER_BASE_URL`        | The Auth0 Issuer URL from Voxel51                                                                                                                |
| `AUTH0_ORGANIZATION`           | The Auth0 Organization from Voxel51                                                                                                              |
| `AUTH0_SECRET`                 | A random string used to encrypt cookies; use something like `openssl rand -hex 32` to generate this string                                       |
| `AUTH0_BASE_URL`               | The URL where you plan to deploy your FiftyOne Teams application                                                                                 |
| `API_BIND_ADDRESS`             | The host address that `fiftyone-teams-api` should bind to; `127.0.0.1` is appropriate for this in most cases                                     |
| `API_BIND_PORT`                | The host port that `fiftyone-teams-api` should bind to; the default is `8000`                                                                    |
| `API_URL`                      | The URL that `fiftyone-teams-app` should use to communicate with `fiftyone-teams-api`; `teams-api` is the compose service name                   |
| `FIFTYONE_ENV`                 | GraphQL verbosity for the `fiftyone-teams-api` service; `production` will not log every GraphQL qury, any other value will                       |
| `GRAPHQL_DEFAULT_LIMIT`        | Default GraphQL limit for results                                                                                                                |
| `FIFTYONE_BASE_DIR`            | This will be mounted as `/fiftyone` in the `fiftyone-teams-app` container and can be used to pass cloud storage credentials into the environment |
| `FIFTYONE_DEFAULT_APP_ADDRESS` | The host address that `fiftyone-app` should bind to; `127.0.0.1` is appropriate for this in most cases                                           |
| `FIFTYONE_DEFAULT_APP_PORT`    | The host port that `fiftyone-app` should bind to; the default is `5151`                                                                          |
| `APP_USE_HTTPS`                | Set this to true if your Application endpoint uses TLS; this should be 'true` in most cases'                                                     |
| `APP_BIND_ADDRESS`             | The host address that `fiftyone-teams-app` should bind to; this should be an externally-facing IP in most cases                                  |
| `APP_BIND_PORT`                | The host port that `fiftyone-teams-app` should bind tothe default is `3000`                                                                      |
| `FIFTYONE_TEAMS_PROXY_URL`     | The URL that `fiftyone-teams-app` will use to proxy requests to `fiftyone-app`                                                                   |


You will need to edit the `docker-compose.yml` if you want to include cloud storage credentials for accessing samples; examples are included in the `docker-compose.yml` file.

In the same directory, run the following command:

`docker-compose up -d`

The FiftyOne Teams App is now exposed on port 5151; an SSL endpoint (Load Balancer or Nginx Proxy or something similar) will need to be configured to route traffic from the SSL endpoint to port 5151 on the host running the FiftyOne Teams App.
