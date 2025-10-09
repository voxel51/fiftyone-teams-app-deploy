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

# Deployment Guide (Customer-Deployed Foundational Technical Review)

This is the summary of FiftyOne Enterprise FTR based on the
[AWS guidelines](https://apn-checklists.s3.amazonaws.com/foundational/customer-deployed/customer-deployed/C0hfGvKGP.html).

<!-- toc -->

- [Prerequisites Skills and Knowledge](#prerequisites-skills-and-knowledge)
- [Technical Requirements](#technical-requirements)

<!-- tocstop -->

## Prerequisites Skills and Knowledge

The following prerequisites skills & knowledge
are required for a successful and properly secured
deployment of FiftyOne Enterprise.

1. A knowledge of kubernetes and
   [AWS EKS](https://aws.amazon.com/pm/eks/).

1. A knowledge of installing helm charts to deploy kubernetes applications.

1. A knowledge MongoDB.

1. A knowledge of
   [AWS Route53](https://aws.amazon.com/route53/)
   and the ability to generate, modify, and delete DNS records.

1. A knowledge of
   [AWS ACM](https://aws.amazon.com/certificate-manager/)
   and the ability to generate TLS/SSL certificates.

1. (optional) A knowledge of network file systems (NFS) or another
   `ReadWriteMany`-compatible storage medium such as
   [AWS EFS](https://docs.aws.amazon.com/eks/latest/userguide/efs-csi.html).

## Technical Requirements

The following technical requirements
are required for a successful and properly secured
deployment of FiftyOne Enterprise.

1. An AWS EKS cluster matching the FiftyOne Enterprise
   [kubernetes version requirements](../fiftyone-teams-app/README.md#kubernetes-cluster-and-kubectl)

1. An installation of `helm` that matches the
   [helm version requirements](../fiftyone-teams-app/README.md#helm).

1. A MongoDB Database that meets FiftyOne's
   [version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).

1. An AWS Route53 record or records for ingress.

1. An AWS ACM certificate or certificates for HTTPS ingress.

1. (optional) An NFS server or `ReadWriteMany` compatible storage medium for
   [delegated operators](../fiftyone-teams-app/README.md#builtin-delegated-operator-orchestrator),
   [plugins](../fiftyone-teams-app/README.md#plugins),
   and
   [API high-availability](../fiftyone-teams-app/README.md#highly-available-fiftyone-teams-api-deployments)
