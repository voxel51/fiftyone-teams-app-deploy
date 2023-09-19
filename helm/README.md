<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>

---

FiftyOne Teams is the enterprise version of the open source
[FiftyOne](https://github.com/voxel51/fiftyone)
project.

Please contact
[Voxel51](https://voxel51.com/#teams-form)
for more information regarding Fiftyone Teams.

# Deploying FiftyOne Teams App Using Helm

We publish container images to these Docker Hub repositories

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch`
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`

For Docker Hub credentials, please contact your Voxel51 support team.

---

## Initial Installation vs. Upgrades

By default, `FIFTYONE_DATABASE_ADMIN` is set to `false` for FiftyOne Teams version 1.4.0.

- When performing an initial installation, in `values.yaml`, set
  `appSettings.env.FIFTYONE_DATABASE_ADMIN: true`
- When performing an upgrade, please review our
  [Upgrade Process Recommendations](#upgrade-process-recommendations)

---

## Notes and Considerations

While not all parameters are required, Voxel51 frequently sees deployments use the following parameters:

- `imagePullSecrets`
- `ingress.annotations`

Please consider if you will require these settings for your deployment.

---

### FiftyOne Teams Upgrade Notes

#### Enabling FiftyOne Teams Authenticated API

FiftyOne Teams v1.3 introduces the capability to connect FiftyOne Teams SDKs through the FiftyOne Teams API (instead of creating a direct connection to MongoDB).

To enable the FiftyOne Teams Authenticated API you will need to
[expose the FiftyOne Teams API endpoint](docs/expose-teams-api.md)
and
[configure your SDK](https://docs.voxel51.com/teams/api_connection.html).

#### Enabling FiftyOne Teams Plugins

FiftyOne Teams v1.3+ includes significant enhancements for
[Plugins](https://docs.voxel51.com/plugins/index.html)
to customize and extend the functionality of FiftyOne Teams in your environment.

There are three modes for plugins

1. Builtin Plugins Only
    - No changes are required for this mode
1. Plugins run in the `fiftyone-app` deployment
    - To enable this mode
        - In `values.yaml`, set the path for a Persistent Volume Claim mounted to the `teams-api` and `fiftyone-app` deployments in both
            - `appSettings.env.FIFTYONE_PLUGINS_DIR`
            - `apiSettings.env.FIFTYONE_PLUGINS_DIR`
        - Mount a [Persistent Volume Claim](docs/plugins-storage.md) that provides
            - `ReadWrite` permissions to the `teams-api` deployment
              at the `FIFTYONE_PLUGINS_DIR` path
            - `ReadOnly` permission to the `fiftyone-app` deployment
              at the `FIFTYONE_PLUGINS_DIR` path
1. Plugins run in a dedicated `teams-plugins` deployment
    - To enable this mode
        - In `values.yaml`, set
            - `pluginsSettings.enabled: true`
            - The path for a Persistent Volume Claim mounted to the `teams-api` and `teams-plugins` deployments in both
                - `pluginsSettings.env.FIFTYONE_PLUGINS_DIR`
                - `apiSettings.env.FIFTYONE_PLUGINS_DIR`
        - Mount a [Persistent Volume Claim](docs/plugins-storage.md) that provides
            - `ReadWrite` permissions to the `teams-api` deployment
              at the `FIFTYONE_PLUGINS_DIR` path
            - `ReadOnly` permission to the `teams-plugins` deployment
              at the `FIFTYONE_PLUGINS_DIR` path

Deploy plugins using the FiftyOne Teams UI at `/settings/plugins`.
Any early-adopter plugins installed via manual methods must be redeployed using the FiftyOne Teams UI.

#### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

Pods based on the `fiftyone-teams-api` and `fiftyone-app` images must include the  `FIFTYONE_ENCRYPTION_KEY` variable.
This key is used to encrypt storage credentials in the MongoDB database.

The generate the `encryptionKey`, run this Python code

```python
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
```

Voxel51 does not have access to this encryption key and cannot reproduce it.
If the key is lost, you will need to

1. Schedule an outage window
    1. Drop the storage credentials collection
    1. Replace the encryption key
    1. Add the storage credentials via the UI again.

Voxel51 strongly recommends storing this key in a safe place.

Storage credentials no longer need to be mounted into pods with appropriate environment variables being set.
Users with `Admin` permissions may add supported storage credentials using `/settings/cloud_storage_credentials` in the Web UI.

FiftyOne Teams continues to support the use of environment variables to set storage credentials in the application context but is providing an alternate configuration path for future functionality.

#### Environment Proxies

FiftyOne Teams supports routing traffic through proxy servers.
To configure this, set the following environment variables on

1. All pods in the environment (`*.env`):

    ```yaml
    http_proxy: http://proxy.yourcompany.tld:3128
    https_proxy: https://proxy.yourcompany.tld:3128
    no_proxy: <apiSettings.service.name>, <appSettings.service.name>, <teamsAppSettings.service.name>
    HTTP_PROXY: http://proxy.yourcompany.tld:3128
    HTTPS_PROXY: https://proxy.yourcompany.tld:3128
    NO_PROXY: <apiSettings.service.name>, <appSettings.service.name>, <teamsAppSettings.service.name>
    ```

1. Pods based on the `fiftyone-teams-app` image (`teamsAppSettings.env`)

    ```yaml
    GLOBAL_AGENT_HTTP_PROXY: http://proxy.yourcompany.tld:3128
    GLOBAL_AGENT_HTTPS_PROXY: https://proxy.yourconpay.tld:3128
    GLOBAL_AGENT_NO_PROXY: <apiSettings.service.name>, <appSettings.service.name>, <teamsAppSettings.service.name>
    ```

The `NO_PROXY` and `GLOBAL_AGENT_NO_PROXY` values must include the Kubernetes service names that may communicate without going through a proxy server.
By default these service names are

- `teams-api`
- `teams-app`
- `fiftyone-app`

If the service names were overridden in `*.service.name`, use these values instead.

By default the Global Agent Proxy will log all outbound connections and identify which connections are routed through the proxy.
To reduce the logging verbosity, add this environment variable to your `teamsAppSettings.env`

```ini
ROARR_LOG: false
```

#### Text Similarity

FiftyOne Teams version 1.2 and higher supports using text similarity searches for images that are indexed with a model that
[supports text queries](https://docs.voxel51.com/user_guide/brain.html#brain-similarity-text).
To use this feature, use a container image containing `torch` (PyTorch) instead of the `fiftyone-app` image.
Use the Voxel51 provided image `fiftyone-app-torch` or build your own base image including `torch`.

To override the default image, add a new `appSettings.image.repository` stanza to the Helm Chart.
Using the included `values.yaml` this configuration might look like:

```yaml
appSettings:
  image:
    repository: voxel51/fiftyone-app-torch
```

---

## Required Helm Chart Values

| Required Values                           | Default | Description                                 |
| ----------------------------------------- | ------- | ------------------------------------------- |
| `secret.fiftyone.apiClientId`             | None    | Voxel51-provided Auth0 API Client ID        |
| `secret.fiftyone.apiClientSecret`         | None    | Voxel51-provided Auth0 API Client Secret    |
| `secret.fiftyone.auth0Domain`             | None    | Voxel51-provided Auth0 Domain               |
| `secret.fiftyone.clientId`                | None    | Voxel51-provided Auth0 Client ID            |
| `secret.fiftyone.cookieSecret`            | None    | Random string for cookie encryption         |
| `secret.fiftyone.encryptionKey`           | None    | Encryption key for storage credentials      |
| `secret.fiftyone.mongodbConnectionString` | None    | MongoDB Connnection String                  |
| `secret.fiftyone.organizationId`          | None    | Voxel51-provided Auth0 Organization ID      |
| `teamsAppSettings.dnsName`                | None    | DNS Name for the FiftyOne Teams App Service |

## Optional Helm Chart Values

You can find a full `values.yaml` with all of the optional values [here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/fiftyone-teams-app/values.yaml)

| Optional Values                                                   | Default                    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
|-------------------------------------------------------------------|----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `apiSettings.affinity`                                            | None                       | FiftyOne Teams API service affinity rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `apiSettings.env`                                                 | defined below              | Arbitrary environment variables to pass to the FiftyOne Teams API pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `apiSettings.env.FIFTYONE_ENV`                                    | production                 | Verbosity for GraphQL query output                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `apiSettings.env.GRAPHQL_DEFAULT_LIMIT`                           | 10                         | Default GraphQL limit for results                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `apiSettings.env.LOGGING_LEVEL`                                   | INFO                       | Logging Verbosity                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `apiSettings.image.repository`                                    | voxel51/fiftyone-teams-api | Container Image for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `apiSettings.image.tag`                                           | Helm Chart Version         | Container Image Tag for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `apiSettings.nodeSelector`                                        | None                       | FiftyOne Teams API pod node selector rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `apiSettings.podAnnotations`                                      | None                       | FiftyOne Teams API pod annotation rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `apiSettings.podSecurityContext`                                  | None                       | FiftyOne Teams API pod security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `apiSettings.resources.limits.cpu`                                | None                       | CPU resource limits for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `apiSettings.resources.limits.memory`                             | None                       | Memory resource limits for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `apiSettings.resources.requests.cpu`                              | None                       | CPU resource requests for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `apiSettings.resources.requests.memory`                           | None                       | Memory resource requests for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `apiSettings.securityContext`                                     | None                       | FiftyOne Teams API service security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `apiSettings.service.annotations`                                 | None                       | FiftyOne Teams Service Annotations                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `apiSettings.service.containerPort`                               | 8000                       | Port for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `apiSettings.service.liveness.initialDelaySeconds`                | 45                         | Delay before Liveness checks for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `apiSettings.service.name`                                        | teams-api                  | FiftyOne Teams API service name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `apiSettings.service.nodePort`                                    | None                       | `nodePort` for Service Type `NodePort`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `apiSettings.service.port`                                        | 80                         | FiftyOne Teams API service port                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `apiSettings.service.readiness.initialDelaySeconds`               | 45                         | Delay before Readiness checks for FiftyOne Teams API pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `apiSettings.service.shortname`                                   | teams-api                  | Short name for port definitions (less than 15 characters)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `apiSettings.service.type`                                        | ClusterIP                  | FiftyOne Teams API service type                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `apiSettings.tolerations`                                         | None                       | FiftyOne Teams API service toleration rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `apiSettings.volumeMounts`                                        | None                       | FiftyOne Teams API pod volume mount definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `apiSettings.volumes`                                             | None                       | FiftyOne Teams API pod volume definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.affinity`                                            | None                       | FiftyOne App service affinity rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `appSettings.autoscaling.enabled`                                 | false                      | Enable Horizontal Autoscaling for the FiftyOne App Pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `appSettings.autoscaling.maxReplicas`                             | 20                         | Maximum Replicas for Horizontal Autoscaling in the FiftyOne App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `appSettings.autoscaling.minReplicas`                             | 2                          | Minimum Replicas for Horizontal Autoscaling in the FiftyOne App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `appSettings.autoscaling.targetCPUUtilizationPercentage`          | 80                         | Percent CPU Utilization for autoscaling the FiftyOne App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `appSettings.autoscaling.targetMemoryUtilizationPercentage`       | 80                         | Percent Memory Utilization for autoscaling the FiftyOne App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `appSettings.env`                                                 | defined below              | Arbitrary environment variables to pass to the FiftyOne App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `appSettings.env.FIFTYONE_DATABASE_ADMIN`                         | false                      | Toggles MongoDB database admin privileges for the FiftyOne App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `appSettings.env.FIFTYONE_MEDIA_CACHE_IMAGES`                     | false                      | Toggle image caching for the local FiftyOne App processes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.env.FIFTYONE_MEDIA_CACHE_SIZE_BYTES`                 | -1 (disabled)              | Set the media cache size for the local FiftyOne App processes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `appSettings.image.repository`                                    | voxel51/fiftyone-app       | Container Image for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `appSettings.image.tag`                                           | Helm Chart Version         | Container Image tag for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.nodeSelector`                                        | None                       | FiftyOne App pod node selector rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `appSettings.podAnnotations`                                      | None                       | FiftyOne App pod annotation rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `appSettings.podSecurityContext`                                  | None                       | FiftyOne App pod security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `appSettings.replicaCount`                                        | 2                          | FiftyOne App replica count if autoscaling is disabled                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `appSettings.resources.limits.cpu`                                | None                       | CPU resource limits for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.resources.limits.memory`                             | None                       | Memory resource limits for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `appSettings.resources.requests.cpu`                              | None                       | CPU resource requests for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `appSettings.resources.requests.memory`                           | None                       | Memory resource requests for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `appSettings.securityContext`                                     | None                       | FiftyOne App service security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `appSettings.service.annotations`                                 | None                       | FiftyOne App service annotations                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `appSettings.service.containerPort`                               | 5151                       | Port for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `appSettings.service.liveness.initialDelaySeconds`                | 45                         | Delay before Liveness checks for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `appSettings.service.name`                                        | fiftyone-app               | FiftyOne App service name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.service.nodePort`                                    | None                       | `nodePort` for Service Type `NodePort`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `appSettings.service.port`                                        | 80                         | FiftyOne App service port                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.service.readiness.initialDelaySeconds`               | 45                         | Delay before Readiness checks for FiftyOne App pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `appSettings.service.shortname`                                   | fiftyone-app               | Short name for port definitions (less than 15 characters)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.service.type`                                        | ClusterIP                  | FiftyOne App service type                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.tolerations`                                         | None                       | FiftyOne App service toleration rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `appSettings.volumeMounts`                                        | None                       | FiftyOne App pod volume mount definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `appSettings.volumes`                                             | None                       | FiftyOne App pod volume definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `ingress.annotations`                                             | None                       | Ingress annotations (if required)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `ingress.api.path`                                                | `/*`                       | Set the ingress path for host-based API ingress routing<br>&ensp;Only used if `apiSettings.dnsName` is used                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `ingress.api.pathType`                                            | ImplementationSpecific     | Set the ingress pathType for host-based API ingress routing<br>&ensp;Only used if `apiSettings.dnsName` is used                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `ingress.className`                                               | ""                         | Ingress class name (if required)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `ingress.enabled`                                                 | true                       | Toggle enabling ingress                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `ingress.paths`                                                   | See Below                  | List of ingress `path` and `pathType`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `ingress.paths.path`                                              | `/*`                       | path to associate with the FiftyOne Teams App service                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `ingress.paths.path.pathType`                                     | ImplementationSpecific     | Ingress path type (`ImplementationSpecific`, `Exact`, `Prefix`)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `ingress.teamsApp.path`                                           | `/*`                       | Set the ingress path for FiftyOne Teams App host-based routing<br>&ensp;Only used if `ingress.paths` is not set.                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `ingress.teamsApp.pathType`                                       | ImplementationSpecific     | Set the ingress path for FiftyOne Teams App host-based routing<br>&ensp;Only used if `ingress.paths` is not set.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `ingress.tlsEnabled`                                              | true                       | Enable TLS for Ingress Controller                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `ingress.tlsSecretName`                                           | fiftyone-teams-tls-secret  | TLS Secret for certificate with all three DNS Names                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `namespace.name`                                                  | fiftyone-teams             | Kubernetes Namespace already created for FiftyOne Teams                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `pluginsSettings.affinity`                                        | None                       | FiftyOne Plugins service affinity rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `pluginsSettings.autoscaling.enabled`                             | false                      | Enable Horizontal Autoscaling for the FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `pluginsSettings.autoscaling.maxReplicas`                         | 20                         | Maximum Replicas for Horizontal Autoscaling in the FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `pluginsSettings.autoscaling.minReplicas`                         | 2                          | Minimum Replicas for Horizontal Autoscaling in the FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `pluginsSettings.autoscaling.targetCPUUtilizationPercentage`      | 80                         | Percent CPU Utilization for autoscaling the FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `pluginsSettings.autoscaling.targetMemoryUtilizationPercentage`   | 80                         | Percent Memory Utilization for autoscaling the FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `pluginsSettings.enabled`                                         | false                      | Enable Dedicated Plugins service for FiftyOne Teams                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `pluginsSettings.env`                                             | defined below              | Arbitrary environment variables to pass to the FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `pluginsSettings.env.FIFTYONE_MEDIA_CACHE_IMAGES`                 | false                      | Toggle image caching for the local FiftyOne Plugins processes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `pluginsSettings.env.FIFTYONE_MEDIA_CACHE_SIZE_BYTES`             | -1 (disabled)              | Set the media cache size for the local FiftyOne Plugins processes                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `pluginsSettings.image.repository`                                | voxel51/fiftyone-app       | Container Image for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `pluginsSettings.image.tag`                                       | Helm Chart Version         | Container Image tag for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `pluginsSettings.nodeSelector`                                    | None                       | FiftyOne Plugins pod node selector rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `pluginsSettings.podAnnotations`                                  | None                       | FiftyOne Plugins pod annotation rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `pluginsSettings.podSecurityContext`                              | None                       | FiftyOne Plugins pod security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `pluginsSettings.replicaCount`                                    | 2                          | FiftyOne Plugins replica count if autoscaling is disabled                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `pluginsSettings.resources.limits.cpu`                            | None                       | CPU resource limits for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `pluginsSettings.resources.limits.memory`                         | None                       | Memory resource limits for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `pluginsSettings.resources.requests.cpu`                          | None                       | CPU resource requests for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `pluginsSettings.resources.requests.memory`                       | None                       | Memory resource requests for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `pluginsSettings.securityContext`                                 | None                       | FiftyOne Plugins service security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `pluginsSettings.service.annotations`                             | None                       | FiftyOne Plugins service annotations                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `pluginsSettings.service.containerPort`                           | 5151                       | Port for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `pluginsSettings.service.liveness.initialDelaySeconds`            | 45                         | Delay before Liveness checks for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `pluginsSettings.service.name`                                    | teams-plugins              | FiftyOne Plugins service name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `pluginsSettings.service.nodePort`                                | None                       | `nodePort` for Service Type `NodePort`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `pluginsSettings.service.port`                                    | 80                         | FiftyOne Plugins service port                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `pluginsSettings.service.readiness.initialDelaySeconds`           | 45                         | Delay before Readiness checks for FiftyOne Plugins pods                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `pluginsSettings.service.shortname`                               | teams-plugins              | Short name for port definitions (less than 15 characters)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `pluginsSettings.service.type`                                    | ClusterIP                  | FiftyOne Plugins service type                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `pluginsSettings.tolerations`                                     | None                       | FiftyOne Plugins service toleration rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `pluginsSettings.volumeMounts`                                    | None                       | FiftyOne Plugins pod volume mount definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `pluginsSettings.volumes`                                         | None                       | FiftyOne Plugins pod volume definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `secret.create`                                                   | true                       | Toggle creation of the FiftyOne secret by Helm                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `secret.name`                                                     | fiftyone-teams-secrets     | Name for the FiftyOne Teams configuration secrets                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `secret.fiftyone.fiftyoneDatabaseName`                            | fiftyone                   | MongoDB Database Name for FiftyOne Teams                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `serviceAccount.annotations`                                      | None                       | Service account annotations                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `serviceAccount.create`                                           | true                       | Toggle creation of a service account for the FiftyOne Teams deployment                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `serviceAccount.name`                                             | fiftyone-teams             | Service account name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `teamsAppSettings.affinity`                                       | None                       | FiftyOne Teams App service affinity rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `teamsAppSettings.autoscaling.enabled`                            | false                      | Enable Horizontal Autoscaling for the FiftyOne Teams App Pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `teamsAppSettings.autoscaling.maxReplicas`                        | 20                         | Maximum Replicas for Horizontal Autoscaling in the FiftyOne Teams App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `teamsAppSettings.autoscaling.minReplicas`                        | 2                          | Minimum Replicas for Horizontal Autoscaling in the FiftyOne Teams App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `teamsAppSettings.autoscaling.targetCPUUtilizationPercentage`     | 80                         | Percent CPU Utilization for autoscaling the FiftyOne Teams App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `teamsAppSettings.autoscaling.targetMemoryUtilizationPercentage`  | 80                         | Percent Memory Utilization for autoscaling the FiftyOne Teams App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `teamsAppSettings.env`                                            | defined below              | Arbitrary environment variables to pass to the FiftyOne Teams App pod                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `teamsAppSettings.env.APP_USE_HTTPS`                              | true                       | Set to `false` if Ingress does not use HTTPS                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `teamsAppSettings.env.FIFTYONE_APP_ALLOW_MEDIA_EXPORT`            | true                       | Set this to `"false"` if you want to disable media export options                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `teamsAppSettings.env.FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION` | None                       | The recommended fiftyone SDK version. This will be displayed in install modal (i.e. `pip install ... fiftyone==0.11.0`)                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `teamsAppSettings.env.FIFTYONE_APP_THEME`                         | dark                       | The default theme configuration for your FiftyOne Teams application:<br>&ensp;- `dark`: Application will default to dark theme when user visits for the first time<br>&ensp;- `light`: Application will default to light theme when user visits for the first time<br>&ensp;- `always-dark`: Application will default to dark theme on each refresh (even if user changes theme to light within the app)<br>&ensp;- `always-light`: Application will default to light theme on each refresh (even if user changes theme to dark within the app) |
| `teamsAppSettings.image.repository`                               | voxel51/fiftyone-teams-app | Container Image for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `teamsAppSettings.image.tag`                                      | Helm Chart Version         | Container Image tag for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `teamsAppSettings.nodeSelector`                                   | None                       | FiftyOne Teams App pod node selector rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `teamsAppSettings.podAnnotations`                                 | None                       | FiftyOne Teams App pod annotation rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `teamsAppSettings.podSecurityContext`                             | None                       | FiftyOne Teams App pod security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `teamsAppSettings.replicaCount`                                   | 2                          | FiftyOne Teams App replica count if autoscaling is disabled                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `teamsAppSettings.resources.limits.cpu`                           | None                       | CPU resource limits for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `teamsAppSettings.resources.limits.memory`                        | None                       | Memory resource limits for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `teamsAppSettings.resources.requests.cpu`                         | None                       | CPU resource requests for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `teamsAppSettings.resources.requests.memory`                      | None                       | Memory resource requests for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `teamsAppSettings.securityContext`                                | None                       | FiftyOne Teams App service security context rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `teamsAppSettings.serverPathPrefix`                               | `/`                        | FiftyOne Teams App prefix for path-based Ingress routing                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `teamsAppSettings.service.annotations`                            | None                       | FiftyOne Teams App service annotations                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `teamsAppSettings.service.containerPort`                          | 3000                       | Port for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `teamsAppSettings.service.liveness.initialDelaySeconds`           | 45                         | Delay before Liveness checks for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `teamsAppSettings.service.name`                                   | teams-app                  | FiftyOne Teams App service name                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `teamsAppSettings.service.nodePort`                               | None                       | `nodePort` for Service Type `NodePort`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `teamsAppSettings.service.port`                                   | 80                         | FiftyOne Teams App service port                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `teamsAppSettings.service.readiness.initialDelaySeconds`          | 45                         | Delay before Readiness checks for FiftyOne Teams App containers                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `teamsAppSettings.service.shortname`                              | teams-app                  | Short name for port definitions (less than 15 characters)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `teamsAppSettings.service.type`                                   | ClusterIP                  | FiftyOne Teams App service type                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `teamsAppSettings.tolerations`                                    | None                       | FiftyOne Teams App service toleration rules                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `teamsAppSettings.volumeMounts`                                   | None                       | FiftyOne Teams App pod volume mount definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `teamsAppSettings.volumes`                                        | None                       | FiftyOne Teams App pod volume definitions                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |

---

## Upgrade Process Recommendations

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this upgrade.
You will need to either create a new IdP or modify your existing configuration in order to migrate to a new Auth0 Tenant.

### From Before FiftyOne Teams Version 1.1.0

The FiftyOne 0.14.0 SDK (database version 0.22.0) is _NOT_ backwards-compatible with FiftyOne Teams Database Versions prior to 0.19.0.
The FiftyOne 0.10.x SDK is not forwards compatible with current FiftyOne Teams Database Versions.
If you are using a FiftyOne SDK older than 0.11.0, upgrading the Web server will require upgrading all FiftyOne SDK installations.

Voxel51 recommends the following upgrade process for upgrading from versions prior to FiftyOne Teams version 1.1.0:

1. Make sure your installation includes the required
   [FIFTYONE_ENCRYPTION_KEY](#fiftyone-teams-upgrade-notes)
   environment variable
1. [Upgrade to FiftyOne Teams version 1.4.0](#deploying-fiftyone-teams)
   with `appSettings.env.FIFTYONE_DATABASE_ADMIN: true`
   (this is not the default in the Helm Chart for this release).
    - **NOTE:** FiftyOne SDK users will lose access to the
      FiftyOne Teams Database at this step until they upgrade to `fiftyone==0.14.0`
1. Upgrade your FiftyOne SDKs to version 0.14.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Have an admin run this to upgrade all datasets to version 0.22.0

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

### From FiftyOne Teams Version 1.1.0 and later

The FiftyOne 0.14.0 SDK is backwards-compatible with FiftyOne Teams Database Versions 0.19.0 and later.
You will not be able to connect to a FiftyOne Teams 1.4.0 database (version 0.22.0) with any FiftyOne SDK before 0.14.0.

Voxel51 always recommends using the latest version of the FiftyOne SDK compatible with your FiftyOne Teams deployment.

Voxel51 recommends the following upgrade process for upgrading from FiftyOne Teams version 1.1.0 or later:

1. Ensure all FiftyOne SDK users either
    - set `FIFTYONE_DATABASE_ADMIN=false`
    - `unset FIFTYONE_DATABASE_ADMIN`
        - This should generally be your default
1. [Upgrade to FiftyOne Teams version 1.4.0](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 0.14.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Have the admin run  to upgrade all datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. To ensure that all datasets are now at version 0.22.0, run

    ```shell
    fiftyone migrate --info
    ```

---

## Deploying FiftyOne Teams

You can find an example, minimal, `values.yaml`
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/values.yaml).

1. Edit the `values.yaml` file
1. Deploy your FiftyOne Teams instance with:

    ```shell
    helm repo add voxel51 https://helm.fiftyone.ai
    helm repo update voxel51
    helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
    ```

1. Upgrade an existing deployment with:

    ```shell
    helm repo update voxel51
    helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
    ```

To show the changes Helm will apply during installations and upgrades,
consider using
[helm diff](https://github.com/databus23/helm-diff)
Voxel51 is not affiliated with the author of this plugin.

For example:

```shell
helm diff -C1 upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
```

---

## A Full GKE Deployment Example

The following instructions represent a full Google Kubernetes Engine [GKE] deployment using these helm charts

- [jetstack/cert-manager](https://github.com/cert-manager/cert-manager)
  - For Let's Encrypt SSL certificates
- [bitnami/mongodb](https://github.com/bitnami/charts/tree/main/bitnami/mongodb)
  - for MongoDB
- voxel51/fiftyone-teams-app

These instructions assume you have

- These tools installed and operating
  - [kubectl](https://kubernetes.io/docs/tasks/tools/)
  - [Helm](https://helm.sh/docs/intro/install/)
- An existing
  [GKE Cluster available](https://cloud.google.com/kubernetes-engine/docs/concepts/kubernetes-engine-overview)
- Received Docker Hub credentials from Voxel51
  - Have `voxel51-docker.json` file in the current directory
    - If `voxel51-docker.json` is not in the current directory, please update the command line accordingly.
- Auth0 configuration information from Voxel51.
  - If you have not received this information, please contact your
    [Voxel51 Support Team](mailto:support@voxel51.com).

### Download the Example Configuration Files

Download the example configuration files from the
[Voxel51 GitHub](https://github.com/voxel51/fiftyone-teams-app-deploy/helm/gke-examples)
repository.

One way to do this might be:

```shell
curl -o values.yaml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/values.yaml
curl -o cluster-issuer.yaml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/cluster-issuer.yaml
curl -o frontend-config.yaml https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/gke-example/frontend-config.yaml
```

Update the `values.yaml` file to include

- The Auth0 configuration provided by Voxel51 in `secret.fiftyone`
- MongoDB username and password in `secret.fiftyone.mongodbConnectionString`
- `secret.fiftyone.cookieSecret`
- `secret.fiftyone.encryptionKey`
- `host` values in `teamsAppSettings.dnsName`

Assuming you follow these directions your MongoDB host will be `fiftyone-mongodb.fiftyone-mongodb.svc.cluster.local`.
<!-- Please modify that hostname if you modify these instructions. -->

### Create the Necessary Helm Repos

Add the Helm repositories

```shell
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add jetstack https://charts.jetstack.io
helm repo add voxel51 https://helm.fiftyone.ai
helm repo update
```

### Install and Configure cert-manager

If you are using a GKE Autopilot cluster, please review the information
[provided by cert-manager](https://github.com/cert-manager/cert-manager/issues/3717#issuecomment-919299192)
and adjust your installation accordingly.

```shell
kubectl create namespace cert-manager
kubectl config set-context --current --namespace cert-manager
helm install cert-manager jetstack/cert-manager --set installCRDs=true
```

You can use the cert-manager instructions to
[verify the cert-manager Installation](https://cert-manager.io/v1.4-docs/installation/verify/).

### Create a ClusterIssuer

`ClusterIssuers` are Kubernetes resources that represent certificate authorities that are able to generate signed certificates by honoring certificate signing requests.
You must create either an `Issuer` in each namespace or a `ClusterIssuer` as part of your cert-manager configuration.
Voxel51 has provided an example `ClusterIssuer` configuration (downloaded [earlier](#download-the-example-configuration-files) in this guide).

```shell
kubectl apply -f ./cluster-issuer.yml
```

### Install and Configure MongoDB

These instructions deploy a single-node MongoDB instance in your GKE cluster.
If you would like to deploy MongoDB with a replicaset configuration, please refer to the
[MongoDB Helm Chart](https://github.com/bitnami/charts/tree/master/bitnami/mongodb)
documentation.

**You will definitely want to edit the `rootUser` and `rootPassword` defined below.**

```shell
kubectl create namespace fiftyone-mongodb
kubectl config set-context --current --namespace fiftyone-mongodb
helm install fiftyone-mongodb \
    --set auth.rootPassword=<REPLACE_ME> \
    --set auth.rootUser=admin \
    --set global.namespaceOverride=fiftyone-mongodb \
    --set image.tag=4.4 \
    --set ingress.enabled=true \
    --set namespaceOverride=fiftyone-mongodb \
    bitnami/mongodb
```

Wait until the MongoDB pods are in the `Ready` state before beginning the "Install FiftyOne Teams App" instructions.

While watiing, [configure a DNS entry](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).

To determine the state of the `fiftyone-mongodb` pods, run

```shell
kubectl get pods
```

### Obtain a Global Static IP Address and Configure a DNS Entry

Reserve a global static IP address for use in your cluster:

```shell
gcloud compute addresses create \
  fiftyone-teams-static-ip --global --ip-version IPV4
gcloud compute addresses describe \
  fiftyone-teams-static-ip --global
```

Record the IP address and either create a DNS entry or contact your Voxel51 support team to have them create an appropriate `fiftyone.ai` DNS entry for you.

### Set up http to https Forwarding

```shell
kubectl apply -f frontend-config.yaml
```

### Install FiftyOne Teams App

```shell
kubectl create namespace fiftyone-teams
kubectl config set-context --current --namespace fiftyone-teams
kubectl create secret generic regcred \
    --from-file=.dockerconfigjson=./voxel51-docker.json \
    --type kubernetes.io/dockerconfigjson
helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
```

Issuing SSL Certificates can take up to 15 minutes.
Be patient while Let's Encrypt and GKE negotiate.

You can verify that your SSL certificates have been properly issued with the following curl command:

```shell
curl -I https://replace.this.dns.name
```

Your SSL certificates have been correctly issued when you see `HTTP/2 200` at the top of the response.
If, however, you encounter a `SSL certificate problem: unable to get local issuer certificate` message you should delete the certificate and allow it to recreate.

```shell
kubectl delete secret fiftyone-teams-cert-secret
```

Further instructions for debugging ACME certificates are on the
[cert-manager docs site](https://cert-manager.io/docs/faq/acme/).

Once your installation is complete, browse to `/settings/cloud_storage_credentials` and add your storage credentials to access sample data.

### Installation Complete

Congratulations! You should now be able to access your FiftyOne Teams installation at the DNS address you created
[earlier](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).
