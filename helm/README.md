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

- [:green_book: Prerequisites Skills and Knowledge](#green_book-prerequisites-skills-and-knowledge)
- [:white_check_mark: Technical Requirements](#white_check_mark-technical-requirements)
  - [Kubernetes Cluster And Kubectl](#kubernetes-cluster-and-kubectl)
  - [Helm](#helm)
- [:clock10: Estimated Completion Time](#clock10-estimated-completion-time)
- [:floppy_disk: Sizing](#floppy_disk-sizing)
- [:wrench: Step 1: Set Up MongoDB Database](#wrench-step-1-set-up-mongodb-database)
- [:closed_lock_with_key: Step 2: Prepare License File](#closed_lock_with_key-step-2-prepare-license-file)
- [:file_folder: Step 3: Choose Authentication Mode](#file_folder-step-3-choose-authentication-mode)
- [:gear: Step 4: Configure `values.yaml`](#gear-step-4-configure-valuesyaml)
- [:file_cabinet: Step 5: Enable Shared Storage](#file_cabinet-step-5-enable-shared-storage)
- [:jigsaw: Step 6: Enable Dedicated Plugins Mode](#jigsaw-step-6-enable-dedicated-plugins-mode)
- [:robot: Step 7: Configure Delegated Operators](#robot-step-7-configure-delegated-operators)
- [:rocket: Step 8: Initial Deployment](#rocket-step-8-initial-deployment)
- [:globe_with_meridians: Step 9: Configure Ingress & TLS](#globe_with_meridians-step-9-configure-ingress--tls)
  - [:compass: Routing Overview (Path-Based Ingress)](#compass-routing-overview-path-based-ingress)
  - [:memo: Notes](#memo-notes)
- [Step 10: Identity Provider (IdP) and Authentication (CAS)](#step-10-identity-provider-idp-and-authentication-cas)
- [Step 11: Initial CAS Setup](#step-11-initial-cas-setup)
  - [Add First Admin User](#add-first-admin-user)
  - [Enable Auto Join](#enable-auto-join)
- [Step 12: Test End User Login](#step-12-test-end-user-login)
- [Recommended Enhancements](#recommended-enhancements)
- [Upgrades](#upgrades)
- [Known Issues](#known-issues)
- [Advanced Configuration](#advanced-configuration)
- [Validating](#validating)
- [Health Checks and Monitoring](#health-checks-and-monitoring)
- [Values](#values)

<!-- tocstop -->

## :green_book: Prerequisites Skills and Knowledge

A successful, properly secured deployment of FiftyOne Enterprise requires
knowledge of Kubernetes, Helm, MongoDB, DNS, and TLS/SSL certificate
management. See the chart's
[Prerequisites Skills and Knowledge](./fiftyone-teams-app/README.md#prerequisites-skills-and-knowledge)
for the full list.

## :white_check_mark: Technical Requirements

The following technical requirements are required for a successful and
properly secured deployment of FiftyOne Enterprise.

1. [Kubernetes Cluster And Kubectl](#kubernetes-cluster-and-kubectl)

1. [Helm](#helm)

1. A MongoDB Database that meets FiftyOne's
   [version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).

1. A DNS record or records for ingress.

1. A TLS/SSL certificate or certificates for HTTPS ingress.

1. (optional) An NFS server or `ReadWriteMany` compatible storage medium for
   [delegated operators](./fiftyone-teams-app/README.md#builtin-delegated-operator-orchestrator),
   [plugins](./fiftyone-teams-app/README.md#plugins),
   and
   [API high-availability](./fiftyone-teams-app/README.md#highly-available-fiftyone-teams-api-deployments)

### Kubernetes Cluster And Kubectl

A kubernetes cluster and `kubectl` installation are required.
The following kubernetes/kubectl versions are required:

Kubernetes: `>=1.31-0`

However, it is recommended to use a
[supported kubernetes version](https://kubernetes.io/releases/).
Please refer to the
[kubernetes installation documentation](https://kubernetes.io/docs/tasks/tools/)
for steps on installing kubernetes and kubectl.

### Helm

Helm version >= 3.14 is required.

Please refer to the
[helm installation documentation](https://helm.sh/docs/intro/install/)
for steps on installing helm.

## :clock10: Estimated Completion Time

The estimated time to deploy FiftyOne Enterprise is approximately 2 hours.

## :floppy_disk: Sizing

Voxel51 recommends the following resource sizing:

- MongoDB: 4 CPU, 16GB RAM, 256GB Storage
- FiftyOne App: 1 CPU, 6GB RAM, 1GB Storage per pod
- Teams API: 1 CPU, 2GB RAM, 1GB Storage per pod
- Teams App: 500 mCPU, 512MB RAM, 512MB Storage per pod
- Teams CAS: 500 mCPU, 512MB RAM, 512MB Storage per pod
- Delegated Operators: 8 CPU, 16GB RAM, 1GB Storage per pod

Voxel51 also recommends monitoring resource consumption across
the applications.
Resource usage varies dramatically with operations, use cases,
and dataset sizes.

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

## :file_cabinet: Step 5: Enable Shared Storage

Dedicated plugins and delegated operators (below) share a common plugin
directory backed by a Kubernetes PersistentVolume (PV) and
PersistentVolumeClaim (PVC).

```yaml
# teams-plugins-pv-pvc.yaml
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: teams-plugins-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
    - ReadOnlyMany
  nfs:
    server: nfs-server
    path: "/fiftyone_teams_app/plugins"

---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: teams-plugins-pvc
spec:
  accessModes:
    - ReadWriteMany
    - ReadOnlyMany
  storageClassName: ""
  resources:
    requests:
      storage: 10Gi
```

```shell
kubectl apply -f teams-plugins-pv-pvc.yaml
```

For NFS export configuration and cloud-provider alternatives (Google
Filestore, AWS EFS, Azure Files), see
[Adding Shared Storage for FiftyOne Enterprise Plugins](./docs/plugins-storage.md).

## :jigsaw: Step 6: Enable Dedicated Plugins Mode

Add the following to your `values.yaml` to run plugins in a dedicated
`teams-plugins` pod, isolated from `fiftyone-app`:

```yaml
pluginsSettings:
  enabled: true
  env:
    FIFTYONE_PLUGINS_DIR: /opt/plugins
  volumes:
    - name: plugins-vol
      persistentVolumeClaim:
        claimName: teams-plugins-pvc
        readOnly: true
  volumeMounts:
    - name: plugins-vol
      mountPath: /opt/plugins

# teams-api also requires read-write access to the plugins directory
apiSettings:
  env:
    FIFTYONE_PLUGINS_DIR: /opt/plugins
  volumes:
    - name: plugins-vol
      persistentVolumeClaim:
        claimName: teams-plugins-pvc
  volumeMounts:
    - name: plugins-vol
      mountPath: /opt/plugins
```

For full configuration options, see
[Configuring Plugins](./docs/configuring-plugins.md).

## :robot: Step 7: Configure Delegated Operators

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
            claimName: teams-plugins-pvc
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

## :rocket: Step 8: Initial Deployment

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

Confirm all pods are running, including `teams-plugins` (from
[Step 6](#jigsaw-step-6-enable-dedicated-plugins-mode)) and your chosen
delegated operator workers (from
[Step 7](#robot-step-7-configure-delegated-operators)):

```shell
kubectl get pods --namespace your-namespace-here
```

## :globe_with_meridians: Step 9: Configure Ingress & TLS

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

## Step 10: Identity Provider (IdP) and Authentication (CAS)

FiftyOne Enterprise uses a Central Authentication Service (CAS), introduced
in v1.6, for centralized login, roles, and user management. You chose your
authentication mode in
[Step 3](#file_folder-step-3-choose-authentication-mode).

The CAS service requires the following in your `values.yaml`:

- `secret.fiftyone.fiftyoneAuthSecret` (set in
  [Step 4](#gear-step-4-configure-valuesyaml))
- When using path-based routing, an ingress rule for `/cas` routed to
  `teams-cas` (see
  [Step 9](#globe_with_meridians-step-9-configure-ingress--tls))

For more detail, see the chart's
[Central Authentication Service](./fiftyone-teams-app/README.md#central-authentication-service)
documentation and the
[Pluggable Authentication](https://docs.voxel51.com/enterprise/pluggable_auth.html)
docs.

## Step 11: Initial CAS Setup

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

## Step 12: Test End User Login

Verify the deployment's IdP setup by logging in as a regular user.

1. In a browser, open `https://<ENVIRONMENT>.fiftyone.ai`.
1. Log in with the user credentials of the admin you created in the CAS
   Super Admin UI.
1. Confirm you are redirected to the FiftyOne Enterprise home page.

## Recommended Enhancements

With dedicated plugins and delegated operators configured in
[Step 6](#jigsaw-step-6-enable-dedicated-plugins-mode) and
[Step 7](#robot-step-7-configure-delegated-operators), consider these
additional enhancements for a production-ready deployment:

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
