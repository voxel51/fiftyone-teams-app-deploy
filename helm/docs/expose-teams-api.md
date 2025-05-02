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

# Exposing the Enterprise `teams-api` Service

There are two methods for SDK access to FiftyOne Enterprise

- Direct MongoDB connection
- FiftyOne Enterprise API

The database direct connection requires each user to have root database privileges.

The FiftyOne Enterprise API provides Role Based Access Control (RBAC) permissions.
By default, the API is not exposed.
To expose the FiftyOne Enterprise API, configure a
Kubernetes Ingress to route traffic to the Kubernetes
`teams-api` service on port 80 via the WebSocket protocol.

We use WebSockets to maintain connections and enable long-running process execution.
Before exposing the `teams-api` service,
validate that your infrastructure supports the WebSockets protocol.
(For example, you may need to replace AWS Classic Load Balancers (LB)
with AWS Application Load Balancers (ALB) for WebSocket support.)

To expose the `teams-api` service, chose one of these two routing methods

- Host-Based
- Path-based

## Host-Based Routing

Add a Second Host to the Ingress Controller

1. Create or update TLS certificate for the new host by either
    1. Obtaining a new certificate for the new host
    1. Updating an existing certificate by adding the new
       host to the list of Subject Alternative Names (SAN)
1. Add a new DNS entry for the new host to route to the Ingress
1. Update `values.yaml`
    1. Set `apiSettings.dnsName` to the hostname to route API requests to
      (e.g. `demo-api.fiftyone.ai`)
1. Upgrade the deployment using the latest Helm chart

    ```shell
    helm repo update voxel51
    helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
      -f ./values.yaml
    ```

## Path-Based Routing

Path based routing doesn't require additional DNS entries and TLS certificates.
In `values.yaml`, remove `apiSettings.dnsName`.
This routes traffic to the API paths to the `teams-api` service.

Every Ingress Controller implementation is different.
Consult your ingress controller documentation.

To use this chart's ingress object

1. Update `values.yaml`
    1. Set `ingress.enabled` to `true`
    1. Set the other ingress values
    1. Configure the Ingress paths to include

        ```yaml
        # values.yaml
        ingress:
          paths:
            - path: /_pymongo
              pathType: Prefix
              serviceName: teams-api
              servicePort: 80
            - path: /health
              pathType: Prefix
              serviceName: teams-api
              servicePort: 80
            - path: /graphql/v1
              pathType: Prefix
              serviceName: teams-api
              servicePort: 80
            - path: /file
              pathType: Prefix
              serviceName: teams-api
              servicePort: 80
            - path: /cas
              pathType: Prefix
              serviceName: teams-cas
              servicePort: 80
        ```

1. Upgrade the deployment using the latest Helm chart

    ```shell
    helm repo update voxel51
    helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
      -f ./values.yaml
    ```

## Note For NGINX IngressClass Users

> [!NOTE]
> Voxel51 is not affiliated with Nginx and you should reference the
> [nginx documentation][nginx-docs] for advanced configuration.

The FiftyOne Enterprise API utilizes websockets for client/server communication
on a variety of methods.
If you are using an `nginx` ingress class for your ingress controller, it is
possible that extra annotations are required for the HTTPS to WSS upgrade to
happen.

If you are experiencing issues when connecting to the FiftyOne Enterprise API
from the SDK, Voxel51 has seen success with the following annotations:

```yaml
# values.yaml
ingress:
  annotations:
    nginx.org/proxy-read-timeout: "3600"
    nginx.org/proxy-send-timeout: "3600"
    nginx.org/websocket-services: teams-api
```

## Configure your SDK

1. In `~/.fiftyone/config.json`, set

    ```json
    {
      "api_uri": "https://<DEPOY_URL>",
      "api_key": "<REDACTED>"
    }
    ```

For more information, see
[API Connection](https://docs.voxel51.com/enterprise/api_connection.html).

## Validation

1. Verify the connectivity by accessing the FiftyOne Enterprise API's health endpoint

    ```shell
    $ curl https://<DEPOY_URL>/health
    {"status":"available"}
    ```

## Advanced Configuration

The server has appropriate default settings for most deployments. However,
there are some server configurations that you may want to change with advice
from your Customer Success team, if you experience timeout or networking issues
when connecting through the exposed API server. Any of the below configurations
can be set in the `values.yaml` file under the `apiSettings` section.

```yaml
apiSettings:
  env:
    # -- How long to hold a TCP connection open (sec). Defaults to 120.
    FIFTYONE_TEAMS_API_KEEP_ALIVE_TIMEOUT: 120

    # -- How big a request header may be (bytes). Defaults to 8192 bytes, max
    # is 16384 bytes.
    FIFTYONE_TEAMS_API_REQUEST_MAX_HEADER_SIZE: 8192

    # -- How big a request may be (bytes). Defaults to 100 megabytes.
    FIFTYONE_TEAMS_API_REQUEST_MAX_SIZE: 100000000

    # -- How long a request can take to arrive (sec). Defaults to 600.
    FIFTYONE_TEAMS_API_REQUEST_TIMEOUT: 600

    # -- How long a response can take to process (sec). Defaults to 600.
    FIFTYONE_TEAMS_API_RESPONSE_TIMEOUT: 600

    # -- Maximum size for incoming websocket messages (bytes). Defaults to 16 MiB.
    FIFTYONE_TEAMS_API_WEBSOCKET_MAX_SIZE: 16777216

    # -- Connection is closed when Pong is not received after ping_timeout seconds.
    # Defaults to 600.
    FIFTYONE_TEAMS_API_WEBSOCKET_PING_TIMEOUT: 600
```

<!-- Reference links -->
[nginx-docs]: https://docs.nginx.com/nginx-ingress-controller/configuration/ingress-resources/advanced-configuration-with-annotations/
