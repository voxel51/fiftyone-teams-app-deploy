<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

**Deploying FiftyOne Teams App Using Helm**
</p>
</div>

---

# Deploying FiftyOne Teams App Using Helm

Due to the number of elements that are unique in each deployment, there is no default `values.yaml` in the Helm chart.  You must provide all parameters to use this Chart.

To obtain the default `values.yaml` file, try:

```
curl -o values.yaml \
  https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/values.yaml
```

Edit the `values.yaml` file, paying particular attention to

| Name                   | Description                                                                                  |
|------------------------|----------------------------------------------------------------------------------------------|
| `namespace.name`       | Create a unique namespace for your deployment, or deploy in `default`                        |
| `secret.name`          | Create a secret to store the FiftyOne Teams secrets                                          |
| `secret.createSecrets` | If you set `secret.create` to `true` you can have this Helm chart create secrets for you.    |
| `env.nonsensitive`     | Non-sensitive environment variables and their values                                         |
| `env.sensitive`        | A mapping of sensitive environment variables and the key that stores their value             |
| `image.repository`     | The image to deploy                                                                          |
| `ingress.hosts.host`   | The Fully Qualified Domain Name [FQDN] of the deployment                                     |
| `tls.secretName`       | The name of the secret that contains `tls.crt` and `tls.key` values for your SSL Certificate |
| `tls.hosts`            | The FQDN of the deployment                                                                   |

You must provide `FIFTYONE_TEAMS_ORGANIZATION`, `FIFTYONE_TEAMS_CLIENT_ID`, and `FIFTYONE_DATABASE_URI` environment variables with values provided by Voxel51.  Without those variables the environment will not load correctly.

`FIFTYONE_DATABASE_ADMIN` is set to `false` by default.  This is in order to make sure that upgrades do not break existing client installs.
- If you are performing a new install, consider setting `env.nonsensitive.FIFTYONE_DATABASE_ADMIN` to `true`
- If you are performing an upgrade, please review our [Upgrade Process Recommendations](#upgrade-process-recommendations)

Please contact [Voxel51](https://voxel51.com/#teams-form) if you would like more information regarding Fiftyone Teams.

Once you have edited the `values.yaml` file you can deploy your FiftyOne Teams instance with:
```
helm repo add voxel51 https://helm.fiftyone.ai
helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
```

---

## Upgrade Process Recommendations

The FiftyOne Teams 0.8.8 Database (version `0.16.6`) is forward-compatible with the FiftyOne Teams 0.9.2 Client (database version `0.17.2`).  Voxel51 recommends the following upgrade process:

1. Ensure all Python clients set `FIFTYONE_DATABASE_ADMIN=false`
1. Upgrade FiftyOne Teams Python clients to FiftyOne Teams 0.9.2
1. Upgrade your FiftyOne Teams Kubernetes deploy to Helm version v0.2.2
1. Have an admin set `FIFTYONE_DATABASE_ADMIN=true` in their local Python client
1. Have the admin run `fiftyone migrate --all` to upgrade all datasets
1. Use `fiftyone migrate --info` to ensure that all datasets are now at version `0.17.2`

---

## A Full GKE Deployment Example

The following instructions represent a full Google Kubernetes Engine [GKE] deployment using:
- The jetstack/cert-manager Helm chart for Let's Encrypt SSL certificates
- The bitnami/mongodb Helm chart for MongoDB
- The voxel51/fiftyone-teams-app Helm chart

These instructions assume you have [kubectl](https://kubernetes.io/docs/tasks/tools/) and [Helm](https://helm.sh/docs/intro/install/) installed and operating, and that you have an existing [GKE Cluster available](https://cloud.google.com/kubernetes-engine/docs/concepts/kubernetes-engine-overview).

These instructions assume you have received Docker Hub credentials from Voxel51 and have placed your `voxel51-docker.json` file in the current directory; if your `voxel51-docker.json` is not in the current directory please update the command line accordingly.

These instructions assume you have received your Auth0 Organization ID and Client ID from Voxel51.  If you have not received these IDs, please contact your [Voxel51 Support Team](mailto:support@voxel51.com).

### Download the Example Configuration Files

Download the example configuration files from the [Voxel51 GitHub](https://github.com/voxel51/fiftyone-teams-app-deploy) repository:

```
curl -o values.yaml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/values.yaml
curl -o clusterissuer.yml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/clusterissuer.yml
```

You will need to edit the `values.yaml` file to include your `mongodbConnectionString`, your `organiationId`, your `clientId`, and your `host` values (search for `replace.this.dns.name` - it appears in two locations).  Assuming you follow these directions your MongoDB host will be `fiftyone-mongodb.fiftyone-mongodb.svc.cluster.local`; please modify that hostname if you modify these instructions.

If you have not configured IAM to allow your GKE cluster to access your Cloud Storage you will want to edit the `values.yaml` file to include a `volume` and `volumeMounts` entry for your cloud storage credentials, set the appropriate `GOOGLE_APPLICATION_CREDIALS` `nonsensitive` environment variable, and follow the instructions in the `values.yaml` to create the appropriate secret.

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

### Installation Complete

Congratulations! You should now be able to access your FiftyOne Teams installation at the DNS address you created [earlier](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).

