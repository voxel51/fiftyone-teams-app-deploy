<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">
</p>
</div>

---

FiftyOne Teams is the enterprise version of the open source [FiftyOne](https://github.com/voxel51/fiftyone) project.

Please contact [Voxel51](https://voxel51.com/#teams-form) if you would like more information regarding Fiftyone Teams.

# Deploying FiftyOne Teams App Using Helm

The `fiftyone-teams-app`, `fiftyone-teams-api`, and `fiftyone-app` images are avaialable via Docker Hub, with the appropriate credentials.  If you do not have Docker Hub credentials for the `voxel51` repositories, please contact your support team for Docker Hub credentials.

---

## Initial Installation vs. Upgrades

`FIFTYONE_DATABASE_ADMIN` is set to `false` by default for FiftyOne Teams version 1.2.1 upgrades and installations.   This is because FiftyOne Teams version 1.2.1 is backwards compatible with FiftyOne Teams database schema 0.19 (Teams Version 1.1).

- If you are performing an initial install, you will either want to connect to your MongoDB database with the 0.12.0 SDK before performing the FiftyOne Teams installation, or you will want to add `FIFTYONE_DATABASE_ADMIN: true` in the `env` section of the `appSettings` configuration.

- If you are performing an upgrade, please review our [Upgrade Process Recommendations](#upgrade-process-recommendations)

---

## Notes and Considerations

While not all parameters are required, Voxel51 frequently sees deployments use the following parameters:

	imagePullSecrets
	ingress.annotations

Please consider if you will require these settings for your deployment.

---

### FiftyOne Teams Upgrade Notes

#### Storage Credentials and `FIFTYONE_ENCRYPTION_KEY`

Containers based on the `fiftyone-teams-api` and `fiftyone-app` images now _REQUIRE_ the inclusion of the `FIFTYONE_ENCRYPTION_KEY` variable.  This key is used to encrypt storage credentials in the MongoDB database.

The `encryptionKey` secret can be generated using the following python:

```
from cryptography.fernet import Fernet
print(Fernet.generate_key().decode())
```

Voxel51 does not have access to this encryption key and cannot reproduce it.  If this key is lost you will need to schedule an outage window to drop the storage credentials collection, replace the encryption key, and add the storage credentials via the UI again.  Voxel51 strongly recommends storing this key in a safe place.

Storage credentials no longer need to be mounted into containers with appropriate environment variables being set; users with `Admin` permissions can use `/settings/cloud_storage_credentials` in the Web UI to add supported storage credentials.

FiftyOne Teams version 1.1.1 continues to support the use of environment variables to set storage credentials in the application context but is providing an alternate configuration path for future functionality.

#### Environment Proxies

FiftyOne Teams version 1.1.1 supports routing traffic through proxy servers; this can be configured by setting the following environment variables on all containers in the environment (`*.env`):

```
http_proxy: http://proxy.yourcompany.tld:3128
https_proxy: https://proxy.yourcompany.tld:3128
no_proxy: <apiSettings.service.name>, <appSettings.service.name>, <teamsAppSettings.service.name>
HTTP_PROXY: http://proxy.yourcompany.tld:3128
HTTPS_PROXY: https://proxy.yourcompany.tld:3128
NO_PROXY: <apiSettings.service.name>, <appSettings.service.name>, <teamsAppSettings.service.name>

```

You must also set the following environment variables on containers based on the `fiftyone-teams-app` image (`teamsAppSettings.env`):

```
GLOBAL_AGENT_HTTP_PROXY: http://proxy.yourcompany.tld:3128
GLOBAL_AGENT_HTTPS_PROXY: https://proxy.yourconpay.tld:3128
GLOBAL_AGENT_NO_PROXY: <apiSettings.service.name>, <appSettings.service.name>, <teamsAppSettings.service.name>
```

The `NO_PROXY` and `GLOBAL_AGENT_NO_PROXY` values must include the names of the kubernetes services to allow FiftyOne Teams services to talk to each other without going through a proxy server.  By default these service names are `teams-api`, `teams-app`, and `fiftyone-app` but may have been changed using the `service.name` parameter for each service.

By default the Global Agent Proxy will log all outbound connections and identify which connections are routed through the proxy.  You can reduce the verbosity of the logging output by adding the following environment variable to your `teamsAppSettings.env`:

```
ROARR_LOG: false
```

#### Text Similarity

FiftyOne Teams now supports using text similarity searches for images that are indexed with a model that [supports text queries](https://docs.voxel51.com/user_guide/brain.html#brain-similarity-text).  If you choose to make use of this feature, you must use the `fiftyone-app-torch` image provided by Voxel51 instead of the `fiftyone-app` image.

You can override the default image by providing a new `appSettings.image.repository` value to the Helm Chart.  Using the included `values.yaml` this configuration might look like:

```
appSettings:
  image:
    repository: voxel51/fiftyone-app-torch
```

---

## Required Helm Chart Values


| Required Values                           | Default | Description                                 |
|-------------------------------------------|---------|---------------------------------------------|
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

| Optional Values                                                  | Default                    | Description                                                               |
|------------------------------------------------------------------|----------------------------|---------------------------------------------------------------------------|
| `apiSettings.affinity`                                           | None                       | FiftyOne Teams API service affinity rules                                 |
| `apiSettings.env`                                                | defined below              | Arbitrary environment variables to pass to the FiftyOne Teams API pod     |
| `apiSettings.env.FIFTYONE_ENV`                                   | production                 | Verbosity for GraphQL query output                                        |
| `apiSettings.env.GRAPHQL_DEFAULT_LIMIT`                          | 10                         | Default GraphQL limit for results                                         |
| `apiSettings.env.LOGGING_LEVEL`                                  | INFO                       | Logging Verbosity                                                         |
| `apiSettings.image.repository`                                   | voxel51/fiftyone-teams-api | Container Image for FiftyOne Teams API containers                         |
| `apiSettings.image.tag`                                          | Helm Chart Version         | Container Image Tag for FiftyOne Teams API containers                     |
| `apiSettings.nodeSelector`                                       | None                       | FiftyOne Teams API pod node selector rules                                |
| `apiSettings.podAnnotations`                                     | None                       | FiftyOne Teams API pod annotation rules                                   |
| `apiSettings.podSecurityContext`                                 | None                       | FiftyOne Teams API pod security context rules                             |
| `apiSettings.resources.limits.cpu`                               | None                       | CPU resource limits for FiftyOne Teams API containers                     |
| `apiSettings.resources.limits.memory`                            | None                       | Memory resource limits for FiftyOne Teams API containers                  |
| `apiSettings.resources.requests.cpu`                             | None                       | CPU resource requests for FiftyOne Teams API containers                   |
| `apiSettings.resources.requests.memory`                          | None                       | Memory resource requests for FiftyOne Teams API containers                |
| `apiSettings.securityContext`                                    | None                       | FiftyOne Teams API service security context rules                         |
| `apiSettings.service.containerPort`                              | 8000                       | Port for FiftyOne Teams API Containers                                    |
| `apiSettings.service.liveness.initialDelaySeconds`               | 45                         | Delay before Liveness checks for FiftyOne Teams API containers            |
| `apiSettings.service.name`                                       | teams-api                  | FiftyOne Teams API service name                                           |
| `apiSettings.service.nodePort`                                   | None                       | `nodePort` for Service Type `NodePort`                                    |
| `apiSettings.service.port`                                       | 80                         | FiftyOne Teams API service port                                           |
| `apiSettings.service.readiness.initialDelaySeconds`              | 45                         | Delay before Readiness checks for FiftyOne Teams API containers           |
| `apiSettings.service.shortname`                                  | teams-api                  | Short name for port definitions (less than 15 characters)                 |
| `apiSettings.service.type`                                       | ClusterIP                  | FiftyOne Teams API service type                                           |
| `apiSettings.tolerations`                                        | None                       | FiftyOne Teams API service toleration rules                               |
| `apiSettings.volumeMounts`                                       | None                       | FiftyOne Teams API pod volume mount definitions                           |
| `apiSettings.volumes`                                            | None                       | FiftyOne Teams API pod volume definitions                                 |
| `appSettings.affinity`                                           | None                       | FiftyOne App service affinity rules                                       |
| `appSettings.autoscaling.enabled`                                | false                      | Enable Horizontal Autoscaling for the FiftyOne App Pod                    |
| `appSettings.autoscaling.maxReplicas`                            | 20                         | Maximum Replicas for Horizontal Autoscaling in the FiftyOne App pod       |
| `appSettings.autoscaling.minReplicas`                            | 2                          | Minimum Replicas for Horizontal Autoscaling in the FiftyOne App pod       |
| `appSettings.autoscaling.targetCPUUtilizationPercentage`         | 80                         | Percent CPU Utilization for autoscaling the FiftyOne App pod              |
| `appSettings.autoscaling.targetMemoryUtilizationPercentage`      | 80                         | Percent Memory Utilization for autoscaling the FiftyOne App pod           |
| `appSettings.env`                                                | defined below              | Arbitrary environment variables to pass to the FiftyOne App pod           |
| `appSettings.env.FIFTYONE_DATABASE_ADMIN`                        | false                      | Toggles MongoDB database admin privileges for the FiftyOne App pod        |
| `appSettings.env.FIFTYONE_MEDIA_CACHE_IMAGES`                    | false                      | Toggle image caching for the local FiftyOne App processes                 |
| `appSettings.env.FIFTYONE_MEDIA_CACHE_SIZE_BYTES`                | -1 (disabled)              | Set the media cache size for the local FiftyOne App processes             |
| `appSettings.image.repository`                                   | voxel51/fiftyone-app       | Container Image for FiftyOne App containers                               |
| `appSettings.image.tag`                                          | Helm Chart Version         | Container Image tag for FiftyOne App containers                           |
| `appSettings.nodeSelector`                                       | None                       | FiftyOne App pod node selector rules                                      |
| `appSettings.podAnnotations`                                     | None                       | FiftyOne App pod annotation rules                                         |
| `appSettings.podSecurityContext`                                 | None                       | FiftyOne App pod security context rules                                   |
| `appSettings.replicaCount`                                       | 2                          | FiftyOne App replica count if autoscaling is disabled                     |
| `appSettings.resources.limits.cpu`                               | None                       | CPU resource limits for FiftyOne App containers                           |
| `appSettings.resources.limits.memory`                            | None                       | Memory resource limits for FiftyOne App containers                        |
| `appSettings.resources.requests.cpu`                             | None                       | CPU resource requests for FiftyOne App containers                         |
| `appSettings.resources.requests.memory`                          | None                       | Memory resource requests for FiftyOne App containers                      |
| `appSettings.securityContext`                                    | None                       | FiftyOne App service security context rules                               |
| `appSettings.service.containerPort`                              | 5151                       | Port for FiftyOne App Containers                                          |
| `appSettings.service.liveness.initialDelaySeconds`               | 45                         | Delay before Liveness checks for FiftyOne App containers                  |
| `appSettings.service.name`                                       | fiftyone-app               | FiftyOne App service name                                                 |
| `appSettings.service.nodePort`                                   | None                       | `nodePort` for Service Type `NodePort`                                    |
| `appSettings.service.port`                                       | 80                         | FiftyOne App service port                                                 |
| `appSettings.service.readiness.initialDelaySeconds`              | 45                         | Delay before Readiness checks for FiftyOne App containers                 |
| `appSettings.service.shortname`                                  | fiftyone-app               | Shirt name for port definitions (less than 15 characters)                 |
| `appSettings.service.type`                                       | ClusterIP                  | FiftyOne App service type                                                 |
| `appSettings.tolerations`                                        | None                       | FiftyOne App service toleration rules                                     |
| `appSettings.volumeMounts`                                       | None                       | FiftyOne App pod volume mount definitions                                 |
| `appSettings.volumes`                                            | None                       | FiftyOne App pod volume definitions                                       |
| `ingress.annotations`                                            | None                       | Ingress annotations (if required)                                         |
| `ingress.className`                                              | ""                         | Ingress class name (if required)                                          |
| `ingress.enabled`                                                | true                       | Toggle enabling ingress                                                   |
| `ingress.paths`                                                  | See Below                  | List of ingress `path` and `pathType`                                     |
| `ingress.paths.path`                                             | `/*`                       | path to associate with the FiftyOne Teams App service                     |
| `ingress.paths.path.pathType`                                    | ImplementationSpecific     | Ingress path type (`ImplementationSpecific`, `Exact`, `Prefix`)           |
| `ingress.tlsEnabled`                                             | true                       | Enable TLS for Ingress Controller                                         |
| `ingress.tlsSecretName`                                          | fiftyone-teams-tls-secret  | TLS Secret for certificate with all three DNS Names                       |
| `namespace.name`                                                 | fiftyone-teams             | Kubernetes Namespace already created for FiftyOne Teams                   |
| `secret.create`                                                  | true                       | Toggle creation of the FiftyOne secret by Helm                            |
| `secret.name`                                                    | fiftyone-teams-secrets     | Name for the FiftyOne Teams configuration secrets                         |
| `secret.fiftyone.fiftyoneDatabaseName`                           | fiftyone                   | MongoDB Database Name for FiftyOne Teams                                  |
| `serviceAccount.annotations`                                     | None                       | Service account annotations                                               |
| `serviceAccount.create`                                          | true                       | Toggle creation of a service account for the FiftyOne Teams deployment    |
| `serviceAccount.name`                                            | fiftyone-teams             | Service account name                                                      |
| `teamsAppSettings.affinity`                                      | None                       | FiftyOne Teams App service affinity rules                                 |
| `teamsAppSettings.autoscaling.enabled`                           | false                      | Enable Horizontal Autoscaling for the FiftyOne Teams App Pod              |
| `teamsAppSettings.autoscaling.maxReplicas`                       | 20                         | Maximum Replicas for Horizontal Autoscaling in the FiftyOne Teams App pod |
| `teamsAppSettings.autoscaling.minReplicas`                       | 2                          | Minimum Replicas for Horizontal Autoscaling in the FiftyOne Teams App pod |
| `teamsAppSettings.autoscaling.targetCPUUtilizationPercentage`    | 80                         | Percent CPU Utilization for autoscaling the FiftyOne Teams App pod        |
| `teamsAppSettings.autoscaling.targetMemoryUtilizationPercentage` | 80                         | Percent Memory Utilization for autoscaling the FiftyOne Teams App pod     |
| `teamsAppSettings.env`                                           | defined below              | Arbitrary environment variables to pass to the FiftyOne Teams App pod     |
| `teamsAppSettings.env.APP_USE_HTTPS`                             | true                       | Set to `false` if Ingress does not use HTTPS                              |
| `teamsAppSettings.image.repository`                              | voxel51/fiftyone-teams-app | Container Image for FiftyOne Teams App containers                         |
| `teamsAppSettings.image.tag`                                     | Helm Chart Version         | Container Image tag for FiftyOne Teams App containers                     |
| `teamsAppSettings.nodeSelector`                                  | None                       | FiftyOne Teams App pod node selector rules                                |
| `teamsAppSettings.podAnnotations`                                | None                       | FiftyOne Teams App pod annotation rules                                   |
| `teamsAppSettings.podSecurityContext`                            | None                       | FiftyOne Teams App pod security context rules                             |
| `teamsAppSettings.replicaCount`                                  | 2                          | FiftyOne Teams App replica count if autoscaling is disabled               |
| `teamsAppSettings.resources.limits.cpu`                          | None                       | CPU resource limits for FiftyOne Teams App containers                     |
| `teamsAppSettings.resources.limits.memory`                       | None                       | Memory resource limits for FiftyOne Teams App containers                  |
| `teamsAppSettings.resources.requests.cpu`                        | None                       | CPU resource requests for FiftyOne Teams App containers                   |
| `teamsAppSettings.resources.requests.memory`                     | None                       | Memory resource requests for FiftyOne Teams App containers                |
| `teamsAppSettings.securityContext`                               | None                       | FiftyOne Teams App service security context rules                         |
| `teamsAppSettings.serverPathPrefix`                              | `/`                        | FiftyOne App prefix for path-based Ingress routing                        |
| `teamsAppSettings.service.containerPort`                         | 3000                       | Port for FiftyOne Teams App containers                                    |
| `teamsAppSettings.service.liveness.initialDelaySeconds`          | 45                         | Delay before Liveness checks for FiftyOne Teams App containers            |
| `teamsAppSettings.service.name`                                  | teams-app                  | FiftyOne Teams App service name                                           |
| `teamsAppSettings.service.nodePort`                              | None                       | `nodePort` for Service Type `NodePort`                                    |
| `teamsAppSettings.service.port`                                  | 80                         | FiftyOne Teams App service port                                           |
| `teamsAppSettings.service.readiness.initialDelaySeconds`         | 45                         | Delay before Readiness checks for FiftyOne Teams App containers           |
| `teamsAppSettings.service.shortname`                             | teams-app                  | Short name for port definitions (less than 15 characters)                 |
| `teamsAppSettings.service.type`                                  | ClusterIP                  | FiftyOne Teams App service type                                           |
| `teamsAppSettings.tolerations`                                   | None                       | FiftyOne Teams App service toleration rules                               |
| `teamsAppSettings.volumeMounts`                                  | None                       | FiftyOne Teams App pod volume mount definitions                           |
| `teamsAppSettings.volumes`                                       | None                       | FiftyOne Teams App pod volume definitions                                 |

---

## Upgrade Process Recommendations

### Upgrade Process Recommendations From Early Adopter Versions (Versions less than 1.0)

Please contact your Voxel51 Customer Success team member to coordinate this upgrade.  You will need to either create a new IdP or modify your existing configuration in order to migrate to a new Auth0 Tenant.

### Upgrade Process Recommendations From Before FiftyOne Teams Version 1.1.0

The FiftyOne 0.12.0 SDK (database version 0.20.0) is _NOT_ backwards-compatible with FiftyOne Teams Database Versions prior to 0.19.0, and the FiftyOne 0.10 SDK is not forwards compatible with current FiftyOne Teams Database Versions.  If you are using a FiftyOne SDK older than 0.11.0, upgrading the Web server will require upgrading all FiftyOne SDK installations.

Voxel51 recommends the following upgrade process for upgrading from versions prior to FiftyOne Teams version 1.1.0:

1. Make sure your installation includes the required [FIFTYONE_ENCRYPTION_KEY](#fiftyone-teams-upgrade-notes) environment variable
1. [Upgrade to FiftyOne Teams version 1.2.1](#deploying-fiftyone-teams) with `appSettings.env.FIFTYONE_DATABASE_ADMIN: true` (this is not the default in the Helm Chart for this release).<br>
    **NOTE:** FiftyOne SDK users will lose access to the FiftyOne Teams Database at this step until they upgrade to `fiftyone==0.12.0`
1. Upgrade your FiftyOne SDKs to version 0.12.0<br>
    The command line for installing the FiftyOne SDK associated with your FiftyOne Teams version is available in the FiftyOne Teams UI under `Account > Install FiftyOne` after a user has logged in.
1. Use `fiftyone migrate --info` to make sure that all datasets have been migrated to version 0.20.0.
    - If not all datasets have been upgraded, have an admin set `FIFTYONE_DATABASE_ADMIN=true` in their local environment
	- Have that admin use `fiftyone migrate --all` to upgrade any remaining datasets

### Upgrade Process Recommendations From FiftyOne Teams Version 1.1.0 and later

The FiftyOne 0.12.0 SDK (database version 0.20.0) is backwards-compatible with FiftyOne Teams Database Versions after 0.19.0, but the FiftyOne 0.11.0 SDK is _NOT_ forwards compatible with FiftyOne Teams Database Version 0.20.0.

Voxel51 always recommends using the latest version of the FiftyOne SDK compatible with your FiftyOne Teams deployment.

Voxel51 recommends the following upgrade process for upgrading from FiftyOne Teams version 1.1.0 or later:

1. Ensure all FiftyOne SDK users set `FIFTYONE_DATABASE_ADMIN=false` or `unset FIFTYONE_DATABASE_ADMIN` (this should generally be your default)
1. [Upgrade to FiftyOne Teams version 1.2.1](#deploying-fiftyone-teams)
1. Upgrade FiftyOne Teams SDK users to FiftyOne Teams version 0.12.0<br>
    The command line for installing the FiftyOne SDK associated with your FiftyOne Teams version is available in the FiftyOne Teams UI under `Account > Install FiftyOne` after a user has logged in.
1. Have an admin set `FIFTYONE_DATABASE_ADMIN=true` in their local environment
1. Have the admin run `fiftyone migrate --all` to upgrade all datasets
1. Use `fiftyone migrate --info` to ensure that all datasets are now at version 0.20.0

---

## Deploying FiftyOne Teams

You can find an example, minimal, `values.yaml` [here](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/values.yaml).

The author of this document recommends the use of [helm diff](https://github.com/databus23/helm-diff) to determine what changes will be applied during installations and upgrades.  Voxel51 is not affiliated with the author of this plugin.

Once you have edited the `values.yaml` file you can deploy your FiftyOne Teams instance with:
```
helm repo add voxel51 https://helm.fiftyone.ai
helm repo update voxel51
helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
```

You can upgrade an existing deployment with:
```
helm repo update voxel51
helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
```

---

## A Full GKE Deployment Example

The following instructions represent a full Google Kubernetes Engine [GKE] deployment using:
- The jetstack/cert-manager Helm chart for Let's Encrypt SSL certificates
- The bitnami/mongodb Helm chart for MongoDB
- The voxel51/fiftyone-teams-app Helm chart

These instructions assume you have [kubectl](https://kubernetes.io/docs/tasks/tools/) and [Helm](https://helm.sh/docs/intro/install/) installed and operating, and that you have an existing [GKE Cluster available](https://cloud.google.com/kubernetes-engine/docs/concepts/kubernetes-engine-overview).

These instructions assume you have received Docker Hub credentials from Voxel51 and have placed your `voxel51-docker.json` file in the current directory; if your `voxel51-docker.json` is not in the current directory please update the command line accordingly.

These instructions assume you have received your Auth0 configuration information from Voxel51.  If you have not received this information, please contact your [Voxel51 Support Team](mailto:support@voxel51.com).

### Download the Example Configuration Files

Download the example configuration files from the [Voxel51 GitHub](https://github.com/voxel51/fiftyone-teams-app-deploy/helm/gke-examples) repository.


One way to do this might be:
```
curl -o values.yaml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/values.yaml
curl -o clusterissuer.yml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/clusterissuer.yml
curl -o frontendconfig.yml https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/gke-example/frontendconfig.yml
```

You will need to edit the `values.yaml` file to include the Auth0 configuration provided by Voxel51, your MongoDB username and password, to set a `cookieSecret`, to set an `encryptionKey` value, and insert your `host` values (search for `replace.this.dns.name`).

Assuming you follow these directions your MongoDB host will be `fiftyone-mongodb.fiftyone-mongodb.svc.cluster.local`; please modify that hostname if you modify these instructions.

### Create the Necessary Helm Repos

Add the jetstack, bitnami, and voxel51 Helm repositories to your local configuration:
```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add jetstack https://charts.jetstack.io
helm repo add voxel51 https://helm.fiftyone.ai
helm repo update
```

### Install and Configure cert-manager

If you are using a GKE Autopilot cluster, please review the information [provided by cert-manager](https://github.com/cert-manager/cert-manager/issues/3717#issuecomment-919299192) and adjust your installation accordingly.

```
kubectl create namespace cert-manager
kubectl config set-context --current --namespace cert-manager
helm install cert-manager jetstack/cert-manager --set installCRDs=true
```

You can use the cert-manager instructions to [verify the cert-manager Installation](https://cert-manager.io/v1.4-docs/installation/verify/).

### Create a ClusterIssuer
`ClusterIssuers` are Kubernetes resources that represent certificate authorities that are able to generate signed certificates by honoring certificate signing requests.  You must create either an `Issuer` in each namespace or a `ClusterIssuer` as part of your cert-manager configuration.  Voxel51 has provided an example `ClusterIssuer` configuration (downloaded [earlier](#download-the-example-configuration-files) in this guide).

```
kubectl apply -f ./clusterissuer.yml
```

### Install and Configure MongoDB

These instructions deploy a single-node MongoDB instance in your GKE cluster.  If you would like to deploy MongoDB with a replicaset configuration, please refer to the [MongoDB Helm Chart](https://github.com/bitnami/charts/tree/master/bitnami/mongodb) documentation.

**You will definitely want to edit the `rootUser` and `rootPassword` defined below.**

```
kubectl create namespace fiftyone-mongodb
kubectl config set-context --current --namespace fiftyone-mongodb
helm install fiftyone-mongodb \
    --set auth.rootPassword=REPLACEME \
    --set auth.rootUser=admin \
    --set global.namespaceOverride=fiftyone-mongodb \
    --set image.tag=4.4 \
    --set ingress.enabled=true \
    --set namespaceOverride=fiftyone-mongodb \
    bitnami/mongodb
```

Wait until the MongoDB pods are in the `Ready` state before beginning the "Install FiftyOne Teams App" instructions.

You should [configure a DNS entry](#obtain-a-global-static-ip-address-and-configure-a-dns-entry) while you wait.

You can use `kubectl get pods` to determine the state of the `fiftyone-mongodb` pods.

### Obtain a Global Static IP Address and Configure a DNS Entry

Reserve a global static IP address for use in your cluster:

```
gcloud compute addresses create fiftyone-teams-static-ip --global --ip-version IPV4
gcloud compute addresses describe fiftyone-teams-static-ip --global
```

Record the IP address and either create a DNS entry or contact your Voxel51 support team to have them create an appropriate `fiftyone.ai` DNS entry for you.

### Set up http to https forwarding
```
kubectl apply -f frontendconfig.yml
```

### Install FiftyOne Teams App

```
kubectl create namespace fiftyone-teams
kubectl config set-context --current --namespace fiftyone-teams
kubectl create secret generic regcred \
    --from-file=.dockerconfigjson=./voxel51-docker.json \
    --type kubernetes.io/dockerconfigjson
helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
```

Issuing SSL Certificates can take up to 15 minutes; be patient while Let's Encrypt and GKE negotiate.

You can verify that your SSL certificates have been properly issued with the following curl command:

`curl -I https://replace.this.dns.name`

Your SSL certificates have been correctly issued if you see `HTTP/2 200` at the top of the response.  If, however, you encounter a `SSL certificate problem: unable to get local issuer certificate` message you should delete the certificate and allow it to recreate.

`kubectl delete secret fiftyone-teams-cert-secret`

Further instructions for debugging ACME certificates are on the [cert-manager docs site](https://cert-manager.io/docs/faq/acme/).

Once your installation is complete, browse to `/settings/cloud_storage_credentials` and add your storage credentials to access sample data.

### Installation Complete

Congratulations! You should now be able to access your FiftyOne Teams installation at the DNS address you created [earlier](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).

