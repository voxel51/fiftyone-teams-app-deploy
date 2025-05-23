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

# Exposing the `teams-api` Service

You may wish to expose your FiftyOne Teams API for SDK access.

You may expose your `teams-api` service in any
manner that suits your deployment strategy.
The following are two possible solutions but
do not represent the entirety of possible solutions.
Any solution allowing the FiftyOne Teams SDK to use websockets
to access the `teams-api` container on port 8000 should work.

**NOTE**: The `teams-api` service uses websockets to maintain connections
and allow for long-running processes to complete.
Please ensure your Infrastructure supports websockets
before attempting to expose the `teams-api` service.
(e.g. You will have to migrate from AWS Classic Load Balancers
to AWS Application Load Balancers to provide websockets support.)

**NOTE**: If you are using file-based storage credentials,
or setting environment variables, the same credentials must
be shared with the `fiftyone-app` and `teams-api` containers.
Voxel51 recommends the use of Database Cloud Storage Credentials,
which can be configured at `/settings/cloud_storage_credentials`.

## Expose `teams-api` Directly

**NOTE**: This method does not protect your API
endpoint with TLS and will send API Keys in clear text.
While it is the simplest mechanism, consider the
security implications before using this method.

1. Edit your `.env` file setting `API_BIND_ADDRESS` to `0.0.0.0`
1. Recreate your environment using the
   [plugin specific](./configuring-plugins.md)
   command

   ```shell
   docker compose
   ```

1. Access your FiftyOne Teams API using the same hostname
   as your FiftyOne Teams App using port 8000

## Expose `teams-api` using Nginx and a unique hostname

1. Create a second hostname for your API (e.g. demo-api.fiftyone.ai)
   and create SSL certificates for that hostname
1. Using
   [example-nginx-api.conf](../example-nginx-api.conf)
   as a template, create a second service for your nginx reverse proxy
1. Reload your nginx configuration
1. Access your FiftyOne Teams API using the new hostname

## Expose `teams-api` Using Path-Based Routing

1. Use
   [example-nginx-path-routing.conf](../example-nginx-path-routing.conf)
   as a template and configure additional `locations` for api-based routes
1. Reload your nginx configuration
1. Access your FiftyOne Teams API using the same hostname

## Advanced Configuration

The server has appropriate default settings for most deployments. However,
there are some server configurations that you may want to change with advice
from your Customer Success team, if you experience timeout or networking issues
when connecting through the exposed API server. Any of the below configurations
can be set in the `.env` file.

- `FIFTYONE_TEAMS_API_KEEP_ALIVE_TIMEOUT`: How long to hold a TCP connection
open (sec). Defaults to 120.
- `FIFTYONE_TEAMS_API_REQUEST_MAX_HEADER_SIZE`: How big a request header may be
(bytes). Defaults to 8192 bytes, max is 16384 bytes.
- `FIFTYONE_TEAMS_API_REQUEST_MAX_SIZE`: How big a request may be (bytes).
Defaults to 100 megabytes.
- `FIFTYONE_TEAMS_API_REQUEST_TIMEOUT`: How long a request can take to arrive
(sec). Defaults to 600.
- `FIFTYONE_TEAMS_API_RESPONSE_TIMEOUT`: How long a response can take to process
(sec). Defaults to 600.
- `FIFTYONE_TEAMS_API_WEBSOCKET_MAX_SIZE`: Maximum size for incoming websocket
messages (bytes). Defaults to 16 MiB.
- `FIFTYONE_TEAMS_API_WEBSOCKET_PING_TIMEOUT`: Connection is closed when Pong
is not received after ping_timeout seconds. Defaults to 600.
