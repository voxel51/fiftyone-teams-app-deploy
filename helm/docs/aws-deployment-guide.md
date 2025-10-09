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
- [AWS FTR Summary](#aws-ftr-summary)
  - [Introduction](#introduction)
  - [Prerequisites and Requirements](#prerequisites-and-requirements)
  - [Architecture Diagrams](#architecture-diagrams)

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

1. An installation of
   [`helm`](https://helm.sh/)
   that matches the
   [`helm` version requirements](../fiftyone-teams-app/README.md#helm).

1. A
   [`MongoDB` Database](https://www.mongodb.com/)
   that meets FiftyOne Enterprise's
   [version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).

1. An AWS Route53 record or records for ingress.

1. An AWS ACM certificate or certificates for HTTPS ingress.

1. (optional) An NFS server or `ReadWriteMany` compatible storage medium for
   [delegated operators](../fiftyone-teams-app/README.md#builtin-delegated-operator-orchestrator),
   [plugins](../fiftyone-teams-app/README.md#plugins),
   and
   [API high-availability](../fiftyone-teams-app/README.md#highly-available-fiftyone-teams-api-deployments)

## AWS FTR Summary

### Introduction

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| INT-001 | Introductory material must contain use cases for the software. | This is covered in the ...|
| INT-002 | Introductory material contains an overview of a typical customer deployment, including lists of all resources that are set up when the deployment is complete. | This is covered in the ... |
| INT-003 | Introductory material contains a description of all deployment options discussed in the user guide (e.g. single-AZ, multi-AZ or multi-region), if applicable. | This is covered in the ... |
| INT-004 | Introductory material contains the expected amount of time to complete the deployment. | Approximately 2 hours. |
| INT-005 | Introductory material contains the regions supported. | There is no limitation on region supported for this service. |

### Prerequisites and Requirements

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| PRQ-001 | All technical prerequisites and requirements to complete the deployment process are listed (e.g. required OS, database type and storage requirements). | This is covered in the [Technical Requirements](#technical-requirements) section. |
| PRQ-002 | The deployment guide lists all prerequisite skills or specialized knowledge (for example, familiarity with AWS, specific AWS services, or a scripting or programming language). |  This is covered in the [Prerequisites Skills and Knowledge](#prerequisites-and-requirements) section. |
| PRQ-003 | The deployment guide lists the environment configuration that is needed for the deployment (e.g. an AWS account, a specific operating system, licensing, DNS). | This is covered ... |

### Architecture Diagrams

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| ARCH-001 | Architecture diagrams must include all AWS services and resources deployed by the solution and illustrate how the services and resources connect with each other in a typical customer environment. | This is covered in the ... |
| ARCH-004 | Architecture diagrams use official AWS Architecture Icons. | This is covered in the ... |
| ARCH-005 | Network diagrams demonstrate virtual private clouds (VPCs) and subnets. | This is covered in the ... |
| ARCH-006 | Architecture diagrams show integration points, including third-party assets/APIs and on-premises/hybrid assets. | This is covered in the ... |
