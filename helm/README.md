# Deploying FiftyOne Teams App Using Helm

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

Please contact [Voxel51](https://voxel51.com/#teams-form) if you would like more information regarding Fiftyone Teams.

To get the default `values.yaml` file, try:
`curl -o values.yaml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/values.yaml`

Once you have edited the `values.yaml` file you can deploy your FiftyOne Teams instance with:
```
helm repo add voxel51 https://helm.fiftyone.ai
helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
```

---

## A full GKE deployment

The following instructions represent a full deployment using:
- The jetstack/cert-manager helm chart for Let's Encrypt SSL certificates
- The bitnami/mongodb helm chart for MongoDB
- The Voxel51 FiftyOne Teams App Helm chart

These instructions assume you have [kubectl](https://kubernetes.io/docs/tasks/tools/) and [Helm](https://helm.sh/docs/intro/install/) installed and operating, that you have cloned the [fiftyone-teams-app-deploy](https://github.com/voxel51/fiftyone-teams-app-deploy) repository, and you are in the `helm` subdirectory of that repository.

### Download the example configuration files

Download the example configuration files from the [Voxel51 GitHub](https://github.com/voxel51/fiftyone-teams-app-deploy) repository:

```
curl -o values.yaml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/values.yaml
curl -o clusterissuer.yml https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/clusterissuer.yml
```

You will need to edit the `values.yaml` file to include your `mongodbConnectionString`, your `organiationId`, your `clientId`, and your `host` values (search for `replace.this.dns.name` - it appears in two locations).

You will most likely want to edit the `values.yaml` file to include a `volume` and `volumenMounts` entry for your cloud storage credentials, and set the appropriate `AWS_CONFIG_FILE`, `GOOGLE_APPLICATION_CREDIALS`, or `MINIO_CONFIG_FILE` `nonsensitive` environment variable and follow the instructions to create the appropriate secret.

These instructions assume you have placed your `voxel51-docker.json` file in the current directory; if you have not please modify the command lines as appropriate.

### Create the necessary Helm repos

Create the jetstack, bitnami, and voxel51 Helm repositories:
```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add jetstack https://charts.jetstack.io
helm repo add voxel51 https://helm.fiftyone.ai
helm repo update
```


### Install and configure cert-manager

If you are using a GKE Autopilot cluster, please review the information [provided by cert-manager](https://github.com/cert-manager/cert-manager/issues/3717#issuecomment-919299192) and adjust your installation accordingly.

```
kubectl create namespace cert-manager
kubectl config set-context --current --namespace cert-manager
helm install cert-manager jetstack/cert-manager --set installCRDs=true
kubectl apply -f ./clusterissuer.yaml
```

### Install and configure MongoDB

These instructions deploy a single-node MongoDB instance in your GKE cluster.  If you would like to deploy MongoDB with a replicaset configuration, please refer to the [MongoDB Helm Chart](https://github.com/bitnami/charts/tree/master/bitnami/mongodb) documentation.

**You will probably want to edit the `rootPassword` and `rootUser` defined below.**

```
kubectl create namespace fiftyone-mongodb
kubectl config set-context --current --namespace fiftyone-mongodb
helm install fiftyone-mongodb \
    --set image.tag=4.4 \
    --set auth.rootPassword=REPLACEME \
    --set auth.rootUser=admin \
    --set global.namespaceOverride=fiftyone-mongodb \
    --set namespaceOverride=fiftyone-mongodb \
    bitnami/mongodb
```

Wait until the MongoDB pods have been created and are OK.

### Install and configure FiftyOne Teams App

First, reserve a global static IP address:
```
gcloud compute addresses create fiftyone-teams-static-ip --global --ip-version IPV4
gcloud compute addresses describe fiftyone-teams-static-ip --global
```

Record the IP address and either create a DNS entry or contact your Voxel51 support team to have them create an appropriate `fiftyone.ai` DNS entry for you.

```
kubectl namespace create fiftyone-teams
kubectl config set-context --current --namespace fiftyone-teams
kubectl create secret generic regcred \
    --from-file=.dockerconfigjson=./gke-example/voxel51-docker.json \
    --type kubernetes.io/dockerconfigjson
helm install fiftyone-teams-app voxel51/fiftyone-teams-app -f ./values.yaml
```

You can verify that your SSL certificates have been properly issued with the following curl command:

`curl -I https://replace.this.dns.name`

Your SSL certificates have been correctly issued if you see `HTTP/2 200` at the top of the response.  If, however, you encounter a `SSL certificate problem: unable to get local issuer certificate` message you should delete the certificate and allow it to be recreated.

`kubectl delete secret fiftyone-teams-cert-secret`

Further instructions for debugging ACME certificates can be found on the [cert-manager docs site](https://cert-manager.io/docs/faq/acme/).
