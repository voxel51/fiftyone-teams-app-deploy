<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# fiftyone-teams-app

<!-- markdownlint-disable line-length -->
![Version: 1.5.6](https://img.shields.io/badge/Version-1.5.6-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.5.6](https://img.shields.io/badge/AppVersion-v1.5.6-informational?style=flat-square)

FiftyOne Teams is the enterprise version of the open source [FiftyOne](https://github.com/voxel51/fiftyone) project.
<!-- markdownlint-enable line-length -->

Please contact Voxel51 for more information regarding Fiftyone Teams.

<!-- toc -->

- [Initial Installation vs. Upgrades](#initial-installation-vs-upgrades)
- [FiftyOne Features](#fiftyone-features)
  - [Snapshot Archival](#snapshot-archival)
  - [FiftyOne Teams Authenticated API](#fiftyone-teams-authenticated-api)
  - [FiftyOne Teams Plugins](#fiftyone-teams-plugins)
  - [Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](#storage-credentials-and-fiftyone_encryption_key)
  - [Proxies](#proxies)
  - [Text Similarity](#text-similarity)
- [Values](#values)
- [Upgrading From Previous Versions](#upgrading-from-previous-versions)
  - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)
  - [From Before FiftyOne Teams Version 1.1.0](#from-before-fiftyone-teams-version-110)
  - [From FiftyOne Teams Version 1.1.0 and later](#from-fiftyone-teams-version-110-and-later)
- [Launch FiftyOne Teams](#launch-fiftyone-teams)

<!-- tocstop -->

We publish container images to these Docker Hub repositories

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch`
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`

For Docker Hub credentials, please contact your Voxel51 support team.

## Initial Installation vs. Upgrades

Upgrades are more frequent than new installations.
Thus, the chart's default behavior supports
upgrades and the `values.yaml` contains

```yaml
appSettings:
  env:
    FIFTYONE_DATABASE_ADMIN: false
```

When performing an initial installation,
in your `values.yaml`, set

```yaml
appSettings:
  env:
    FIFTYONE_DATABASE_ADMIN: true
```

After the initial installation, we recommend either commenting
this environment variable or changing the value to `false`.

When performing an upgrade, please review
[Upgrading From Previous Versions](#upgrading-from-previous-versions)
.

## FiftyOne Features

Consider if you will require these settings for your deployment.

### Snapshot Archival

Since version v1.5, FiftyOne Teams supports
[archiving snapshots](https://docs.voxel51.com/teams/dataset_versioning.html#snapshot-archival)
to cold storage locations to prevent filling up the MongoDB database.
To enable this feature, set the `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
environment variable to the path of a chosen storage location.

Supported locations are network mounted filesystems and cloud storage folders.

- Network mounted filesystem
  - In `values.yaml`, set the path for a Persistent Volume Claim mounted to the
    `teams-api` deployment (not necessary to mount to other deployments) in both
    - `appSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
    - `teamsAppSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
  - Mount a Persistent Volume Claim with `ReadWrite` permissions to
    the `teams-api` deployment at the `FIFTYONE_SNAPSHOTS_ARCHIVE_PATH` path.
    For an example, see
    [Plugins Storage](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md)
    .
- Cloud storage folder
  - In `values.yaml`, set the cloud storage path (for example
    `gs://my-voxel51-bucket/dev-deployment-snapshot-archives/`)
    in
    - `appSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
    - `apiSettings.env.FIFTYONE_SNAPSHOTS_ARCHIVE_PATH`
  - Ensure the
    [cloud credentials](https://docs.voxel51.com/teams/installation.html#cloud-credentials)
    loaded in the `teams-api` deployment have full edit capabilities to this bucket

See the
[configuration documentation](https://docs.voxel51.com/teams/dataset_versioning.html#dataset-versioning-configuration)
for other configuration values that control the behavior of automatic snapshot archival.

### FiftyOne Teams Authenticated API

FiftyOne Teams v1.3 introduced the capability to connect FiftyOne Teams SDKs
through the FiftyOne Teams API (instead of direct MongoDB connection).

To enable the FiftyOne Teams Authenticated API,
[expose the FiftyOne Teams API endpoint](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/expose-teams-api.md)
and
[configure your SDK](https://docs.voxel51.com/teams/api_connection.html)
.

### FiftyOne Teams Plugins

FiftyOne Teams v1.3 introduced significant enhancements for
[Plugins](https://docs.voxel51.com/plugins/index.html)
to customize and enhance functionality.

There are three modes for plugins

1. Builtin Plugins Only
    - No changes are required for this mode
1. Plugins run in the `fiftyone-app` deployment
    - To enable this mode
        - In `values.yaml`, set the path for a Persistent Volume Claim
          mounted to the `teams-api` and `fiftyone-app` deployments in both
            - `appSettings.env.FIFTYONE_PLUGINS_DIR`
            - `apiSettings.env.FIFTYONE_PLUGINS_DIR`
        - Mount a
          [Persistent Volume Claim](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md)
          that provides
            - `ReadWrite` permissions to the `teams-api` deployment
              at the `FIFTYONE_PLUGINS_DIR` path
            - `ReadOnly` permission to the `fiftyone-app` deployment
              at the `FIFTYONE_PLUGINS_DIR` path
1. Plugins run in a dedicated `teams-plugins` deployment
    - To enable this mode
        - In `values.yaml`, set
            - `pluginsSettings.enabled: true`
            - The path for a Persistent Volume Claim mounted to the
              `teams-api` and `teams-plugins` deployments in both
                - `pluginsSettings.env.FIFTYONE_PLUGINS_DIR`
                - `apiSettings.env.FIFTYONE_PLUGINS_DIR`
        - Mount a
          [Persistent Volume Claim](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md)
          that provides
            - `ReadWrite` permissions to the `teams-api` deployment
              at the `FIFTYONE_PLUGINS_DIR` path
            - `ReadOnly` permission to the `teams-plugins` deployment
              at the `FIFTYONE_PLUGINS_DIR` path
        - If you are
          [using a proxy](#proxies)
          , add the `teams-plugins` service name to your `no_proxy` and
          `NO_PROXY` environment variables.

Use the FiftyOne Teams UI to deploy plugins by navigating to `https://<DEPOY_URL>/settings/plugins`.
Early-adopter plugins installed manually must be
redeployed using the FiftyOne Teams UI.

### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

Pods based on the `fiftyone-teams-api` and `fiftyone-app`
images must include the `FIFTYONE_ENCRYPTION_KEY` variable.
This key is used to encrypt storage credentials in the MongoDB database.

The generate an `encryptionKey`, run this Python code

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

Users with `Admin` permissions may use the FiftyOne Teams UI to manage storage
credentials by navigating to `https://<DEPOY_URL>/settings/cloud_storage_credentials`.

If added via the UI, storage credentials no longer need to be
mounted into pods or provided via environment variables.

FiftyOne Teams continues to support the use of environment variables to set
storage credentials in the application context and is providing an alternate
configuration path for future functionality.

### Proxies

FiftyOne Teams supports routing traffic through proxy servers.
To configure this, set the following environment variables on

1. All pods, in the environment (`*.env`):

    ```yaml
    http_proxy: http://proxy.yourcompany.tld:3128
    https_proxy: https://proxy.yourcompany.tld:3128
    no_proxy: fiftyone-app, teams-app, teams-api, teams-cas, <your_other_exclusions>
    HTTP_PROXY: http://proxy.yourcompany.tld:3128
    HTTPS_PROXY: https://proxy.yourcompany.tld:3128
    NO_PROXY: fiftyone-app, teams-app, teams-api, teams-cas, <your_other_exclusions>
    ```

    > **NOTE**: If you have enabled a
    > [dedicated `teams-plugins`](#fiftyone-teams-plugins)
    > deployment you will need to include `teams-plugins` in your `NO_PROXY` and
    > `no_proxy` configurations

    ---

    > **NOTE**: If you have overridden your service names with `*.service.name`
    > you will need to include the override service names in your `NO_PROXY` and
    > `no_proxy` configurations instead

1. The pod based on the `fiftyone-teams-app` image (`teamsAppSettings.env`)

    ```yaml
    GLOBAL_AGENT_HTTP_PROXY: http://proxy.yourcompany.tld:3128
    GLOBAL_AGENT_HTTPS_PROXY: https://proxy.yourconpay.tld:3128
    GLOBAL_AGENT_NO_PROXY: fiftyone-app, teams-app, teams-api, teams-cas, <your_other_exclusions>
    ```

    > **NOTE**: If you have enabled a
    > [dedicated `teams-plugins`](#fiftyone-teams-plugins)
    > deployment you will need to include `teams-plugins` in your
    > `GLOBAL_AGENT_NO_PROXY` configuration

    ---

    > **NOTE**: If you have overridden your service names with `*.service.name`
    > you will need to include the override service names in your
    > `GLOBAL_AGENT_NO_PROXY` configuration instead

The `NO_PROXY`, `no_proxy`, and `GLOBAL_AGENT_NO_PROXY` values must include the
Kubernetes service names that may communicate without going through a proxy
server.
By default, these service names are

- `fiftyone-app`
- `teams-app`
- `teams-api`
- `teams-cas`

This list may also include `teams-plugins` if you have enabled a dedicated
plugins service.

If the service names were overridden in `*.service.name`, use the override
values instead.

By default, the Global Agent Proxy will log all outbound connections
and identify which connections are routed through the proxy.
To reduce the logging verbosity, add this environment variable to your `teamsAppSettings.env`

```ini
ROARR_LOG: false
```

### Text Similarity

Since version v1.2, FiftyOne Teams supports using text similarity
searches for images that are indexed with a model that
[supports text queries](https://docs.voxel51.com/user_guide/brain.html#brain-similarity-text)
.
Use the Voxel51 provided image `fiftyone-app-torch` or
build your own base image including `torch` (PyTorch).

To override the default image, add
`appSettings.image.repository` to your `values.yaml`.
For example,

```yaml
appSettings:
  image:
    repository: voxel51/fiftyone-app-torch
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| apiSettings.affinity | object | `{}` | Affinity and anti-affinity for teams-api. [Reference][affinity]. |
| apiSettings.dnsName | string | `""` | Controls whether teams-api is added to the chart's ingress. When an empty string, a rule for teams-api is not added to the chart managed ingress. When not an empty string, becomes the value to the `host` in the ingress' rule and set `ingress.api` too. |
| apiSettings.env.FIFTYONE_ENV | string | `"production"` | Controls FiftyOne GraphQL verbosity. When "production", debug mode is disabled and the default logging level is "INFO". When "development", debug mode is enabled and the default logging level is "DEBUG". Can be overridden by setting `apiSettings.env.LOGGING_LEVEL`. |
| apiSettings.env.FIFTYONE_INTERNAL_SERVICE | bool | `true` | Whether the SDK is running in an internal service context. When running in FiftyOne Teams, set to `true`. |
| apiSettings.env.GRAPHQL_DEFAULT_LIMIT | int | `10` | Default number of returned items when listing in GraphQL queries. Can be overridden in the request. |
| apiSettings.env.LOGGING_LEVEL | string | `"INFO"` | Logging level. Overrides the value of `FIFTYONE_ENV`. Can be one of "DEBUG", "INFO", "WARN", "ERROR", or "CRITICAL". |
| apiSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. [Reference][image-pull-policy]. |
| apiSettings.image.repository | string | `"voxel51/fiftyone-teams-api"` | Container image for the teams-api. |
| apiSettings.image.tag | string | `""` | Image tag for teams-api. Defaults to the chart version. |
| apiSettings.nodeSelector | object | `{}` | nodeSelector for teams-api. [Reference][node-selector]. |
| apiSettings.podAnnotations | object | `{}` | Annotations for pods for teams-api. [Reference][annotations]. |
| apiSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-api. [Reference][security-context]. |
| apiSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-api. [Reference][resources]. |
| apiSettings.securityContext | object | `{}` | Container security configuration for teams-api. [Reference][container-security-context]. |
| apiSettings.service.annotations | object | `{}` | Service annotations for teams-api. [Reference][annotations]. |
| apiSettings.service.containerPort | int | `8000` | Service container port for teams-api. |
| apiSettings.service.liveness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the liveness probe for teams-api. [Reference][probes]. |
| apiSettings.service.name | string | `"teams-api"` | Service name. |
| apiSettings.service.nodePort | int | `nil` | Service nodePort set only when `apiSettings.service.type: NodePort` for teams-api. |
| apiSettings.service.port | int | `80` | Service port for teams-api. |
| apiSettings.service.readiness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the readiness probe for teams-api. [Reference][probes]. |
| apiSettings.service.shortname | string | `"teams-api"` | Port name (maximum length is 15 characters) for teams-api. [Reference][ports]. |
| apiSettings.service.type | string | `"ClusterIP"` | Service type for teams-api. [Reference][service-type]. |
| apiSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule pods with matching taints for teams-api. [Reference][taints-and-tolerations]. |
| apiSettings.volumeMounts | list | `[]` | Volume mounts for teams-api. [Reference][volumes]. |
| apiSettings.volumes | list | `[]` | Volumes for teams-api. [Reference][volumes]. |
| appSettings.affinity | object | `{}` | Affinity and anti-affinity for fiftyone-app. [Reference][affinity]. |
| appSettings.autoscaling.enabled | bool | `false` | Controls horizontal pod autoscaling for fiftyone-app. [Reference][autoscaling]. |
| appSettings.autoscaling.maxReplicas | int | `20` | Maximum replicas for horizontal pod autoscaling for fiftyone-app. |
| appSettings.autoscaling.minReplicas | int | `2` | Minimum Replicas for horizontal pod autoscaling for fiftyone-app. |
| appSettings.autoscaling.targetCPUUtilizationPercentage | int | `80` | Percent CPU utilization for autoscaling for fiftyone-app. |
| appSettings.autoscaling.targetMemoryUtilizationPercentage | int | `80` | Percent memory utilization for autoscaling for fiftyone-app. |
| appSettings.env.FIFTYONE_DATABASE_ADMIN | bool | `false` | Controls whether the client is allowed to trigger database migrations. [Reference][fiftyone-config]. |
| appSettings.env.FIFTYONE_INTERNAL_SERVICE | bool | `true` | Whether the SDK is running in an internal service context. When running in FiftyOne Teams, set to `true`. |
| appSettings.env.FIFTYONE_MEDIA_CACHE_APP_IMAGES | bool | `false` | Controls whether cloud media images will be downloaded and added to the local cache upon viewing media in the app. |
| appSettings.env.FIFTYONE_MEDIA_CACHE_SIZE_BYTES | int | `-1` | Set the media cache size (in bytes) for the local FiftyOne App processes. The default value is 32 GiB. `-1` is disabled. |
| appSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. [Reference][image-pull-policy]. |
| appSettings.image.repository | string | `"voxel51/fiftyone-app"` | Container image for fiftyone-app. |
| appSettings.image.tag | string | `""` | Image tag for fiftyone-app. Defaults to the chart version. |
| appSettings.nodeSelector | object | `{}` | nodeSelector for fiftyone-app. [Reference][node-selector]. |
| appSettings.podAnnotations | object | `{}` | Annotations for pods for fiftyone-app. [Reference][annotations]. |
| appSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for fiftyone-app. [Reference][security-context]. |
| appSettings.replicaCount | int | `2` | Number of pods in the fiftyone-app deployment's ReplicaSet. Ignored when `appSettings.autoscaling.enabled: true`. [Reference][deployment]. |
| appSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for fiftyone-app. [Reference][resources]. |
| appSettings.securityContext | object | `{}` | Container security configuration for fiftyone-app. [Reference][container-security-context]. |
| appSettings.service.annotations | object | `{}` | Service annotations for fiftyone-app. [Reference][annotations]. |
| appSettings.service.containerPort | int | `5151` | Service container port for fiftyone-app. |
| appSettings.service.liveness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the liveness probe for fiftyone-app. [Reference][probes]. |
| appSettings.service.name | string | `"fiftyone-app"` | Service name. |
| appSettings.service.nodePort | int | `nil` | Service nodePort set only when `appSettings.service.type: NodePort` for fiftyone-app. |
| appSettings.service.port | int | `80` | Service port. |
| appSettings.service.readiness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the readiness probe for fiftyone-app. [Reference][probes]. |
| appSettings.service.shortname | string | `"fiftyone-app"` | Port name (maximum length is 15 characters) for fiftyone-app. [Reference][ports]. |
| appSettings.service.type | string | `"ClusterIP"` | Service type for fiftyone-app. [Reference][service-type]. |
| appSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule fiftyone-app pods with matching taints. [Reference][taints-and-tolerations]. |
| appSettings.volumeMounts | list | `[]` | Volume mounts for fiftyone-app. [Reference][volumes]. |
| appSettings.volumes | list | `[]` | Volumes for fiftyone-app. [Reference][volumes]. |
| casSettings.affinity | object | `{}` | Affinity and anti-affinity for teams-cas. [Reference][affinity]. |
| casSettings.env.CAS_DATABASE_NAME | string | `"cas"` | Provide the name for the CAS database |
| casSettings.env.CAS_DEFAULT_USER_ROLE | string | `"GUEST"` | Set the default user role for new users One of `GUEST`, `COLLABORATOR`, `MEMBER`, `ADMIN` |
| casSettings.env.CAS_LOG_LEVEL | string | `"INFO"` | Set the CAS Log Level One of `DEBUG`, `INFO`, `WARN`, `ERROR` |
| casSettings.env.CAS_MONGODB_URI_KEY | string | `"mongodbConnectionString"` | The key from `secret.fiftyone` that contains the CAS MongoDB Connection String. |
| casSettings.env.ENABLE_LEGACY_MODE | bool | `true` | Toggle CAS Legacy Mode, which continues to use Auth0 integration |
| casSettings.env.FEATURE_FLAG_ENABLE_INVITATIONS | bool | `true` | Allow Admins to invite users by email NOTE: This is not supported when ENABLE_LEGACY_MODE is `false` |
| casSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. [Reference][image-pull-policy]. |
| casSettings.image.repository | string | `"voxel51/teams-cas"` | Container image for teams-cas. |
| casSettings.image.tag | string | `""` | Image tag for teams-cas. Defaults to the chart version. |
| casSettings.nodeSelector | object | `{}` | nodeSelector for teams-cas. [Reference][node-selector]. |
| casSettings.podAnnotations | object | `{}` | Annotations for pods for teams-cas. [Reference][annotations]. |
| casSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-cas. [Reference][security-context]. |
| casSettings.replicaCount | int | `2` | Number of pods in the teams-cas deployment's ReplicaSet. [Reference][deployment]. |
| casSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-cas. [Reference][resources]. |
| casSettings.securityContext | object | `{}` | Container security configuration for teams-cas. [Reference][container-security-context]. |
| casSettings.service.annotations | object | `{}` | Service annotations for teams-cas. [Reference][annotations]. |
| casSettings.service.containerPort | int | `3000` | Service container port for teams-cas. |
| casSettings.service.liveness.initialDelaySeconds | int | `15` | Number of seconds to wait before performing the liveness probe for fiftyone-app. [Reference][probes]. |
| casSettings.service.name | string | `"teams-cas"` | Service name. |
| casSettings.service.nodePort | int | `nil` | Service nodePort set only when `casSettings.service.type: NodePort` for teams-cas. |
| casSettings.service.port | int | `80` | Service port. |
| casSettings.service.readiness.initialDelaySeconds | int | `15` | Number of seconds to wait before performing the readiness probe for fiftyone-app. [Reference][probes]. |
| casSettings.service.shortname | string | `"teams-cas"` | Port name (maximum length is 15 characters) for teams-cas. [Reference][ports]. |
| casSettings.service.type | string | `"ClusterIP"` | Service type for teams-cas. [Reference][service-type]. |
| casSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule teams-cas pods with matching taints. [Reference][taints-and-tolerations]. |
| casSettings.volumeMounts | list | `[]` | Volume mounts for teams-cas. [Reference][volumes]. |
| casSettings.volumes | list | `[]` | Volumes for teams-cas. [Reference][volumes]. |
| imagePullSecrets | list | `[]` | Container image registry keys. [Reference][image-pull-secrets]. |
| ingress.annotations | object | `{}` | Ingress annotations. [Reference][annotations]. |
| ingress.api | object | `{"path":"/*","pathType":"ImplementationSpecific"}` | The ingress rule values for teams-api, when `apiSettings.dnsName` is not empty. [Reference][ingress-rules]. |
| ingress.className | string | `""` | Name of the ingress class.  When empty, a default Ingress class should be defined. When not empty and Kubernetes version is >1.18.0, this value will be the Ingress class name. [Reference][ingress-default-ingress-class] |
| ingress.enabled | bool | `true` | Controls whether to create the ingress. When `false`, uses a pre-existing ingress. [Reference][ingress]. |
| ingress.labels | object | `{}` | Additional labels for the ingress. [Reference][labels-and-selectors]. |
| ingress.paths | list | `[{"path":"/cas","pathType":"Prefix","serviceName":"teams-cas","servicePort":80},{"path":"/*","pathType":"ImplementationSpecific","serviceName":"teams-app","servicePort":80}]` | Additional ingress rules for the host `teamsAppSettings.dnsName` for the chart managed ingress (when `ingress.enabled: true`). [Reference][ingress-rules]. |
| ingress.paths[0] | object | `{"path":"/cas","pathType":"Prefix","serviceName":"teams-cas","servicePort":80}` | Ingress path for teams-cas |
| ingress.paths[0].pathType | string | `"Prefix"` | Ingress path type |
| ingress.paths[0].serviceName | string | `"teams-cas"` | Ingress path service name |
| ingress.paths[0].servicePort | int | `80` | Ingress path service port |
| ingress.paths[1] | object | `{"path":"/*","pathType":"ImplementationSpecific","serviceName":"teams-app","servicePort":80}` | Ingress path for teams-app |
| ingress.paths[1].pathType | string | `"ImplementationSpecific"` | Ingress path type |
| ingress.paths[1].serviceName | string | `"teams-app"` | Ingress path service name |
| ingress.paths[1].servicePort | int | `80` | Ingress path service port |
| ingress.tlsEnabled | bool | `true` | Controls whether the chart managed ingress contains a `spec.tls` stanza. |
| ingress.tlsSecretName | string | `"fiftyone-teams-tls-secret"` | Name of secret containing TLS certificate for teams-app. Certificate should contain the host names `apiSettings.dnsName` and `teamsAppSettings.dnsName`. When `ingress.tlsEnabled=True`, sets's the value of ingress' `spec.tls[0].secretName`. |
| namespace.create | bool | `false` | Controls whether to create the namespace. When `false`, the namespace must already exists. |
| namespace.name | string | `"fiftyone-teams"` | The namespace name used for chart resources. |
| pluginsSettings.affinity | object | `{}` | Affinity and anti-affinity for teams-plugins. [Reference][affinity]. |
| pluginsSettings.autoscaling.enabled | bool | `false` | Controls horizontal pod autoscaling for teams-plugins. [Reference][autoscaling]. |
| pluginsSettings.autoscaling.maxReplicas | int | `20` | Maximum replicas for horizontal pod autoscaling for teams-plugins. |
| pluginsSettings.autoscaling.minReplicas | int | `2` | Minimum Replicas for horizontal pod autoscaling for teams-plugins. |
| pluginsSettings.autoscaling.targetCPUUtilizationPercentage | int | `80` | Percent CPU utilization for autoscaling for teams-plugins. |
| pluginsSettings.autoscaling.targetMemoryUtilizationPercentage | int | `80` | Percent memory utilization for autoscaling for teams-plugins. |
| pluginsSettings.enabled | bool | `false` | Controls whether to create a dedicated "teams-plugins" deployment. |
| pluginsSettings.env.FIFTYONE_INTERNAL_SERVICE | bool | `true` | Whether the SDK is running in an internal service context. When running in FiftyOne Teams, set to `true`. |
| pluginsSettings.env.FIFTYONE_MEDIA_CACHE_APP_IMAGES | bool | `false` | Controls whether cloud media images will be downloaded and added to the local cache upon viewing media in the app. |
| pluginsSettings.env.FIFTYONE_MEDIA_CACHE_SIZE_BYTES | int | `-1` | Set the media cache size (in bytes) for the local FiftyOne Plugins processes. The default value is 32 GiB. `-1` is disabled. |
| pluginsSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. [Reference][image-pull-policy]. |
| pluginsSettings.image.repository | string | `"voxel51/fiftyone-app"` | Container image for teams-plugins. |
| pluginsSettings.image.tag | string | `""` | Image tag for teams-plugins. Defaults to the chart version. |
| pluginsSettings.nodeSelector | object | `{}` | nodeSelector for teams-plugins. [Reference][node-selector]. |
| pluginsSettings.podAnnotations | object | `{}` | Annotations for teams-plugins pods. [Reference][annotations]. |
| pluginsSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-plugins. [Reference][security-context]. |
| pluginsSettings.replicaCount | int | `2` | Number of pods in the teams-plugins deployment's ReplicaSet. Ignored when `pluginsSettings.autoscaling.enabled: true`. [Reference][deployment]. |
| pluginsSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-plugins. [Reference][resources]. |
| pluginsSettings.securityContext | object | `{}` | Container security configuration for teams-plugins. [Reference][container-security-context]. |
| pluginsSettings.service.annotations | object | `{}` | Service annotations for teams-plugins. [Reference][annotations]. |
| pluginsSettings.service.containerPort | int | `5151` | Service container port for teams-plugins. |
| pluginsSettings.service.liveness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the liveness probe teams-plugins. [Reference][probes]. |
| pluginsSettings.service.name | string | `"teams-plugins"` | Service name. |
| pluginsSettings.service.nodePort | int | `nil` | Service nodePort set only when `pluginsSettings.service.type: NodePort` for teams-plugins. |
| pluginsSettings.service.port | int | `80` | Service port. |
| pluginsSettings.service.readiness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the readiness probe for teams-plugins. [Reference][probes]. |
| pluginsSettings.service.shortname | string | `"teams-plugins"` | Port name (maximum length is 15 characters) for teams-plugins. [Reference][ports]. |
| pluginsSettings.service.type | string | `"ClusterIP"` | Service type for teams-plugins. [Reference][service-type]. |
| pluginsSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule teams-plugins pods with matching taints. [Reference][taints-and-tolerations]. |
| pluginsSettings.volumeMounts | list | `[]` | Volume mounts for teams-plugins pods. [Reference][volumes]. |
| pluginsSettings.volumes | list | `[]` | Volumes for teams-plugins. [Reference][volumes]. |
| secret.create | bool | `true` | Controls whether to create the secret named `secret.name`. |
| secret.fiftyone.apiClientId | string | `""` | Voxel51-provided Auth0 API Client ID. |
| secret.fiftyone.apiClientSecret | string | `""` | Voxel51-provided Auth0 API Client Secret. |
| secret.fiftyone.auth0Domain | string | `""` | Voxel51-provided Auth0 Domain. |
| secret.fiftyone.clientId | string | `""` | Voxel51-provided Auth0 Client ID. |
| secret.fiftyone.clientSecret | string | `""` | Voxel51-provided Auth0 Client Secret. |
| secret.fiftyone.cookieSecret | string | `""` | A randomly generated string for cookie encryption. To generate, run `openssl rand -hex 32`. |
| secret.fiftyone.encryptionKey | string | `""` | Encryption key for storage credentials. [Reference][fiftyone-encryption-key]. |
| secret.fiftyone.fiftyoneAuthSecret | string | `""` | A randomly generated string for CAS Authentication. |
| secret.fiftyone.fiftyoneDatabaseName | string | `""` | MongoDB Database Name for FiftyOne Teams. |
| secret.fiftyone.mongodbConnectionString | string | `""` | MongoDB Connection String. [Reference][mongodb-connection-string]. |
| secret.fiftyone.organizationId | string | `""` | Voxel51-provided Auth0 Organization ID. |
| secret.name | string | `"fiftyone-teams-secrets"` | Name of the secret (existing or to be created) in the namespace `namespace.name`. |
| serviceAccount.annotations | object | `{}` | Service Account annotations. [Reference][annotations]. |
| serviceAccount.create | bool | `true` | Controls whether to create the service account named `serviceAccount.name`. |
| serviceAccount.name | string | `"fiftyone-teams"` | Name of the service account (existing or to be created) in the namespace `namespace.name` used for deployments. [Reference][service-account]. |
| teamsAppSettings.affinity | object | `{}` | Affinity and anti-affinity for teams-app. [Reference][affinity]. |
| teamsAppSettings.autoscaling.enabled | bool | `false` | Controls horizontal pod autoscaling for teams-app. [Reference][autoscaling]. |
| teamsAppSettings.autoscaling.maxReplicas | int | `5` | Maximum Replicas for horizontal autoscaling for teams-app. |
| teamsAppSettings.autoscaling.minReplicas | int | `2` | Minimum Replicas for horizontal autoscaling for teams-app. |
| teamsAppSettings.autoscaling.targetCPUUtilizationPercentage | int | `80` | Percent CPU utilization for autoscaling for teams-app. |
| teamsAppSettings.autoscaling.targetMemoryUtilizationPercentage | int | `80` | Percent memory utilization for autoscaling for teams-app. |
| teamsAppSettings.dnsName | string | `""` | DNS Name for the teams-app service. Used in the chart managed ingress (`spec.tls.hosts` and `spec.rules[0].host`) and teams-app deployment environment variable `AUTH0_BASE_URL`. |
| teamsAppSettings.env.APP_USE_HTTPS | bool | `true` | Controls the protocol of the teams-app. Configure your ingress to match. When `true`, uses the https protocol. When `false`, uses the http protocol. |
| teamsAppSettings.env.FIFTYONE_APP_ALLOW_MEDIA_EXPORT | bool | `true` | When `false`, disables media export options |
| teamsAppSettings.env.FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION | string | `"0.15.6"` | The recommended fiftyone SDK version that will be displayed in the install modal (i.e. `pip install ... fiftyone==0.11.0`). |
| teamsAppSettings.env.FIFTYONE_APP_THEME | string | `"dark"` | The default theme configuration. `dark`: Theme will be dark when user visits for the first time. `light`: Theme will be light theme when user visits for the first time. `always-dark`: Sets dark theme on each refresh (overrides user theme changes in the app). `always-light`: Sets light theme on each refresh (overrides user theme changes in the app). |
| teamsAppSettings.env.RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED | bool | `false` | Disable duplicate atom/selector key checking that generated false-positive errors. [Reference][recoil-env]. |
| teamsAppSettings.fiftyoneApiOverride | string | `""` | Overrides the `FIFTYONE_API_URI` environment variable. When set `FIFTYONE_API_URI` controls the value shown in the API Key Modal providing guidance for connecting to the FiftyOne Teams API. `FIFTYONE_API_URI` uses the value from apiSettings.dnsName if it is set, or uses the teamsAppSettings.dnsName |
| teamsAppSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. Reference][image-pull-policy]. |
| teamsAppSettings.image.repository | string | `"voxel51/fiftyone-teams-app"` | Container image for teams-app. |
| teamsAppSettings.image.tag | string | `""` | Image tag for teams-app. Defaults to the chart version. |
| teamsAppSettings.nodeSelector | object | `{}` | nodeSelector for teams-app. [Reference][node-selector]. |
| teamsAppSettings.podAnnotations | object | `{}` | Annotations for teams-app pods. [Reference][annotations]. |
| teamsAppSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-app. [Reference][security-context]. |
| teamsAppSettings.replicaCount | int | `2` | Number of pods in the teams-app deployment's ReplicaSet. Ignored when `teamsAppSettings.autoscaling.enabled: true`. [Reference][deployment]. |
| teamsAppSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-app. [Reference][resources]. |
| teamsAppSettings.securityContext | object | `{}` | Container security configuration for teams-app. [Reference][container-security-context]. |
| teamsAppSettings.serverPathPrefix | string | `"/"` | Prefix for path-based Ingress routing for teams-app. |
| teamsAppSettings.service.annotations | object | `{}` | Service annotations for teams-app. [Reference][annotations]. |
| teamsAppSettings.service.containerPort | int | `3000` | Service container port for teams-app. |
| teamsAppSettings.service.liveness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the liveness probe for teams-app. [Reference][probes]. |
| teamsAppSettings.service.name | string | `"teams-app"` | Service name. |
| teamsAppSettings.service.nodePort | int | `nil` | Service nodePort set only when `teamsAppSettings.service.type: NodePort` for teams-app. |
| teamsAppSettings.service.port | int | `80` | Service port. |
| teamsAppSettings.service.readiness.initialDelaySeconds | int | `45` | Number of seconds to wait before performing the readiness probe for teams-app. [Reference][probes]. |
| teamsAppSettings.service.shortname | string | `"teams-app"` | Port name (maximum length is 15 characters) for teams-app. [Reference][ports]. |
| teamsAppSettings.service.type | string | `"ClusterIP"` | Service type for teams-app. [Reference][service-type]. |
| teamsAppSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule teams-app pods with matching taints. [Reference][taints-and-tolerations]. |
| teamsAppSettings.volumeMounts | list | `[]` | Volume mounts for teams-app pods. [Reference][volumes]. |
| teamsAppSettings.volumes | list | `[]` | Volumes for teams-app pods. [Reference][volumes]. |

## Upgrading From Previous Versions

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success
team member to coordinate this upgrade.
You will need to either create a new Identity Provider (IdP)
or modify your existing configuration to migrate to a new Auth0 Tenant.

### From Before FiftyOne Teams Version 1.1.0

The FiftyOne 0.15.6 SDK (database version 0.23.5) is _NOT_ backwards-compatible
with FiftyOne Teams Database Versions prior to 0.19.0.
The FiftyOne 0.10.x SDK is not forwards compatible
with current FiftyOne Teams Database Versions.
If you are using a FiftyOne SDK version older than 0.11.0, upgrading the Web
server will require upgrading all FiftyOne SDK installations.

Voxel51 recommends this upgrade process from
versions prior to FiftyOne Teams version 1.1.0:

1. In your `values.yaml`, set the required
   [FIFTYONE_ENCRYPTION_KEY](#storage-credentials-and-fiftyone_encryption_key)
   environment variable
1. [Upgrade to FiftyOne Teams version 1.5.6](#launch-fiftyone-teams)
   with `appSettings.env.FIFTYONE_DATABASE_ADMIN: true`
   (this is not the default value in `values.yaml` and must be overridden).
    > **NOTE:** At this step, FiftyOne SDK users will lose access to the
    > FiftyOne Teams Database until they upgrade to `fiftyone==0.15.6`
1. Upgrade your FiftyOne SDKs to version 0.15.6
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Check if the datasets were migrated to version 0.23.5

    ```shell
    fiftyone migrate --info
    ```

    - If not all datasets have been upgraded, have an admin run

        ```shell
        FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
        ```

### From FiftyOne Teams Version 1.1.0 and later

The FiftyOne 0.15.6 SDK is backwards-compatible with
FiftyOne Teams Database Versions 0.19.0 and later.
You will not be able to connect to a FiftyOne Teams 1.5.6
database (version 0.23.5) with any FiftyOne SDK before 0.15.6.

We recommend using the latest version of the FiftyOne SDK
compatible with your FiftyOne Teams deployment.

We recommend the following upgrade process for
upgrading from FiftyOne Teams version 1.1.0 or later:

1. Ensure all FiftyOne SDK users either
    - set `FIFTYONE_DATABASE_ADMIN=false`
    - `unset FIFTYONE_DATABASE_ADMIN`
        - This should generally be your default
1. [Upgrade to FiftyOne Teams version 1.5.6](#launch-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 0.15.6
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Have an admin run to upgrade all datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

    > **NOTE** Any FiftyOne SDK less than 0.15.6 will lose database connectivity
    >  at this point. Upgrading to `fiftyone==0.15.6` is required

1. Validate that all datasets are now at version 0.23.5, by running

    ```shell
    fiftyone migrate --info
    ```

## Launch FiftyOne Teams

A minimal example `values.yaml` may be found
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/values.yaml)
.

1. Edit the `values.yaml` file
1. Deploy FiftyOne Teams with `helm install`
    1. For a new installation, run

        ```shell
        helm repo add voxel51 https://helm.fiftyone.ai
        helm repo update voxel51
        helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
        ```

    1. To upgrade an existing helm installation, run

        ```shell
        helm repo update voxel51
        helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
        ```

        > **NOTE**  To view the changes Helm would apply during installations
        > and upgrades, consider using
        > [helm diff](https://github.com/databus23/helm-diff)
        > .
        > Voxel51 is not affiliated with the author of this plugin.
        >
        >    For example:
        >
        >    ```shell
        >    helm diff -C1 upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
        >    ```

<!-- Reference Links -->
[affinity]: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/
[annotations]: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
[autoscaling]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[container-security-context]: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container
[deployment]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
[image-pull-policy]: https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy
[image-pull-secrets]: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
[ingress-default-ingress-class]: https://kubernetes.io/docs/concepts/services-networking/ingress/#default-ingress-class
[ingress-rules]: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules
[ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[labels-and-selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
[node-selector]: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
[ports]: https://kubernetes.io/docs/concepts/services-networking/service/#field-spec-ports
[probes]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
[resources]: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
[security-context]: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
[service-account]: https://kubernetes.io/docs/concepts/security/service-accounts/
[service-type]: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
[taints-and-tolerations]: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
[volumes]: https://kubernetes.io/docs/concepts/storage/volumes/

[mongodb-connection-string]: https://www.mongodb.com/docs/manual/reference/connection-string/

[recoil-env]: https://recoiljs.org/docs/api-reference/core/RecoilEnv/

[fiftyone-encryption-key]: https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/README.md#storage-credentials-and-fiftyone_encryption_key
[fiftyone-config]: https://docs.voxel51.com/user_guide/config.html
