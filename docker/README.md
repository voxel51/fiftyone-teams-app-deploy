# Deploying FiftyOne Teams App using a Dockerfile and Docker Compose

To build this container image you will require an authentication token from Voxel51.  If you are already a Voxel51 customer please contact your support team to obtain a token, otherwise please contact [Voxel51](https://voxel51.com/#teams-form) if you would like more information regarding FiftyOne Teams.

To deploy this image you will require an Organization ID, and a Client ID provided by Voxel51.  If you are already a Voxel51 customer please contact your support team to obtain an Organization ID and Client ID, otherwise please contact [Voxel51](https://voxel51.com/#teams-form) if you would like more information regarding FiftyOne Teams.

The fiftyone-teams-app container is avaialable via Docker Hub, with the appropriate credentials.  If you would like to use the Voxel51-built container image, please contact your support team for Docker Hub credentials.

## Building the FiftyOne Teams App image

In a directory that contains the `Dockerfile` included in this repository, on a system with docker installed, run the following command:

`docker build --no-cache --build-arg TOKEN=${TOKEN} -t voxel51/fiftyone-teams-app:v0.3.1 .`

## Initial Installation vs. Upgrades

`FIFTYONE_DATABASE_ADMIN` is set to `false` by default.  This is in order to make sure that upgrades do not break existing client installs.

- If you are performing a new install, consider setting `FIFTYONE_DATABASE_ADMIN` to `true`
- If you are performing an upgrade, please review our [Upgrade Process Recommendations](#upgrade-process-recommendations)

### Upgrade Process Recommendations

The FiftyOne Teams 0.10.0 Client (database version `0.18.0`) is backwards-compatible with the FiftyOne Teams 0.8.8 Database (version `0.16.6`) and all the versions between. Voxel51 recommends the following upgrade process:

1. Ensure all Python clients set `FIFTYONE_DATABASE_ADMIN=false` (this should generally be your default)
1. Upgrade FiftyOne Teams Python clients to FiftyOne Teams v0.10.0
1. Upgrade your FiftyOne Teams App deploy to version v0.3.1
1. Have an admin set `FIFTYONE_DATABASE_ADMIN=true` in their local Python client
1. Have the admin run `fiftyone migrate --all` to upgrade all datasets
1. Use `fiftyone migrate --info` to ensure that all datasets are now at version `0.18.0``

## Deploying the FiftyOne Teams App container

In a directory that contains the `docker-compose.yml` and `.env` files included in this directory, on a system with docker-compose installed, edit the `.env` file to set the four parameters required for this deployment.

| Variable                    | Purpose                                                                                                                                          |
|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| FIFTYONE_BASE_DIR           | This will be mounted as `/fiftyone` in the `fiftyone-teams-app` container and can be used to pass cloud storage credentials into the environment |
| FIFTYONE_DB_DIR             | This will be mounted as `/data/db` in the `db` container and is used to store MongoDB data files                                                 |
| FIFTYONE_DB_PASSWORD        | This will set the root user password and add it to the MongoDB connection string                                                                 |
| FIFTYONE_DB_USERNAME        | This will set the root user username and add it to the MongoDB connection string                                                                 |
| FIFTYONE_TEAMS_CLIENT_ID    | This Auth0 Client ID is provided by Voxel51                                                                                                      |
| FIFTYONE_TEAMS_ORGANIZATION | This Auth0 Organization ID is provided by Voxel51                                                                                                |

You will need to edit the `docker-compose.yml` if you want to include cloud storage credentials for accessing samples; examples are included in the `docker-compose.yml` file.

In the same directory, run the following command:

`docker-compose up -d`

The FiftyOne Teams App is now exposed on port 5151; an SSL endpoint (Load Balancer or Nginx Proxy or something similar) will need to be configured to route traffic from the SSL endpoint to port 5151 on the host running the FiftyOne Teams App.
