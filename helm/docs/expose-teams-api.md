<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>

---

# Exposing the `teams-api` Service

You may wish to expose your FiftyOne Teams API for SDK access.

You may expose your `teams-api` service in any manner that suits your deployment strategy.
The following is one solution, but does not represent the entirety of possible solutions.
Any solution allowing the FiftyOne Teams SDK to use websockets to access the `teams-api` service on port 80 should work.

**NOTE**: The `teams-api` service uses websockets to maintain connections and allow for long-running processes to complete.
Please ensure your Infrastructure supports websockets before attempting to expose the `teams-api` service.
(e.g. You will have to migrate from AWS Classic Load Balancers to AWS Application Load Balancers to provide websockets support.)

**NOTE**: If you are using file-based storage credentials, or setting environment variables, the same credentials must be shared with the `fiftyone-app` and `teams-api` pods.
Voxel51 recommends the use of Database Cloud Storage Credentials, which can be configured at `/settings/cloud_storage_credentials`.

## Adding a Second Host to the Ingress Controller (Host-Based Routing)

1. Set `apiSettings.dnsName` to the hostname to route API requests to
   (e.g. demo-api.fiftyone.ai)
1. Upgrade the deployment using the latest Helm chart

    ```shell
    helm repo update voxel51
    helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
      -f ./values.yaml
    ```

Modifications to the `teams-api` paths should be done using `ingress.teamsApi`.

## Use `ingress.paths` at the Ingress Controller (path-based routing)

Configure path-based routing for your ingress to route API paths to the `teams-api` service.

Depending on your ingress controller, a configuration for path-based routing may look like:

```yaml
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
