<!-- markdownlint-disable-next-line first-line-heading no-inline-html -->
<div align="center">
<!-- markdownlint-disable-next-line no-inline-html -->
<p align="center">

<!-- markdownlint-disable-next-line no-inline-html line-length -->
<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<!-- markdownlint-disable-next-line no-inline-html line-length -->
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>

---

# Exposing the `teams-api` Service

There are two methods for SDK access to Fiftyone Teams

- Direct MongoDB connection
- FiftyOne Teams API

The database direct connection requires each user to have root database privileges.

The FiftyOne Teams API provides Role Based Access Control (RBAC) permissions.
By default, the API is not exposed.
To expose the FiftyOne Teams API, configure a
Kubernetes Ingress to route traffic to the Kubernetes
`teams-api` service on port 80 via the WebSocket protocol.

We use WebSockets to maintain connections and enable long-running process execution.
Before exposing the `teams-api` service,
validate that your infrastructure supports the WebSockets protocol.
(For example, you may need to replace AWS Classic Load Balancers (LB)
with AWS Application Load Balancers (ALB) for WebSocket support.)

To expose the `teams-api`` service, chose one of these two routing methods

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
            - path: /
              pathType: Prefix
              serviceName: teams-app
              servicePort: 80
        ```

1. Upgrade the deployment using the latest Helm chart

    ```shell
    helm repo update voxel51
    helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
      -f ./values.yaml
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
[API Connection](https://docs.voxel51.com/teams/api_connection.html).

## Validation

1. Verify the connectivity by accessing the FiftyOne Teams API's the health endpoint

    ```shell
    $ curl https://<DEPOY_URL>/health
    {"status":"available"}
    ```
