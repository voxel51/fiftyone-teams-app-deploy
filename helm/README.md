<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img alt="Voxel51 Logo" src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px">
&nbsp;
<img alt="Voxel51 FiftyOne" src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# FiftyOne Enterprise: Helm Deployment Guide

FiftyOne Enterprise is the enterprise version of the open source
[FiftyOne](https://github.com/voxel51/fiftyone) project.

The FiftyOne Enterprise Helm chart is the recommended way to install and
configure FiftyOne Enterprise on Kubernetes.

This guide walks you through the steps for installing FiftyOne Enterprise
using Helm. It also includes advanced configuration and upgrade
considerations. This page assumes general knowledge of FiftyOne Enterprise
and how to use it. Please contact Voxel51 for more information regarding
FiftyOne Enterprise.

## Table of Contents

<!-- toc -->

- [:open_file_folder: Directory Contents](#open_file_folder-directory-contents)
- [:green_book: Prerequisites Skills and Knowledge](#green_book-prerequisites-skills-and-knowledge)
- [:white_check_mark: Technical Requirements](#white_check_mark-technical-requirements)
- [:clock10: Estimated Completion Time](#clock10-estimated-completion-time)
- [:floppy_disk: Sizing](#floppy_disk-sizing)
- [:wrench: Step 1: Set Up MongoDB Database](#wrench-step-1-set-up-mongodb-database)
- [:closed_lock_with_key: Step 2: Prepare License File](#closed_lock_with_key-step-2-prepare-license-file)
- [:file_folder: Step 3: Choose Authentication Mode](#file_folder-step-3-choose-authentication-mode)
- [:gear: Step 4: Configure `values.yaml`](#gear-step-4-configure-valuesyaml)
  - [Enable Dedicated Plugins (required)](#enable-dedicated-plugins-required)
  - [Configure Delegated Operators (required)](#configure-delegated-operators-required)
- [:rocket: Step 5: Initial Deployment](#rocket-step-5-initial-deployment)
- [:globe_with_meridians: Step 6: Configure Ingress & TLS](#globe_with_meridians-step-6-configure-ingress--tls)
  - [:compass: Routing Overview (Path-Based Ingress)](#compass-routing-overview-path-based-ingress)
  - [:memo: Notes](#memo-notes)
- [:page_facing_up: Step 7: Configure Delegated Operation Logs](#page_facing_up-step-7-configure-delegated-operation-logs)
- [Step 8: Identity Provider (IdP) and Authentication (CAS)](#step-8-identity-provider-idp-and-authentication-cas)
- [Step 9: Initial CAS Setup](#step-9-initial-cas-setup)
  - [Add First Admin User](#add-first-admin-user)
  - [Enable Auto Join](#enable-auto-join)
- [Step 10: Test End User Login](#step-10-test-end-user-login)
- [:books: Full Worked Example](#books-full-worked-example)
- [Recommended Enhancements](#recommended-enhancements)
- [Upgrades](#upgrades)
- [Known Issues](#known-issues)
- [Advanced Configuration](#advanced-configuration)
- [Validating](#validating)
- [Health Checks and Monitoring](#health-checks-and-monitoring)
- [Values](#values)

<!-- tocstop -->

## :open_file_folder: Directory Contents

This directory contains resources and information related to Helm deployments.

- Directories
  - `docs` contains additional documentation, including cloud-specific
    deployment guides and configuration references.
  - `fiftyone-teams-app` contains the Helm chart `voxel51/fiftyone-teams-app`.
    For the full chart documentation (prerequisites, sizing, advanced
    configuration, and the complete values reference), see
    [`fiftyone-teams-app/README.md`](./fiftyone-teams-app/README.md).
  - `gke-example` contains example Kubernetes resources used in the
    [GKE Deployment Guide](./docs/gke-deployment-guide.md).
  - `local-self-signed-example` contains resources for a local, self-signed
    development setup.
- Files
  - `values.yaml` is an example of overrides for the chart's defaults for a
    deployment.

## :green_book: Prerequisites Skills and Knowledge

A successful, properly secured deployment of FiftyOne Enterprise requires
knowledge of Kubernetes, Helm, MongoDB, DNS, and TLS/SSL certificate
management. See the chart's
[Prerequisites Skills and Knowledge](./fiftyone-teams-app/README.md#prerequisites-skills-and-knowledge)
for the full list.

## :white_check_mark: Technical Requirements

See the chart's
[Technical Requirements](./fiftyone-teams-app/README.md#technical-requirements)
for the full list of required tooling and versions (Kubernetes, `kubectl`,
Helm, MongoDB, DNS, TLS/SSL, and optionally NFS/`ReadWriteMany` storage).

## :clock10: Estimated Completion Time

The estimated time to deploy FiftyOne Enterprise is approximately 2 hours.
See the chart's
[Estimated Completion Time](./fiftyone-teams-app/README.md#estimated-completion-time)
section for details.

## :floppy_disk: Sizing

See the chart's
[Sizing](./fiftyone-teams-app/README.md#sizing)
recommendations for MongoDB, `fiftyone-app`, `teams-api`, `teams-app`,
`teams-cas`, dedicated plugins, and delegated operator pods.

## :wrench: Step 1: Set Up MongoDB Database

Before deploying FiftyOne Enterprise, you must have a running MongoDB database.
FiftyOne Enterprise supports:

- **MongoDB Atlas** (managed cloud)
- **MongoDB Community Edition** (self-hosted, open source)
- **MongoDB Enterprise** (self-hosted, commercial)

Ensure your MongoDB version meets FiftyOne's
[version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).
MongoDB 8.0+ is recommended.

Once your database is running, record your MongoDB connection URI.
You will need it in
[Step 4](#gear-step-4-configure-valuesyaml)
to set `secret.fiftyone.mongodbConnectionString` in your `values.yaml`.
The URI follows this format:

```text
mongodb://username:password@mongodb-example.fiftyone.ai:27017/?authSource=admin
```

## :closed_lock_with_key: Step 2: Prepare License File

> Required for **v2.0+**

Use the license file provided by the Voxel51 Customer Success Team to create
a license secret:

```shell
kubectl create namespace your-namespace-here
kubectl --namespace your-namespace-here create secret generic fiftyone-license \
  --from-file=license=./your-license-file
```

Set the name of this secret in your `values.yaml`'s `fiftyoneLicenseSecrets`
list (see the example [`values.yaml`](./values.yaml)).

> [!TIP]
> When rotating the license, to ensure that the new license values are
> picked up immediately, restart the `teams-cas` and `teams-api` deployments:
>
> ```shell
> kubectl rollout restart deploy \
>   -n your-namespace-here \
>   teams-cas \
>   teams-api
> ```

## :file_folder: Step 3: Choose Authentication Mode

FiftyOne Enterprise supports two authentication modes.
Choose the one that matches your deployment:

| Mode       | Use when                                                                   | `values.yaml` setting                          |
| ---------- | -------------------------------------------------------------------------- | ---------------------------------------------- |
| `internal` | Deployment is Air-gapped or Identity Provider is **OpenID Connect (OIDC)** | `casSettings.env.FIFTYONE_AUTH_MODE: internal` |
| `legacy`   | Identity Provider is **SAML**                                              | `casSettings.env.FIFTYONE_AUTH_MODE: legacy`   |

> [!NOTE]
> The example [`values.yaml`](./values.yaml) ships with
> `casSettings.env.FIFTYONE_AUTH_MODE: legacy` by default. Update it if your
> Identity Provider uses OIDC.

You can refer to the following docs to set up your Identity Provider with
FiftyOne:

- [Pluggable authentication docs](https://docs.voxel51.com/enterprise/pluggable_auth.html#pluggable-authentication)
  includes information on configuring CAS.
- To set up authentication for internal mode: refer to the
  [Getting Started with Internal Mode documentation](https://docs.voxel51.com/enterprise/pluggable_auth.html#getting-started-with-internal-mode).

## :gear: Step 4: Configure `values.yaml`

Edit your `values.yaml` file (see the example
[`values.yaml`](./values.yaml) in this directory) to set:

- `fiftyoneLicenseSecrets` — the name of the secret created in
  [Step 2](#closed_lock_with_key-step-2-prepare-license-file)
- `secret.fiftyone.mongodbConnectionString` — the MongoDB connection URI from
  [Step 1](#wrench-step-1-set-up-mongodb-database)
- `secret.fiftyone.cookieSecret` — a randomly generated string
  (`openssl rand -hex 32`)
- `secret.fiftyone.encryptionKey` — used to encrypt storage credentials; see
  [Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`](./fiftyone-teams-app/README.md#storage-credentials-and-fiftyone_encryption_key)
  for how to generate it
- `secret.fiftyone.fiftyoneAuthSecret` — a randomly generated shared secret
  used for CAS
- `casSettings.env.FIFTYONE_AUTH_MODE` — the mode chosen in
  [Step 3](#file_folder-step-3-choose-authentication-mode)
- `teamsAppSettings.dnsName` — your ingress hostname

If you are using the Voxel51 Docker Hub registry to pull container images,
create an image pull secret and reference it in `imagePullSecrets`:

```shell
kubectl --namespace your-namespace-here create secret generic regcred \
  --from-file=.dockerconfigjson=./voxel51-docker.json \
  --type kubernetes.io/dockerconfigjson
```

> [!NOTE]
> This chart uses `namespace.name` from your `values.yaml` to set the
> namespace for all chart resources. The Helm `--namespace` flag alone does
> **not** control where resources are created — see the chart's
> [Usage](./fiftyone-teams-app/README.md#usage) section for the full
> `namespace.name` configuration note.

For the full list of available settings, see
[Values](./fiftyone-teams-app/README.md#values).

### Enable Dedicated Plugins (required)

Dedicated plugins and delegated operators (below) share a common plugin
directory backed by a Kubernetes PersistentVolume (PV) and
PersistentVolumeClaim (PVC). If you do not already have shared storage
configured, see
[Adding Shared Storage for FiftyOne Enterprise Plugins](./docs/plugins-storage.md)
before continuing.

Add the following to your `values.yaml` to run plugins in a dedicated
`teams-plugins` pod, isolated from `fiftyone-app`:

```yaml
pluginsSettings:
  enabled: true
  env:
    FIFTYONE_PLUGINS_DIR: /opt/plugins

# teams-api also requires access to the plugins directory
apiSettings:
  env:
    FIFTYONE_PLUGINS_DIR: /opt/plugins
```

For full configuration options, see
[Configuring Plugins](./docs/configuring-plugins.md).

### Configure Delegated Operators (required)

Delegated operators allow long-running or compute-heavy tasks (computing
embeddings, model evaluation, dataset import, annotation workflows) to be
scheduled from the FiftyOne UI and executed in the background. Choose the
mode that fits your workload:

| Mode                    | Use when                                                                                | Configuration                                   |
| ----------------------- | --------------------------------------------------------------------------------------- | ----------------------------------------------- |
| **Always-On Workers**   | Steady or unpredictable delegated-operation volume; workers should be ready immediately | `delegatedOperatorDeployments` in `values.yaml` |
| **On-Demand Executors** | Infrequent or GPU-heavy jobs; avoid paying for idle workers                             | Kubernetes Jobs spun up per run                 |

For **Always-On Workers**, add the following to your `values.yaml`:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDoCpuDefault:
      enabled: true
      env:
        FIFTYONE_PLUGINS_DIR: /opt/plugins
      volumes:
        - name: plugins-vol
          persistentVolumeClaim:
            claimName: plugins-pvc
            readOnly: true
      volumeMounts:
        - name: plugins-vol
          mountPath: /opt/plugins
```

For **On-Demand Executors**, see
[Configuring On-Demand Orchestrator](../docs/configuring-on-demand-orchestrator.md)
for setup instructions.

For full always-on delegated operator configuration options, see
[Configuring Delegated Operators](./docs/configuring-delegated-operators.md).

## :rocket: Step 5: Initial Deployment

Add the Voxel51 Helm repository and install FiftyOne Enterprise:

```shell
helm repo add voxel51 https://helm.fiftyone.ai
helm repo update voxel51
helm install fiftyone-teams-app voxel51/fiftyone-teams-app \
  --namespace your-namespace-here \
  -f ./values.yaml
```

For upgrades, run:

```shell
helm repo update voxel51
helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app \
  --namespace your-namespace-here \
  -f ./values.yaml
```

> [!TIP]
> Prior to running `helm upgrade`, you may view the changes Helm would apply
> using the [helm diff](https://github.com/databus23/helm-diff) plugin
> (Voxel51 is not affiliated with the author of this plugin):
>
> ```shell
> helm diff --context 1 upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
> ```

Confirm all pods are running, including `teams-plugins` and your chosen
delegated operator workers, both configured in
[Step 4](#gear-step-4-configure-valuesyaml):

```shell
kubectl get pods --namespace your-namespace-here
```

## :globe_with_meridians: Step 6: Configure Ingress & TLS

Next, configure an **Ingress controller** and **TLS termination** in front of
your FiftyOne Enterprise services — for example, using
[cert-manager](https://cert-manager.io/) with your cloud provider's ingress
controller or load balancer.

### :compass: Routing Overview (Path-Based Ingress)

| Path                 | Proxied To  | Description                          |
| -------------------- | ----------- | ------------------------------------ |
| `/`                  | `teams-app` | Main web UI                          |
| `/cas`               | `teams-cas` | Central Authentication Service (CAS) |
| `/cloud_credentials` | `teams-api` | Cloud credentials API endpoint       |
| `/graphql/v1`        | `teams-api` | GraphQL API endpoint                 |
| `/rpc`               | `teams-api` | RPC API endpoint                     |
| `/file`              | `teams-api` | File import handling                 |
| `/_pymongo`          | `teams-api` | MongoDB requests via SDK             |
| `/health`            | `teams-api` | Health check endpoint                |

### :memo: Notes

- FiftyOne Enterprise supports routing traffic through proxy servers. Please
  refer to the
  [proxy configuration documentation](./docs/configuring-proxies.md) for
  information on how to configure proxies.
- To expose the `teams-api` for host-based routing, see
  [Exposing the `teams-api`](./docs/expose-teams-api.md).
- For a complete, cloud-specific worked example of ingress/TLS setup, see the
  [GKE Deployment Guide](./docs/gke-deployment-guide.md) or the
  [AWS Deployment Guide](./docs/aws-deployment-guide.md).

## :page_facing_up: Step 7: Configure Delegated Operation Logs

Add the log path for delegated operation runs to your `values.yaml`:

```yaml
delegatedOperatorDeployments:
  deployments:
    teamsDoCpuDefault:
      env:
        FIFTYONE_DELEGATED_OPERATION_LOG_PATH: "gs://your-bucket/logs"
```

Logs are stored in the format:

```text
/mnt/shared/logs/do_logs/<YYYY>/<MM>/<DD>/<RUN_ID>.log
```

This is useful for auditing, debugging, or monitoring delegated operator
executions in shared storage or cloud buckets.

## Step 8: Identity Provider (IdP) and Authentication (CAS)

FiftyOne Enterprise uses a Central Authentication Service (CAS), introduced
in v1.6, for centralized login, roles, and user management. You chose your
authentication mode in
[Step 3](#file_folder-step-3-choose-authentication-mode).

The CAS service requires the following in your `values.yaml`:

- `secret.fiftyone.fiftyoneAuthSecret` (set in
  [Step 4](#gear-step-4-configure-valuesyaml))
- When using path-based routing, an ingress rule for `/cas` routed to
  `teams-cas` (see
  [Step 6](#globe_with_meridians-step-6-configure-ingress--tls))

For more detail, see the chart's
[Central Authentication Service](./fiftyone-teams-app/README.md#central-authentication-service)
documentation and the
[Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
docs.

## Step 9: Initial CAS Setup

1. Navigate to the CAS Super Admin UI at
   `https://<ENVIRONMENT>.fiftyone.ai/cas/configurations`.
1. In the **API Key** field (upper right corner), enter the value of
   `secret.fiftyone.fiftyoneAuthSecret` (from your `values.yaml`).

### Add First Admin User

1. Navigate to the **Admins** tab at
   `https://<ENVIRONMENT>.fiftyone.ai/cas/admins`.
1. Select **Add admin**.
1. Provide **Name** and **Email**.
1. Select **Add**.

### Enable Auto Join

1. Navigate to `https://<ENVIRONMENT>.fiftyone.ai/cas/providers`.
1. Select **+ Edit**.
1. Select **Allow auto join**.
1. Select **Save**.

## Step 10: Test End User Login

Verify the deployment's IdP setup by logging in as a regular user.

1. In a browser, open `https://<ENVIRONMENT>.fiftyone.ai`.
1. Log in with the user credentials of the admin you created in the CAS
   Super Admin UI.
1. Confirm you are redirected to the FiftyOne Enterprise home page.

## :books: Full Worked Example

The generic steps above apply to any Kubernetes cluster. For a complete,
cloud-specific walkthrough that combines all of the steps above, see:

- [GKE Deployment Guide](./docs/gke-deployment-guide.md)
- [AWS Deployment Guide](./docs/aws-deployment-guide.md)

## Recommended Enhancements

With dedicated plugins and delegated operators configured in
[Step 4](#gear-step-4-configure-valuesyaml), consider these additional
enhancements for a production-ready deployment:

| Enhancement | What it enables |
| --- | --- |
| **GPU-Scheduled Workers** | *(Optional)* Schedule delegated operator pods on GPU-enabled nodes for compute-heavy tasks like embeddings or model evaluation |
| **Multiple Orchestrators** | *(Optional)* Register separate CPU- and GPU-targeted delegated operator orchestrators for mixed workloads |
| **Custom Job Priorities** | *(Advanced)* Use Kubernetes PriorityClasses with on-demand delegated operator Jobs |

For step-by-step configuration instructions, see
[Recommended Post-Installation Configuration](./docs/post-install-recommended-configuration.md).

## Upgrades

When performing an upgrade, please review
[Upgrading From Previous Versions](./docs/upgrading.md).

## Known Issues

For a list of common issues and their solutions, refer to the
[Known Issues documentation](./docs/known-issues.md).

If you encounter a new issue, please open a ticket on the
[GitHub Issues page](https://github.com/voxel51/fiftyone-teams-app-deploy/issues).

## Advanced Configuration

For backup and recovery, secrets management, GPU workloads, high
availability, plugins, proxies, snapshot archival, storage credentials,
static banners, Terms of Service URLs, text similarity, and workload identity
federation, see the chart's
[Advanced Configuration](./fiftyone-teams-app/README.md#advanced-configuration)
documentation.

## Validating

After deploying FiftyOne Enterprise and configuring authentication, please
follow
[Validating Your Deployment](../docs/validating-deployment.md).

## Health Checks and Monitoring

See the chart's
[Health Checks And Monitoring](./fiftyone-teams-app/README.md#health-checks-and-monitoring)
documentation for basic health assessment and troubleshooting unhealthy pods.

## Values

For the full list of configurable `values.yaml` settings, see the chart's
generated [Values](./fiftyone-teams-app/README.md#values) reference table.
