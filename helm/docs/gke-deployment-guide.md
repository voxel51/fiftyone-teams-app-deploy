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

This guide walks through a full Google Kubernetes Engine [GKE] deployment of
FiftyOne Enterprise combining the following Helm charts:

- [jetstack/cert-manager](https://github.com/cert-manager/cert-manager)
  - For Let's Encrypt SSL certificates
- [mongodb/community-operator](https://github.com/mongodb/helm-charts/tree/main/charts/community-operator)
  - for MongoDB
- voxel51/fiftyone-teams-app

This guide is a complete, cloud-specific worked example of the generic steps
in the [Helm README](../README.md). If you are deploying to AWS EKS instead,
see the [AWS Deployment Guide](./aws-deployment-guide.md).

## Prerequisites

These instructions assume you have

- These tools installed and operating
  - [kubectl](https://kubernetes.io/docs/tasks/tools/)
  - [Helm](https://helm.sh/docs/intro/install/)
- An existing
  [GKE Cluster available](https://cloud.google.com/kubernetes-engine/docs/concepts/kubernetes-engine-overview)
- Received Docker Hub credentials from Voxel51
  - Have `voxel51-docker.json` file in the current directory
    - If `voxel51-docker.json` is not in the current directory,
      please update the command line accordingly.
- Your license file from the Voxel51 Customer Success Team
  - If you have not received this information, please contact your
    Voxel51 Support Team via your agreed-upon mechanism (Slack, email, etc.)

> **NOTE**: Anytime a license file secret is updated, you
> must restart the `teams-cas` and `teams-api` services.
> You may delete the pods, or run
>
> ```shell
> kubectl rollout restart deploy \
>   -n your-namespace \
>   teams-cas \
>   teams-api
> ```

## Download the Example Configuration Files

Download the example configuration files from the
[voxel51/fiftyone-teams-app-deploy](https://github.com/voxel51/fiftyone-teams-app-deploy/tree/main/helm/gke-example)
GitHub repository.

For example

```shell
curl -o values.yaml \
  https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/values.yaml
curl -o cluster-issuer.yaml \
  https://raw.githubusercontent.com/voxel51/fiftyone-teams-app-deploy/main/helm/gke-example/cluster-issuer.yaml
curl -o frontend-config.yaml \
  https://github.com/voxel51/fiftyone-teams-app-deploy/blob/main/helm/gke-example/frontend-config.yaml
```

Update the `values.yaml` file with

- In `secret.fiftyone`
  - MongoDB
    - Set `mongodbConnectionString` containing your MongoDB username and password
  - Set `cookieSecret`
  - Set `encryptionKey`
  - Set `fiftyoneAuthSecret`
- In `teamsAppSettings.dnsName`
  - Set ingress `host` values

Assuming you follow these directions your MongoDB host will be
`fiftyone-teams-mongodb.fiftyone-teams-mongodb.svc.cluster.local`.
<!-- Please modify this hostname if you modify these instructions. -->

## Create the Necessary Helm Repos

Add the Helm repositories

```shell
helm repo add mongodb https://mongodb.github.io/helm-charts
helm repo add jetstack https://charts.jetstack.io
helm repo add voxel51 https://helm.fiftyone.ai
helm repo update
```

## Install and Configure cert-manager

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

`ClusterIssuers` are Kubernetes resources that represent certificate authorities
that are able to generate signed certificates by honoring certificate signing requests.
You must create either an `Issuer` in each namespace or a `ClusterIssuer`
as part of your cert-manager configuration.
Voxel51 has provided an example `ClusterIssuer` configuration (downloaded
[earlier](#download-the-example-configuration-files)
in this guide).

```shell
kubectl apply -f ./cluster-issuer.yaml
```

## Install and Configure MongoDB

These
[instructions](https://github.com/mongodb/helm-charts/tree/main/charts/community-operator#deploying-a-mongodb-replica-set)
can be used to deploy a MongoDB replicaset in your GKE cluster.

Wait until the MongoDB pods are in the `Ready` state before
beginning the "Install FiftyOne Enterprise App" instructions.

While waiting,
[configure a DNS entry](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).

To determine the state of the `fiftyone-teams-mongodb` pods, run

```shell
kubectl get pods
```

## Obtain a Global Static IP Address and Configure a DNS Entry

Reserve a global static IP address for use in your cluster:

```shell
gcloud compute addresses create \
  fiftyone-teams-static-ip --global --ip-version IPV4
gcloud compute addresses describe \
  fiftyone-teams-static-ip --global
```

Record the IP address and either create a DNS entry or contact your Voxel51
support team to have them create an appropriate `fiftyone.ai` DNS entry for you.

## Set up HTTP to HTTPS Forwarding

```shell
kubectl apply -f frontend-config.yaml
```

For more information, see
[HTTP to HTTPS redirects](https://cloud.google.com/kubernetes-engine/docs/how-to/ingress-configuration#https_redirect).

## Install FiftyOne Enterprise App

```shell
kubectl create namespace fiftyone-teams
kubectl config set-context --current --namespace fiftyone-teams
kubectl create secret generic regcred \
  --from-file=.dockerconfigjson=./voxel51-docker.json \
  --type kubernetes.io/dockerconfigjson
helm install fiftyone-teams-app voxel51/fiftyone-teams-app \
  --values ./values.yaml
```

Issuing SSL Certificates can take up to 15 minutes.
Be patient while Let's Encrypt and GKE negotiate.

You can verify that your SSL certificates have been
properly issued with the following curl command:

```shell
curl -I https://replace.this.dns.name
```

Your SSL certificates have been correctly issued when
you see `HTTP/2 200` at the top of the response.
If, however, you encounter a
`SSL certificate problem: unable to get local issuer certificate`
message you should delete the certificate and allow it to recreate.

```shell
kubectl delete secret fiftyone-teams-cert-secret
```

Further instructions for debugging ACME certificates are on the
[cert-manager docs site](https://cert-manager.io/docs/faq/acme/).

Once your installation is complete, browse to
`/settings/cloud_storage_credentials`
and add your storage credentials to access sample data.

## Installation Complete

Congratulations! You should now be able to access your
FiftyOne Enterprise installation at the DNS address you created
[earlier](#obtain-a-global-static-ip-address-and-configure-a-dns-entry).

Next, continue with the remaining steps in the
[Helm README](../README.md), starting from
[Step 7: Configure Delegated Operation Logs](../README.md#page_facing_up-step-7-configure-delegated-operation-logs)
(ingress/TLS is already handled by this guide),
to finish setting up authentication and CAS.
