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
![Version: 2.2.0](https://img.shields.io/badge/Version-2.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v2.2.0](https://img.shields.io/badge/AppVersion-v2.2.0-informational?style=flat-square)

FiftyOne Teams is the enterprise version of the open source [FiftyOne](https://github.com/voxel51/fiftyone) project.
<!-- markdownlint-enable line-length -->

Please contact Voxel51 for more information regarding Fiftyone Teams.

## FiftyOne Teams v2.0+ Requires a License File

FiftyOne Teams v2.0 introduces a new requirement for a license file.  This
license file should be obtained from your Customer Success Team before
upgrading to FiftyOne Teams 2.0 or beyond.

The license file now contains all of the Auth0 configuration that was
previously provided through kubernetes secrets; you may remove those secrets
from your `values.yaml` and from any secrets created outside of the Voxel51
install process.

Use the license file provided by the Voxel51 Customer Success Team to create
a new license file secret:

```shell
kubectl --namespace your-namepace-here create secret generic fiftyone-license \
--from-file=license=./your-license-file
```

## FiftyOne Teams v2.2+ Delegated Operator Changes

FiftyOne Teams v2.2 introduces some changes to delegated operators, detailed
below.

### Delegated Operation Capacity

By default, all deployments are provisioned with capacity to support up to 3
delegated operations simultaneously. You will need to configure the [builtin
orchestrator](#builtin-delegated-operator-orchestrator) or an external
orchestrator, with enough workers, to be able to utilize this full capacity.
If your team finds the usage is greater than this, please reach out to your
Voxel51 support team for guidance and to increase this limit!

### Existing Orchestrators

> [!NOTE]
> If you are currently utilizing an external orchestrator for delegated
> operations, such as Airflow or Flyte, you may have an outdated execution
> definition that could negatively affect the experience. Please reach out to
> Voxel51 support team for guidance on updating this code.

Additionally,

> [!WARNING]
> If you cannot update the orchestrator DAG/workflow code, you must set
> `delegatedOperatorExecutorSettings.env.FIFTYONE_ALLOW_LEGACY_ORCHESTRATORS: true`
> in `values.yaml` in order for the delegated operation system to function
> properly.

## Known Issues for FiftyOne Teams v1.6.0 and Above

### Invitations Disabled for Internal Authentication Mode (v1.6.0-v2.1.X)

FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS), which
includes both
[`legacy` authentication mode][legacy-auth-mode]
and
[`internal` authentication mode][internal-auth-mode].

Prior to v2.2.0, inviting users to join your FiftyOne Teams instance was not supported
when `FIFTYONE_AUTH_MODE` is set to `internal`.
Starting in v2.2.0+, you can enable invitations for your organization through the
CAS SuperAdmin UI. To enable sending invitations as emails, you must also
configure an SMTP connection.

## Table of Contents

<!-- toc -->

- [Initial Installation vs. Upgrades](#initial-installation-vs-upgrades)
- [FiftyOne Teams Features](#fiftyone-teams-features)
  - [Builtin Delegated Operator Orchestrator](#builtin-delegated-operator-orchestrator)
  - [Central Authentication Service](#central-authentication-service)
  - [Snapshot Archival](#snapshot-archival)
  - [FiftyOne Teams Authenticated API](#fiftyone-teams-authenticated-api)
  - [FiftyOne Teams Plugins](#fiftyone-teams-plugins)
    - [Builtin Plugins Only](#builtin-plugins-only)
    - [Shared Plugins](#shared-plugins)
    - [Dedicated Plugins](#dedicated-plugins)
  - [Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](#storage-credentials-and-fiftyone_encryption_key)
  - [Proxies](#proxies)
  - [Text Similarity](#text-similarity)
- [Requirements](#requirements)
- [Values](#values)
- [Upgrading From Previous Versions](#upgrading-from-previous-versions)
  - [From Early Adopter Versions (Versions less than 1.0)](#from-early-adopter-versions-versions-less-than-10)
  - [From Before FiftyOne Teams Version 1.1.0](#from-before-fiftyone-teams-version-110)
  - [From FiftyOne Teams Versions After 1.1.0 and Before Version 1.6.0](#from-fiftyone-teams-versions-after-110-and-before-version-160)
  - [From FiftyOne Teams Versions 1.6.0 to 1.7.1](#from-fiftyone-teams-versions-160-to-171)
  - [From FiftyOne Teams Version 2.0.0](#from-fiftyone-teams-version-200)
- [Deploying FiftyOne Teams](#deploying-fiftyone-teams)
  - [Deploying On GKE](#deploying-on-gke)

<!-- tocstop -->

We publish the following FiftyOne Teams private images to Docker Hub:

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch`
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`
- `voxel51/fiftyone-teams-cas`

For Docker Hub credentials, please contact your Voxel51 support team.

## Initial Installation vs. Upgrades

Upgrades are more frequent than new installations.
The chart's default behavior supports upgrades and the `values.yaml` contains

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
[Upgrading From Previous Versions](#upgrading-from-previous-versions).

## FiftyOne Teams Features

Consider if you will require these settings for your deployment.

### Builtin Delegated Operator Orchestrator

FiftyOne Teams v2.2 introduces a builtin orchestrator to run
[Delegated Operations](https://docs.voxel51.com/teams/teams_plugins.html#delegated-operations),
instead of (or in addition to) configuring your own orchestrator such as Airflow.

This option can be added to any of the 3 existing
[plugin modes](#fiftyone-teams-plugins). If you're using the builtin-operator
only option, the Persistent Volume Claim should be omitted.

To enable this mode

- In `values.yaml`, set
  - `delegatedOperatorExecutorSettings.enabled: true`
  - The path for a Persistent Volume Claim mounted to the
    `teams-do` deployment in
    - `delegatedOperatorExecutorSettings.env.FIFTYONE_PLUGINS_DIR`
- See
  [Adding Shared Storage for FiftyOne Teams Plugins](../docs/plugins-storage.md)
  - Mount a Persistent Volume Claim (PVC) that provides
    - `ReadWrite` permissions to the `teams-do` deployment
      at the `FIFTYONE_PLUGINS_DIR` path

Optionally, the logs generated during running of a delegated operation can be
uploaded to a network-mounted file system or cloud storage path that is
available to this deployment. Logs are uploaded in this format:
`<configured_path>/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log`
In `values.yaml`, set `configured_path`

- `delegatedOperatorExecutorSettings.env.FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH`

To use plugins with custom dependencies, build and use
[Custom Plugins Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).

### Central Authentication Service

FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
CAS requires additional configurations and consumes additional resources.
Please review these notes, and the
[Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
documentation before completing your upgrade.

Voxel51 recommends upgrading your deployment using
[`legacy` authentication mode][legacy-auth-mode]
and migrating to
[`internal` authentication mode][internal-auth-mode]
after confirming your initial upgrade was successful.

Please contact your Voxel51 customer success representative for assistance
in migrating to internal mode.

The CAS service requires changes to your `values.yaml` files.
A brief summary of those changes include

- Add the `fiftyoneAuthSecret` secret to either
  - `secret.fiftyone`
  - secret specified in `secret.name`

When using path-based routing, update your `values.yaml`
to include the rule (add it before the `path: /` rule)

```yaml
- path: /cas
  pathType: Prefix
  serviceName: teams-cas
  servicePort: 80
```

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
    [Plugins Storage][plugins-storage].
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
[configure your SDK](https://docs.voxel51.com/teams/api_connection.html).

### FiftyOne Teams Plugins

FiftyOne Teams v1.3 introduced significant enhancements for
[Plugins](https://docs.voxel51.com/plugins/index.html)
to customize and enhance functionality.

There are three modes for plugins

1. Builtin Plugins Only
    - This is the default mode
    - Users may only run the builtin plugins shipped with Fiftyone Teams
    - Cannot run custom plugins
1. Shared Plugins
    - Users may run builtin and custom plugins
    - Requires creating a Persistent Volume backed by NFSwith the PVCs
      - `teams-api` (ReadWrite)
      - `fiftyone-app` (ReadOnly)
    - Plugins run in the existing `fiftyone-app` deployment
      - Plugins resource consumption may starve `fiftyone-app`,
        causing the app to be slow or crash
1. Dedicated Plugins
    - Users may run builtin and custom plugins
    - Plugins run in an additional `teams-plugins` deployment
    - Requires creating a Persistent Volume backed by NFS with the PVCs
      - `teams-plugins` (ReadWrite)
      - `fiftyone-app` (ReadOnly)
    - Plugins run in a dedicated `teams-plugins` deployment
      - Plugins resource consumption does not affect `fiftyone-app`

To use plugins with custom dependencies, build and use
[Custom Plugins Images](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/docs/custom-plugins.md).

To use the FiftyOne Teams UI to deploy plugins,
navigate to `https://<DEPLOY_URL>/settings/plugins`.
Early-adopter plugins installed manually must
be redeployed using the FiftyOne Teams UI.

#### Builtin Plugins Only

Enabled by default.
No additional configurations are required.

#### Shared Plugins

Plugins run in the `fiftyone-app` deployment.
To enable this mode

- In `values.yaml`, set the path for a Persistent Volume Claim (PVC)
  mounted to the `teams-api` and `fiftyone-app` deployments in both
  - `appSettings.env.FIFTYONE_PLUGINS_DIR`
  - `apiSettings.env.FIFTYONE_PLUGINS_DIR`
- See
  [Adding Shared Storage for FiftyOne Teams Plugins][plugins-storage]
  - Mount a PVC that provides
    - `ReadWrite` permissions to the `teams-api` deployment
      at the `FIFTYONE_PLUGINS_DIR` path
    - `ReadOnly` permission to the `fiftyone-app` deployment
      at the `FIFTYONE_PLUGINS_DIR` path

#### Dedicated Plugins

To enable this mode

- In `values.yaml`, set
  - `pluginsSettings.enabled: true`
  - The path for a Persistent Volume Claim mounted to the
    `teams-api` and `teams-plugins` deployments in both
    - `pluginsSettings.env.FIFTYONE_PLUGINS_DIR`
    - `apiSettings.env.FIFTYONE_PLUGINS_DIR`
- See
  [Adding Shared Storage for FiftyOne Teams Plugins][plugins-storage]
  - Mount a Persistent Volume Claim (PVC) that provides
    - `ReadWrite` permissions to the `teams-api` deployment
      at the `FIFTYONE_PLUGINS_DIR` path
    - `ReadOnly` permission to the `teams-plugins` deployment
      at the `FIFTYONE_PLUGINS_DIR` path
- If you are
  [using a proxy](#proxies),
  add the `teams-plugins` service name to your `no_proxy` and
  `NO_PROXY` environment variables.

### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

Pods based on the `fiftyone-teams-cas`, `fiftyone-teams-api`,
and `fiftyone-app` images must include the `FIFTYONE_ENCRYPTION_KEY` variable.
This key is used to encrypt storage credentials in the MongoDB database.

To generate a value for `secret.fiftyone.encryptionKey`, run this
Python code and add the output to your `values.yaml` override file,
or to your deployment's secret

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
configuration path.

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

1. The deployments based on the `fiftyone-teams-app` (`teamsAppSettings.env`) or
   `fiftyone-teams-cas` (`casSettings.env`) images

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
[supports text queries](https://docs.voxel51.com/user_guide/brain.html#brain-similarity-text).
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

## Requirements

Kubernetes: `>=1.18-0`

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
| apiSettings.initContainers.enabled | bool | `true` | Whether to enable init containers for teams-api. [Reference][init-containers]. |
| apiSettings.initContainers.image.repository | string | `"docker.io/busybox"` | Init container images repositories for teams-api. [Reference][init-containers]. |
| apiSettings.initContainers.image.tag | string | `"stable-glibc"` | Init container images tags for teams-api. [Reference][init-containers]. |
| apiSettings.labels | object | `{}` | Additional labels for the `teams-api` deployment. [Reference][labels-and-selectors]. |
| apiSettings.nodeSelector | object | `{}` | nodeSelector for teams-api. [Reference][node-selector]. |
| apiSettings.podAnnotations | object | `{}` | Annotations for pods for teams-api. [Reference][annotations]. |
| apiSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-api. [Reference][security-context]. |
| apiSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-api. [Reference][resources]. |
| apiSettings.securityContext | object | `{}` | Container security configuration for teams-api. [Reference][container-security-context]. |
| apiSettings.service.annotations | object | `{}` | Service annotations for teams-api. [Reference][annotations]. |
| apiSettings.service.containerPort | int | `8000` | Service container port for teams-api. |
| apiSettings.service.name | string | `"teams-api"` | Service name. |
| apiSettings.service.nodePort | int | `nil` | Service nodePort set only when `apiSettings.service.type: NodePort` for teams-api. |
| apiSettings.service.port | int | `80` | Service port for teams-api. |
| apiSettings.service.shortname | string | `"teams-api"` | Port name (maximum length is 15 characters) for teams-api. [Reference][ports]. |
| apiSettings.service.startup.failureThreshold | int | `5` | Number of times to retry the startup probe for the teams-api. [Reference][probes]. |
| apiSettings.service.startup.periodSeconds | int | `15` | How often (in seconds) to perform the startup probe for teams-api. [Reference][probes]. |
| apiSettings.service.type | string | `"ClusterIP"` | Service type for teams-api. [Reference][service-type]. |
| apiSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule pods with matching taints for teams-api. [Reference][taints-and-tolerations]. |
| apiSettings.topologySpreadConstraints | list | `[]` | Control how Pods are spread across your distributed footprint. Label selectors will be defaulted to those of the teams-api deployment. [Reference][topology-spread-constraints]. |
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
| appSettings.env.FIFTYONE_SIGNED_URL_EXPIRATION | int | `24` | Set the time-to-live for signed URLs generated by the application in hours |
| appSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. [Reference][image-pull-policy]. |
| appSettings.image.repository | string | `"voxel51/fiftyone-app"` | Container image for fiftyone-app. |
| appSettings.image.tag | string | `""` | Image tag for fiftyone-app. Defaults to the chart version. |
| appSettings.initContainers.enabled | bool | `true` | Whether to enable init containers for fiftyone-app. [Reference][init-containers]. |
| appSettings.initContainers.image.repository | string | `"docker.io/busybox"` | Init container images repositories for fiftyone-app. [Reference][init-containers]. |
| appSettings.initContainers.image.tag | string | `"stable-glibc"` | Init container images tags for fiftyone-app. [Reference][init-containers]. |
| appSettings.labels | object | `{}` | Additional labels for the `fiftyone-app` deployment. [Reference][labels-and-selectors]. |
| appSettings.nodeSelector | object | `{}` | nodeSelector for fiftyone-app. [Reference][node-selector]. |
| appSettings.podAnnotations | object | `{}` | Annotations for pods for fiftyone-app. [Reference][annotations]. |
| appSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for fiftyone-app. [Reference][security-context]. |
| appSettings.replicaCount | int | `2` | Number of pods in the fiftyone-app deployment's ReplicaSet. Ignored when `appSettings.autoscaling.enabled: true`. [Reference][deployment]. |
| appSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for fiftyone-app. [Reference][resources]. |
| appSettings.securityContext | object | `{}` | Container security configuration for fiftyone-app. [Reference][container-security-context]. |
| appSettings.service.annotations | object | `{}` | Service annotations for fiftyone-app. [Reference][annotations]. |
| appSettings.service.containerPort | int | `5151` | Service container port for fiftyone-app. |
| appSettings.service.name | string | `"fiftyone-app"` | Service name. |
| appSettings.service.nodePort | int | `nil` | Service nodePort set only when `appSettings.service.type: NodePort` for fiftyone-app. |
| appSettings.service.port | int | `80` | Service port. |
| appSettings.service.shortname | string | `"fiftyone-app"` | Port name (maximum length is 15 characters) for fiftyone-app. [Reference][ports]. |
| appSettings.service.startup.failureThreshold | int | `5` | Number of times to retry the startup probe for the fiftyone-app. [Reference][probes]. |
| appSettings.service.startup.periodSeconds | int | `15` | How often (in seconds) to perform the startup probe for fiftyone-app. [Reference][probes]. |
| appSettings.service.type | string | `"ClusterIP"` | Service type for fiftyone-app. [Reference][service-type]. |
| appSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule fiftyone-app pods with matching taints. [Reference][taints-and-tolerations]. |
| appSettings.topologySpreadConstraints | list | `[]` | Control how Pods are spread across your distributed footprint. Label selectors will be defaulted to those of the fiftyone-app deployment. [Reference][topology-spread-constraints]. |
| appSettings.volumeMounts | list | `[]` | Volume mounts for fiftyone-app. [Reference][volumes]. |
| appSettings.volumes | list | `[]` | Volumes for fiftyone-app. [Reference][volumes]. |
| casSettings.affinity | object | `{}` | Affinity and anti-affinity for teams-cas. [Reference][affinity]. |
| casSettings.enable_invitations | bool | `true` | Allow ADMINs to invite users by email NOTE: This is currently not supported when `FIFTYONE_AUTH_MODE: internal` |
| casSettings.env.CAS_DATABASE_NAME | string | `"cas"` | Provide the name for the CAS database. When multiple deployments use the same database instance, set `CAS_DATABASE_NAME` to a unique value for each deployment. |
| casSettings.env.CAS_DEFAULT_USER_ROLE | string | `"GUEST"` | Set the default user role for new users One of `GUEST`, `COLLABORATOR`, `MEMBER`, `ADMIN` |
| casSettings.env.CAS_MONGODB_URI_KEY | string | `"mongodbConnectionString"` | The key from `secret.fiftyone.name` that contains the CAS MongoDB Connection String. |
| casSettings.env.DEBUG | string | `"cas:*,-cas:*:debug"` | Set the log level for CAS examples: `DEBUG: cas:*` - shows all CAS logs `DEBUG: cas:*:info` - shows all CAS INFO logs `DEBUG: cas:*,-cas:*:debug` - shows all CAS logs except DEBUG logs |
| casSettings.env.FIFTYONE_AUTH_MODE | string | `"legacy"` | Configure Authentication Mode. One of `legacy` or `internal` |
| casSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. [Reference][image-pull-policy]. |
| casSettings.image.repository | string | `"voxel51/fiftyone-teams-cas"` | Container image for teams-cas. |
| casSettings.image.tag | string | `""` | Image tag for teams-cas. Defaults to the chart version. |
| casSettings.initContainers.enabled | bool | `true` | Whether to enable init containers for teams-cas. [Reference][init-containers]. |
| casSettings.initContainers.image.repository | string | `"docker.io/busybox"` | Init container images repositories for teams-cas. [Reference][init-containers]. |
| casSettings.initContainers.image.tag | string | `"stable-glibc"` | Init container images tags for teams-cas. [Reference][init-containers]. |
| casSettings.labels | object | `{}` | Additional labels for the `teams-cas` deployment. [Reference][labels-and-selectors]. |
| casSettings.nodeSelector | object | `{}` | nodeSelector for teams-cas. [Reference][node-selector]. |
| casSettings.podAnnotations | object | `{}` | Annotations for pods for teams-cas. [Reference][annotations]. |
| casSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-cas. [Reference][security-context]. |
| casSettings.replicaCount | int | `2` | Number of pods in the teams-cas deployment's ReplicaSet. [Reference][deployment]. |
| casSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-cas. [Reference][resources]. |
| casSettings.securityContext | object | `{}` | Container security configuration for teams-cas. [Reference][container-security-context]. |
| casSettings.service.annotations | object | `{}` | Service annotations for teams-cas. [Reference][annotations]. |
| casSettings.service.containerPort | int | `3000` | Service container port for teams-cas. |
| casSettings.service.name | string | `"teams-cas"` | Service name. |
| casSettings.service.nodePort | int | `nil` | Service nodePort set only when `casSettings.service.type: NodePort` for teams-cas. |
| casSettings.service.port | int | `80` | Service port. |
| casSettings.service.shortname | string | `"teams-cas"` | Port name (maximum length is 15 characters) for teams-cas. [Reference][ports]. |
| casSettings.service.startup.failureThreshold | int | `5` | Number of times to retry the startup probe for the teams-cas. [Reference][probes]. |
| casSettings.service.startup.periodSeconds | int | `15` | How often (in seconds) to perform the startup probe for teams-cas. [Reference][probes]. |
| casSettings.service.type | string | `"ClusterIP"` | Service type for teams-cas. [Reference][service-type]. |
| casSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule teams-cas pods with matching taints. [Reference][taints-and-tolerations]. |
| casSettings.topologySpreadConstraints | list | `[]` | Control how Pods are spread across your distributed footprint. Label selectors will be defaulted to those of the teams-cas deployment. [Reference][topology-spread-constraints]. |
| casSettings.volumeMounts | list | `[]` | Volume mounts for teams-cas. [Reference][volumes]. |
| casSettings.volumes | list | `[]` | Volumes for teams-cas. [Reference][volumes]. |
| delegatedOperatorExecutorSettings.affinity | object | `{}` | Affinity and anti-affinity for delegated-operator-executor. [Reference][affinity]. |
| delegatedOperatorExecutorSettings.enabled | bool | `false` | Controls whether to create a dedicated "teams-do" deployment. Disabled by default, meaning delegated operations will not be executed without an external executor system. |
| delegatedOperatorExecutorSettings.env.FIFTYONE_DELEGATED_OPERATION_RUN_LINK_PATH | string | `""` | Full path to a network-mounted file system or a cloud storage path to use for storing logs generated by delegated operation runs, one file per job. The default `""` means log upload is disabled. |
| delegatedOperatorExecutorSettings.env.FIFTYONE_INTERNAL_SERVICE | bool | `true` | Whether the SDK is running in an internal service context. When running in FiftyOne Teams, set to `true`. |
| delegatedOperatorExecutorSettings.env.FIFTYONE_MEDIA_CACHE_SIZE_BYTES | int | `-1` | Set the media cache size (in bytes) for the local FiftyOne Delegated Operator Executor processes. The default value is 32 GiB. `-1` is disabled. |
| delegatedOperatorExecutorSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. [Reference][image-pull-policy]. |
| delegatedOperatorExecutorSettings.image.repository | string | `"voxel51/fiftyone-app"` | Container image for delegated-operator-executor. |
| delegatedOperatorExecutorSettings.image.tag | string | `""` | Image tag for delegated-operator-executor. Defaults to the chart version. |
| delegatedOperatorExecutorSettings.labels | object | `{}` | Additional labels for the `delegated-operator-executor` deployment. [Reference][labels-and-selectors]. |
| delegatedOperatorExecutorSettings.liveness.failureThreshold | int | `5` | Number of times to retry the liveness probe for the teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.liveness.periodSeconds | int | `30` | How often (in seconds) to perform the liveness probe for teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.liveness.timeoutSeconds | int | `30` | Timeout for the liveness probe for the teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.name | string | `"teams-do"` | Deployment name |
| delegatedOperatorExecutorSettings.nodeSelector | object | `{}` | nodeSelector for delegated-operator-executor. [Reference][node-selector]. |
| delegatedOperatorExecutorSettings.podAnnotations | object | `{}` | Annotations for delegated-operator-executor pods. [Reference][annotations]. |
| delegatedOperatorExecutorSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for delegated-operator-executor. [Reference][security-context]. |
| delegatedOperatorExecutorSettings.readiness | object | `{"failureThreshold":5,"periodSeconds":30,"timeoutSeconds":30}` | Container security configuration for delegated-operator-executor. [Reference][container-security-context]. |
| delegatedOperatorExecutorSettings.readiness.failureThreshold | int | `5` | Number of times to retry the readiness probe for the teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.readiness.periodSeconds | int | `30` | How often (in seconds) to perform the readiness probe for teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.readiness.timeoutSeconds | int | `30` | Timeout for the readiness probe for the teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.replicaCount | int | `3` | Number of pods in the delegated-operator-executor deployment's ReplicaSet. This should not exceed the value set in the deployment's license file for  max concurrent delegated operators, which defaults to 3. |
| delegatedOperatorExecutorSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for delegated-operator-executor. [Reference][resources]. |
| delegatedOperatorExecutorSettings.securityContext | object | `{}` |  |
| delegatedOperatorExecutorSettings.startup.failureThreshold | int | `5` | Number of times to retry the startup probe for the teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.startup.periodSeconds | int | `30` | How often (in seconds) to perform the startup probe for teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.startup.timeoutSeconds | int | `30` | Timeout for the startup probe for the teams-do. [Reference][probes]. |
| delegatedOperatorExecutorSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule delegated-operator-executor pods with matching taints. [Reference][taints-and-tolerations]. |
| delegatedOperatorExecutorSettings.volumeMounts | list | `[]` | Volume mounts for delegated-operator-executor pods. [Reference][volumes]. |
| delegatedOperatorExecutorSettings.volumes | list | `[]` | Volumes for delegated-operator-executor. [Reference][volumes]. |
| fiftyoneLicenseSecrets | list | `["fiftyone-license"]` | List of secrets for FiftyOne Teams Licenses (one per org) |
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
| pluginsSettings.initContainers.enabled | bool | `true` | Whether to enable init containers for teams-plugins. [Reference][init-containers]. |
| pluginsSettings.initContainers.image.repository | string | `"docker.io/busybox"` | Init container images repositories for teams-plugins. [Reference][init-containers]. |
| pluginsSettings.initContainers.image.tag | string | `"stable-glibc"` | Init container images tags for teams-plugins. [Reference][init-containers]. |
| pluginsSettings.labels | object | `{}` | Additional labels for the `teams-plugins` deployment. [Reference][labels-and-selectors]. |
| pluginsSettings.nodeSelector | object | `{}` | nodeSelector for teams-plugins. [Reference][node-selector]. |
| pluginsSettings.podAnnotations | object | `{}` | Annotations for teams-plugins pods. [Reference][annotations]. |
| pluginsSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-plugins. [Reference][security-context]. |
| pluginsSettings.replicaCount | int | `2` | Number of pods in the teams-plugins deployment's ReplicaSet. Ignored when `pluginsSettings.autoscaling.enabled: true`. [Reference][deployment]. |
| pluginsSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-plugins. [Reference][resources]. |
| pluginsSettings.securityContext | object | `{}` | Container security configuration for teams-plugins. [Reference][container-security-context]. |
| pluginsSettings.service.annotations | object | `{}` | Service annotations for teams-plugins. [Reference][annotations]. |
| pluginsSettings.service.containerPort | int | `5151` | Service container port for teams-plugins. |
| pluginsSettings.service.name | string | `"teams-plugins"` | Service name. |
| pluginsSettings.service.nodePort | int | `nil` | Service nodePort set only when `pluginsSettings.service.type: NodePort` for teams-plugins. |
| pluginsSettings.service.port | int | `80` | Service port. |
| pluginsSettings.service.shortname | string | `"teams-plugins"` | Port name (maximum length is 15 characters) for teams-plugins. [Reference][ports]. |
| pluginsSettings.service.startup.failureThreshold | int | `5` | Number of times to retry the startup probe for the teams-plugins. [Reference][probes]. |
| pluginsSettings.service.startup.periodSeconds | int | `15` | How often (in seconds) to perform the startup probe for teams-plugins. [Reference][probes]. |
| pluginsSettings.service.type | string | `"ClusterIP"` | Service type for teams-plugins. [Reference][service-type]. |
| pluginsSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule teams-plugins pods with matching taints. [Reference][taints-and-tolerations]. |
| pluginsSettings.topologySpreadConstraints | list | `[]` | Control how Pods are spread across your distributed footprint. Label selectors will be defaulted to those of the teams-plugins deployment. [Reference][topology-spread-constraints]. |
| pluginsSettings.volumeMounts | list | `[]` | Volume mounts for teams-plugins pods. [Reference][volumes]. |
| pluginsSettings.volumes | list | `[]` | Volumes for teams-plugins. [Reference][volumes]. |
| secret.create | bool | `true` | Controls whether to create the secret named `secret.name`. |
| secret.fiftyone.cookieSecret | string | `""` | A randomly generated string for cookie encryption. To generate, run `openssl rand -hex 32`. |
| secret.fiftyone.encryptionKey | string | `""` | Encryption key for storage credentials. [Reference][fiftyone-encryption-key]. |
| secret.fiftyone.fiftyoneAuthSecret | string | `""` | A randomly generated string for CAS Authentication. This can be any string you care to use generated by any mechanism you   prefer. This is used for inter-service authentication and for the SuperUser to  authenticate at the CAS UI to configure the Central Authentication Service. |
| secret.fiftyone.fiftyoneDatabaseName | string | `""` | MongoDB Database Name for FiftyOne Teams. |
| secret.fiftyone.mongodbConnectionString | string | `""` | MongoDB Connection String. [Reference][mongodb-connection-string]. |
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
| teamsAppSettings.dnsName | string | `""` | DNS Name for the teams-app service. Used in the chart managed ingress (`spec.tls.hosts` and `spec.rules[0].host`) |
| teamsAppSettings.env.APP_USE_HTTPS | bool | `true` | Controls the protocol of the teams-app. Configure your ingress to match. When `true`, uses the https protocol. When `false`, uses the http protocol. |
| teamsAppSettings.env.FIFTYONE_APP_ALLOW_MEDIA_EXPORT | bool | `true` | When `false`, disables media export options |
| teamsAppSettings.env.FIFTYONE_APP_ANONYMOUS_ANALYTICS_ENABLED | bool | `true` | Controls whether anonymous analytics are captured for the teams application. Set to false to opt-out of anonymous analytics. |
| teamsAppSettings.env.FIFTYONE_APP_DEFAULT_QUERY_PERFORMANCE | bool | `true` | Controls whether Query Performance mode is enabled by default for every dataset for the teams application. Set to false to set default mode to off. |
| teamsAppSettings.env.FIFTYONE_APP_ENABLE_QUERY_PERFORMANCE | bool | `true` | Controls whether Query Performance mode is enabled for the teams application. Set to false to disable Query Performance mode for entire application. |
| teamsAppSettings.env.FIFTYONE_APP_TEAMS_SDK_RECOMMENDED_VERSION | string | `"2.2.0"` | The recommended fiftyone SDK version that will be displayed in the install modal (i.e. `pip install ... fiftyone==0.11.0`). |
| teamsAppSettings.env.FIFTYONE_APP_THEME | string | `"dark"` | The default theme configuration. `dark`: Theme will be dark when user visits for the first time. `light`: Theme will be light theme when user visits for the first time. `always-dark`: Sets dark theme on each refresh (overrides user theme changes in the app). `always-light`: Sets light theme on each refresh (overrides user theme changes in the app). |
| teamsAppSettings.env.RECOIL_DUPLICATE_ATOM_KEY_CHECKING_ENABLED | bool | `false` | Disable duplicate atom/selector key checking that generated false-positive errors. [Reference][recoil-env]. |
| teamsAppSettings.fiftyoneApiOverride | string | `""` | Overrides the `FIFTYONE_API_URI` environment variable. When set `FIFTYONE_API_URI` controls the value shown in the API Key Modal providing guidance for connecting to the FiftyOne Teams API. `FIFTYONE_API_URI` uses the value from apiSettings.dnsName if it is set, or uses the teamsAppSettings.dnsName |
| teamsAppSettings.image.pullPolicy | string | `"Always"` | Instruct when the kubelet should pull (download) the specified image. One of `IfNotPresent`, `Always` or `Never`. Reference][image-pull-policy]. |
| teamsAppSettings.image.repository | string | `"voxel51/fiftyone-teams-app"` | Container image for teams-app. |
| teamsAppSettings.image.tag | string | `""` | Image tag for teams-app. Defaults to the chart version. |
| teamsAppSettings.initContainers.enabled | bool | `true` | Whether to enable init containers for teams-app. [Reference][init-containers]. |
| teamsAppSettings.initContainers.image.repository | string | `"docker.io/busybox"` | Init container images repositories for teams-app. [Reference][init-containers]. |
| teamsAppSettings.initContainers.image.tag | string | `"stable-glibc"` | Init container images tags for teams-app. [Reference][init-containers]. |
| teamsAppSettings.labels | object | `{}` | Additional labels for the `teams-app` deployment. [Reference][labels-and-selectors]. |
| teamsAppSettings.nodeSelector | object | `{}` | nodeSelector for teams-app. [Reference][node-selector]. |
| teamsAppSettings.podAnnotations | object | `{}` | Annotations for teams-app pods. [Reference][annotations]. |
| teamsAppSettings.podSecurityContext | object | `{}` | Pod-level security attributes and common container settings for teams-app. [Reference][security-context]. |
| teamsAppSettings.replicaCount | int | `2` | Number of pods in the teams-app deployment's ReplicaSet. Ignored when `teamsAppSettings.autoscaling.enabled: true`. [Reference][deployment]. |
| teamsAppSettings.resources | object | `{"limits":{},"requests":{}}` | Container resource requests and limits for teams-app. [Reference][resources]. |
| teamsAppSettings.securityContext | object | `{}` | Container security configuration for teams-app. [Reference][container-security-context]. |
| teamsAppSettings.service.annotations | object | `{}` | Service annotations for teams-app. [Reference][annotations]. |
| teamsAppSettings.service.containerPort | int | `3000` | Service container port for teams-app. |
| teamsAppSettings.service.name | string | `"teams-app"` | Service name. |
| teamsAppSettings.service.nodePort | int | `nil` | Service nodePort set only when `teamsAppSettings.service.type: NodePort` for teams-app. |
| teamsAppSettings.service.port | int | `80` | Service port. |
| teamsAppSettings.service.shortname | string | `"teams-app"` | Port name (maximum length is 15 characters) for teams-app. [Reference][ports]. |
| teamsAppSettings.service.startup.failureThreshold | int | `5` | Number of times to retry the startup probe for the teams-app. [Reference][probes]. |
| teamsAppSettings.service.startup.periodSeconds | int | `15` | How often (in seconds) to perform the startup probe for teams-app. [Reference][probes]. |
| teamsAppSettings.service.type | string | `"ClusterIP"` | Service type for teams-app. [Reference][service-type]. |
| teamsAppSettings.tolerations | list | `[]` | Allow the k8s scheduler to schedule teams-app pods with matching taints. [Reference][taints-and-tolerations]. |
| teamsAppSettings.topologySpreadConstraints | list | `[]` | Control how Pods are spread across your distributed footprint. Label selectors will be defaulted to those of the teams-app deployment. [Reference][topology-spread-constraints]. |
| teamsAppSettings.volumeMounts | list | `[]` | Volume mounts for teams-app pods. [Reference][volumes]. |
| teamsAppSettings.volumes | list | `[]` | Volumes for teams-app pods. [Reference][volumes]. |

## Upgrading From Previous Versions

Voxel51 assumes you use the published
Helm Chart to deploy your FiftyOne Teams environment.
If you are using a custom deployment
mechanism, carefully review the changes in the
[Helm Chart](https://github.com/voxel51/fiftyone-teams-app-deploy)
and update your deployment accordingly.

### From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success
team member to coordinate this upgrade.
You will need to either create a new Identity Provider (IdP)
or modify your existing configuration to migrate to a new Auth0 Tenant.

### From Before FiftyOne Teams Version 1.1.0

> **NOTE**: Upgrading from versions of FiftyOne Teams prior to v1.1.0
> requires upgrading the database and will interrupt all SDK connections.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted Web App.
> You should coordinate this upgrade carefully with your end-users.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
> from your `values.yaml` and from any secrets created outside of the Voxel51
> install process.

---

1. In your `values.yaml`, set the required values
    1. `secret.fiftyone.encryptionKey` (or your deployment's equivalent)
        1. This sets the `FIFTYONE_ENCRYPTION_KEY` environment variable
           in the appropriate service pods
    1. `secret.fiftyone.fiftyoneAuthSecret` (or your deployment's equivalent)
        1. This sets the `FIFTYONE_AUTH_SECRET` environment variable
           in the appropriate service pods
    1. `appSettings.env.FIFTYONE_DATABASE_ADMIN: true`
        1. This is not the default value in the Helm Chart and must be overridden
    1. When using path-based routing, update your ingress with the rule
       (add it before the `path: /` rule)

        ```yaml
        ingress:
            paths:
              - path: /cas
                pathType: Prefix
                serviceName: teams-cas
                servicePort: 80
        ```

1. Use the license file provided by the Voxel51 Customer Success Team to create
   a new kubernetes secret:

    ```shell
    kubectl --namespace your-namepace-here create secret generic \
        fiftyone-license --from-file=license=./your-license-file
    ```

1. [Upgrade to FiftyOne Teams v2.2.0](#deploying-fiftyone-teams)
    > **NOTE**: At this step, FiftyOne SDK users will lose access to the
    > FiftyOne Teams Database until they upgrade to `fiftyone==2.2.0`
1. Upgrade your FiftyOne SDKs to version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated
      with your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. Validate that all datasets are now at version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Versions After 1.1.0 and Before Version 1.6.0

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_
> your users to log in after the upgrade is complete.
> This will interrupt active workflows in the FiftyOne Teams Hosted
> Web App. You should coordinate this upgrade carefully with your
> end-users.

---

> **NOTE**: FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS).
> CAS requires additional configurations and consumes additional resources.
> Please review the upgrade instructions, the
> [Central Authentication Service](#central-authentication-service)
> documentation and the
> [Pluggable Authentication](https://docs.voxel51.com/teams/pluggable_auth.html)
> documentation before completing your upgrade.

---

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
> from your `values.yaml` and from any secrets created outside of the Voxel51
> install process.

---

1. Ensure all FiftyOne SDK users either
    - Set the `FIFTYONE_DATABASE_ADMIN` to `false`

      ```shell
      FIFTYONE_DATABASE_ADMIN=false
      ```

    - Unset the environment variable `FIFTYONE_DATABASE_ADMIN`
      (this should generally be your default)

        ```shell
        unset FIFTYONE_DATABASE_ADMIN
        ```

1. Use the license file provided by the Voxel51 Customer Success Team to create
   a new kubernetes secret:

    ```shell
    kubectl --namespace your-namepace-here create secret generic \
        fiftyone-license --from-file=license=./your-license-file
    ```

1. In your `values.yaml`, set the required values
    1. `secret.fiftyone.encryptionKey` (or your deployment's
       equivalent)
        1. This sets the `FIFTYONE_ENCRYPTION_KEY` environment variable
           in the appropriate service pods
    1. `secret.fiftyone.fiftyoneAuthSecret` (or your deployment's equivalent)
        1. This sets the `FIFTYONE_AUTH_SECRET` environment variable
           in the appropriate service pods
1. [Upgrade to FiftyOne Teams version 2.2.0](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    > **NOTE** Any FiftyOne SDK less than 2.2.0 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.2.0` is recommended before migrating
        > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. Validate that all datasets are now at version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Versions 1.6.0 to 1.7.1

> **NOTE**: Upgrading to FiftyOne Teams v2.2.0 _requires_ a license file.
> Please contact your Customer Success Team before upgrading to FiftyOne Teams
> 2.0 or beyond.
>
> The license file now contains all of the Auth0 configuration that was
> previously provided through kubernetes secrets; you may remove those secrets
> from your `values.yaml` and from any secrets created outside of the Voxel51
> install process.

---

> **NOTE**: If you had previously set
> `teamsAppSettings.env.FIFTYONE_APP_INSTALL_FIFTYONE_OVERRIDE` to include your
> Voxel51 private PyPI token, you can remove it from your configuration. The
> Voxel51 private PyPI token is now loaded correctly from your license file.

---

1. Ensure all FiftyOne SDK users either
    - Set the `FIFTYONE_DATABASE_ADMIN` to `false`

      ```shell
      FIFTYONE_DATABASE_ADMIN=false
      ```

    - Unset the environment variable `FIFTYONE_DATABASE_ADMIN`
      (this should generally be your default)

        ```shell
        unset FIFTYONE_DATABASE_ADMIN
        ```

1. Use the license file provided by the Voxel51 Customer Success Team to create
   a new kubernetes secret:

    ```shell
    kubectl --namespace your-namepace-here create secret generic \
        fiftyone-license --from-file=license=./your-license-file
    ```

1. [Upgrade to FiftyOne Teams version 2.2.0](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 2.2.0
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Upgrade all the datasets

    > **NOTE** Any FiftyOne SDK less than 2.2.0 will lose connectivity after
    > this point.
    > Upgrading all SDKs to `fiftyone==2.2.0` is recommended before migrating
        > your database.

    ```shell
    FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
    ```

1. Validate that all datasets are now at version 0.25.1

    ```shell
    fiftyone migrate --info
    ```

### From FiftyOne Teams Version 2.0.0

1. [Upgrade to FiftyOne Teams version 2.2.0](#deploying-fiftyone-teams)
1. Voxel51 recommends upgrading all FiftyOne Teams SDK users to FiftyOne Teams
   version 2.2.0, but it is not required
    - Login to the FiftyOne Teams UI
    - To obtain the CLI command to install the FiftyOne SDK associated with
      your FiftyOne Teams version, navigate to `Account > Install FiftyOne`
1. Voxel51 recommends that you upgrade all your datasets, but it is not
   required.  Users using the FiftyOne Teams 2.0.0 SDK will continue to operate
   uninterrupted during, and after, this migration

   ```shell
   FIFTYONE_DATABASE_ADMIN=true fiftyone migrate --all
   ```

1. To ensure that all datasets are now at version 0.25.1, run

   ```shell
   fiftyone migrate --info
   ```

## Deploying FiftyOne Teams

A minimal example `values.yaml` may be found
[here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/values.yaml).

1. Edit the `values.yaml` file
1. Deploy FiftyOne Teams with `helm install`
    1. For a new installation
        1. Create a new namespace and set the current namespace for your kubectl
           context

           ```shell
           kubectl create namespace your-namespace-here
           kubectl config set-context --current --namespace your-namespace-here
           ```

        1. If you are using the Voxel51 DockerHub registry to install your
           container images, use the Voxel51-provided DockerHub credentials to
           create an Image Pull Secret, and uncomment the `imagePullSecrets`
           section of your `values.yaml`

           ```shell
           kubectl --namespace your-namespace-here create secret generic \
           regcred --from-file=.dockerconfigjson=./voxel51-docker.json \
           --type kubernetes.io/dockerconfigjson
           ```

        1. Use your Voxel51-provided License file to create a FiftyOne License
           Secret

           ```shell
           kubectl --namespace your-namepace-here create secret generic \
           fiftyone-license --from-file=license=./your-license-file
           ```

        1. Add the Voxel51 Helm repository and install FiftyOne Teams

           ```shell
           helm repo add voxel51 https://helm.fiftyone.ai
           helm repo update voxel51
           helm install fiftyone-teams-app voxel51/fiftyone-teams-app \
           -f ./values.yaml
           ```

    1. To upgrade an existing helm installation

        1. Make sure you have followed the appropriate directions for
           [Upgrading From Previous Versions](#upgrading-from-previous-versions)

        1. Update your kubectl configuration to set your current namespace for
           your kubectl context

           ```shell
           kubectl config set-context --current --namespace your-namespace-here
           ```

        1. Update your Voxel51 Helm repository and upgrade your FiftyOne Teams
           deployment

           ```shell
           helm repo update voxel51
           helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
           -f ./values.yaml
           ```

        > **NOTE**  To view the changes Helm would apply during installations
        > and upgrades, consider using
        > [helm diff](https://github.com/databus23/helm-diff).
        > Voxel51 is not affiliated with the author of this plugin.
        >
        > For example:
        >
        > ```shell
        > helm diff -C1 upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
        > ```

### Deploying On GKE

Voxel51 FiftyOne Teams supports
[Workload Identity Federation for GKE][about-wif]
when installing via Helm into Google Kubernetes Engine (GKE).
Workload Identity is achieved using service account annotations
that can be defined in the `values.yaml` file when installing
or upgrading the application.

Please follow the steps
[outlined by Google][howto-wif]
to allow your cluster to utilize workload identity federation and to
create a service account with the required IAM permissions.

Once the cluster and service account are configured, you can permit your
workloads to utilize the GCP service account via service account annotations
defined in the `values.yaml` file:

```yaml
serviceAccount:
  annotations:
    iam.gke.io/gcp-service-account: <GSA_NAME>@<GSA_PROJECT>.iam.gserviceaccount.com
```

<!-- Reference Links -->
[about-wif]: https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity
[affinity]: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/
[annotations]: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
[autoscaling]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[container-security-context]: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container
[deployment]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
[howto-wif]: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
[image-pull-policy]: https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy
[image-pull-secrets]: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
[ingress-default-ingress-class]: https://kubernetes.io/docs/concepts/services-networking/ingress/#default-ingress-class
[ingress-rules]: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules
[ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[init-containers]: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
[internal-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#internal-mode
[labels-and-selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
[legacy-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#legacy-mode
[mongodb-connection-string]: https://www.mongodb.com/docs/manual/reference/connection-string/
[node-selector]: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
[plugins-storage]: https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/docs/plugins-storage.md
[ports]: https://kubernetes.io/docs/concepts/services-networking/service/#field-spec-ports
[probes]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
[recoil-env]: https://recoiljs.org/docs/api-reference/core/RecoilEnv/
[resources]: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
[security-context]: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
[service-account]: https://kubernetes.io/docs/concepts/security/service-accounts/
[service-type]: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
[taints-and-tolerations]: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
[topology-spread-constraints]: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
[volumes]: https://kubernetes.io/docs/concepts/storage/volumes/
