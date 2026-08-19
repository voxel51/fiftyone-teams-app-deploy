<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img alt="Voxel51 Logo" src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img alt="Voxel51 FiftyOne" src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# GKE Deployment Guide

<!-- toc -->

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Download the Example Configuration Files](#download-the-example-configuration-files)
- [Create the Necessary Helm Repos](#create-the-necessary-helm-repos)
- [Install and Configure cert-manager](#install-and-configure-cert-manager)
  - [Create a ClusterIssuer](#create-a-clusterissuer)
- [Install and Configure MongoDB](#install-and-configure-mongodb)
- [Obtain a Global Static IP Address and Configure a DNS Entry](#obtain-a-global-static-ip-address-and-configure-a-dns-entry)
- [Set up HTTP to HTTPS Forwarding](#set-up-http-to-https-forwarding)
- [Install FiftyOne Enterprise App](#install-fiftyone-enterprise-app)
- [Installation Complete](#installation-complete)

<!-- tocstop -->

## Introduction

This guide walks through a full Google Kubernetes Engine [GKE] deployment
of FiftyOne Enterprise combining the following Helm charts:

- [jetstack/cert-manager](https://github.com/cert-manager/cert-manager)
  - For Let's Encrypt SSL certificates
- [mongodb/community-operator](https://github.com/mongodb/helm-charts/tree/main/charts/community-operator)
  - for MongoDB
- voxel51/fiftyone-teams-app

This guide is a complete, cloud-specific worked example of the generic steps
in the [Helm README](../README.md).
If you are deploying to AWS EKS instead,
see the [AWS Deployment Guide](./aws-deployment-guide.md).

## Prerequisites

These instructions assume you have

- These tools installed and operating
  - [kubectl](https://kubernetes.io/docs/tasks/tools/)
  - [Helm](https://helm.sh/docs/intro/install/)
- An existing
  [GKE Cluster available](https://cloud.google.com/kubernetes-engine/docs/concepts/kubernetes-engine-overview)
- A Docker Hub credentials provided by us
  - Have `voxel51-docker.json` file in the current directory
    - If `voxel51-docker.json` is not in the current directory,
      please update the command line accordingly
- A license file provide by us (Voxel51 Customer Success Team)
  - To obtain a license file,
    please contact your Voxel51 Support Team
    via Slack or email

> **NOTE**: When you update a license file secret,
> restart the `teams-cas` and `teams-api` services.
> Run
>
> ```shell
> kubectl rollout restart deploy \
>   -n your-namespace-here \
>   teams-cas \
>   teams-api
> ```

## Download the Example Configuration Files

Clone the this repository
[voxel51/fiftyone-teams-app-deploy](https://github.com/voxel51/fiftyone-teams-app-deploy)

```shell
# clone using ssh
git clone git@github.com:voxel51/fiftyone-teams-app-deploy.git

# clone using https
git clone https://github.com/voxel51/fiftyone-teams-app-deploy.git

# navigate to the example directory
cd fiftyone-teams-app-deploy/helm/gke-example
```

Update the `values.yaml` file setting

```yaml
secret:
  fiftyone:
    # Update the hostname when necessary
    mongodbConnectionString: mongodb://<YOUR_USERNAME>:<YOUR_PASSWORD>@fiftyone-teams-mongodb.fiftyone-teams-mongodb.svc.cluster.local/?authSource=admin
    cookieSecret: <YOUR_COOKIE_SECRET>
    encryptionKey: <YOUR_ENCRYPTION_KEY>
    fiftyoneAuthSecret: <YOUR_FIFTYONE_AUTH_SECRET>
teamsAppSettings:
  dnsName: <YOUR_DNS_HOST>
```

## Add Helm Repositories

Add the Helm repositories

```shell
helm repo add mongodb https://mongodb.github.io/helm-charts
helm repo add jetstack https://charts.jetstack.io
helm repo add voxel51 https://helm.fiftyone.ai
helm repo update
```

## Install and Configure cert-manager

If you are using a GKE Autopilot cluster,
please review the information
[provided by cert-manager](https://github.com/cert-manager/cert-manager/issues/3717#issuecomment-919299192)
and adjust your installation accordingly.

```shell
kubectl create namespace cert-manager
helm install cert-manager jetstack/cert-manager \
  --set installCRDs=true\
  --namespace cert-manager
```

Follow the cert-manager instructions to
[verify the cert-manager Installation](https://cert-manager.io/v1.4-docs/installation/verify/).

### Create a ClusterIssuer

`ClusterIssuers` are Kubernetes resources
that represent certificate authorities.
They generate signed certificates by honoring certificate signing requests.
You must create either an `Issuer` in each namespace
or a `ClusterIssuer` as part of your cert-manager configuration.
Apply the example `ClusterIssuer` configuration in
[cluster-issuer.yaml](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/gke-example/cluster-issuer.yaml).

```shell
kubectl apply -f ./cluster-issuer.yaml
```

## Install and Configure MongoDB

Use these
[Deploying a MongoDB Replica Set](https://github.com/mongodb/helm-charts/tree/main/charts/community-operator#deploying-a-mongodb-replica-set)
instructions to deploy a MongoDB Replica Set in your GKE cluster.

Wait until the MongoDB pods are in the `Ready` state
before following [Install FiftyOne Enterprise App](#install-fiftyone-enterprise-app).

While waiting,
[configure a DNS entry](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).

To determine the state of the `fiftyone-teams-mongodb` pods, run

```shell
kubectl get pods \
  -n your-mongodb-namespace
```

## Obtain a Global Static IP Address and Configure a DNS Entry

Reserve a global static IP address for use in your cluster:

```shell
gcloud compute addresses create \
  fiftyone-teams-static-ip \
  --global \
  --ip-version IPV4
gcloud compute addresses describe \
  fiftyone-teams-static-ip \
  --global
```

Record the IP address
and create a DNS entry.

## Set up HTTP to HTTPS Forwarding

Apply the example `FrontendConfig` in
[frontend-config.yaml](https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/gke-example/frontend-config.yaml).

```shell
kubectl apply \
  -f frontend-config.yaml \
  --namespace your-namespace-here
```

For more information about `FrontendConfig`s, see
[HTTP to HTTPS redirects](https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-configuration#https_redirect).

## Install FiftyOne Enterprise App

```shell
kubectl create namespace your-namespace-here
kubectl create secret generic regcred \
  --from-file=.dockerconfigjson=./voxel51-docker.json \
  --type kubernetes.io/dockerconfigjson
helm install fiftyone-teams-app voxel51/fiftyone-teams-app \
  --values ./values.yaml \
  --namespace your-namespace-here
```

Issuing SSL Certificates may take up to 15 minutes.
Be patient while Let's Encrypt and GKE negotiate.

Run this curl command to check that your SSL certificates were issued:

```shell
curl -I https://<YOUR_DNS_NAME>
```

You'll know your SSL certificates were issued correctly
when you see `HTTP/2 200` at the top of the response.
If, however, you encounter a
`SSL certificate problem: unable to get local issuer certificate` message,
delete the certificate.
It will recreate automatically.

```shell
kubectl delete secret fiftyone-teams-cert-secret
```

For further instructions for debugging ACME certificates,
see
[Troubleshooting Problems with ACME / Let's Encrypt Certificates](https://cert-manager.io/docs/faq/acme/).

Once your installation is complete,
browse to `https://<YOUR_DNS_NAME>/settings/cloud_storage_credentials`.
Add your storage credentials to access sample data.

## Installation Complete

Congratulations!
You may now access your FiftyOne Enterprise installation
at the DNS address you created in
[Obtain a Global Static IP Address and Configure a DNS Entry](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).

Proceed to
[Helm README](../README.md)
and follow the rest of the document starting at
[Step 10: Initial CAS Setup](../README.md#step-10-initial-cas-setup).
