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
  - [Security](#security)
  - [Costs](#costs)
  - [Sizing](#sizing)
  - [Deployment Assets](#deployment-assets)
  - [Health Check](#health-check)
  - [Backup and Recovery](#backup-and-recovery)
  - [Routine Maintenance](#routine-maintenance)
  - [Emergency Maintenance](#emergency-maintenance)
  - [Support](#support)

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

### Security

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| DSEC-002 | The application does not require the use of AWS account root privileges for deployment or operation. | This is covered in the ... |
| DSEC-003 | The deployment guide provides prescriptive guidance on following the policy of least privilege for all access granted as part of the deployment. | This is covered in the ... |
| DSEC-004 | The deployment guide clearly documents any public resources (e.g. Amazon S3 buckets with bucket policies allowing public access). | The deployment guide is not using public resources. |
| DSEC-006 | The deployment guide describes the purpose of each AWS Identity and Access Management (IAM) role and IAM policy the user is instructed to create. | This is covered in the ... |
| DSEC-007 | The deployment guide provides clear instruction on maintaining any stored secrets such as database credentials stored in AWS Secrets Manager. | This is covered in the ... |
| DSEC-008 | The deployment guide includes details on where customer sensitive data are stored | This is covered in the ... |
| DSEC-009 | The deployment guide must explain all data encryption configuration (for example. Amazon Simple Storage Service (Amazon S3) server-side encryption, Amazon Elastic Block Store (Amazon EBS) encryption, and Linux Unified Key Setup (LUKS)) |This is covered in the ... |
| DSEC-010 | For deployments involving more than a single element, include network configuration (for example, VPCs, subnets, security groups, network access control lists (network ACLs), and route tables) in the deployment guide. | This is covered in the ... |
| DSEC-011 | The solution must support the ability for the customer to disable Instance Metadata Service Version 1 (IMDSv1). | AWS EKS does not offer direct control over the Instance Metadata Service (IMDS). To mitigate risk linked with IMDS we are using the least privilege principle with a specific role for task execution and specific security group and VPC to control network access to EKS pods. |

### Costs

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| CST-001 | The deployment guide includes a list of billable services and guidance on whether each service is mandatory or optional. | This is covered in the ... |
| CST-002 | The deployment guide includes the cost model and licensing costs. | This is covered in the ... |

### Sizing

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| SIZ-001 | Either provide scripts to provision required resources or provide guidance for type and size selection for resources. | This is covered in the ... |

### Deployment Assets

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| DAS-001 | The deployment guide provides step-by-step instructions for deploying the workload on AWS according to the typical deployment architecture. | This is covered in the ... |
| DAS-004 | The deployment guide contains prescriptive guidance for testing and troubleshooting. | This is covered in the ... |

### Health Check

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| HLCH-001 | The deployment guide provides step-by-step instructions for how to assess and monitor the health and proper function of the application. | This is covered in the ... |

### Backup and Recovery

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| BAR-001 | Identify the data stores and the configurations to be backed up. If any of the data stores are proprietary, provide step-by-step instructions for backup and recovery. | This is covered in the ... |

### Routine Maintenance

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| RM-001 | The deployment guide provides step-by-step instructions for rotating programmatic system credentials and cryptographic keys. | This is covered in the ... |
| RM-002 | The deployment guide provides prescriptive guidance for software patches and upgrades. | This is covered in the ... |
| RM-003 | The deployment guide provides prescriptive guidance on managing licenses. | This is covered in the ... |
| RM-004 | The deployment guide provides prescriptive guidance on managing AWS service limits. | This is covered in the ... |

### Emergency Maintenance

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| EMER-001 | The deployment guide provides step-by-step instructions on handling fault conditions. | This is covered in the ... |
| EMER-002 | The deployment guide provides step-by-step instructions on how to recover the software. | This is covered in the ... |

### Support

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| SUP-001 | The deployment guide provides details on how to receive support. | This is covered in the ... |
| SUP-002 | The deployment guide provides details on technical support tiers. | This is covered in the ... |
| SUP-003 | The deployment guide provides prescriptive guidance on managing licenses. | This is covered in the ... |
