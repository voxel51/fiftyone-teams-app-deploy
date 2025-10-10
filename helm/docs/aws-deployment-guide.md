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
[AWS guidelines][aws-ftr].

<!-- toc -->

- [Prerequisites and Requirements](#prerequisites-and-requirements)
  - [Prerequisites Skills and Knowledge](#prerequisites-skills-and-knowledge)
  - [Technical Requirements](#technical-requirements)
- [Security](#security)
  - [Princple Of Least Privilege](#princple-of-least-privilege)
  - [Public Resources](#public-resources)
  - [Root Privileges](#root-privileges)
  - [Secrets And Sensitive Data](#secrets-and-sensitive-data)
  - [Instance Metadata Service Version 1](#instance-metadata-service-version-1)
- [Costs](#costs)
  - [Billable Services](#billable-services)
  - [License Model](#license-model)
- [Sizing](#sizing)
- [Deployment Assets](#deployment-assets)
- [Health Check](#health-check)
- [Backup and Recovery](#backup-and-recovery)
- [AWS FTR Summary](#aws-ftr-summary)
  - [Introduction](#introduction)
  - [Prerequisites and Requirements](#prerequisites-and-requirements-1)
  - [Architecture Diagrams](#architecture-diagrams)
  - [Security](#security-1)
  - [Costs](#costs-1)
  - [Sizing](#sizing-1)
  - [Deployment Assets](#deployment-assets-1)
  - [Health Check](#health-check-1)
  - [Backup and Recovery](#backup-and-recovery-1)
  - [Routine Maintenance](#routine-maintenance)
  - [Emergency Maintenance](#emergency-maintenance)
  - [Support](#support)

<!-- tocstop -->

## Prerequisites and Requirements

### Prerequisites Skills and Knowledge

The following prerequisites skills & knowledge
are required for a successful and properly secured
deployment of FiftyOne Enterprise.

1. A knowledge of kubernetes and
   [AWS EKS][aws-eks].

1. A knowledge of using
   [`helm`][helm-sh]
   to install and deploy kubernetes applications.

1. A knowledge
  [MongoDB][mongodb-com].

1. A knowledge of
   [AWS Route53][aws-route-53]
   and the ability to generate, modify, and delete DNS records.

1. A knowledge of
   [AWS ACM][aws-acm]
   and the ability to generate TLS/SSL certificates.

1. (optional) A knowledge of network file systems (NFS) or another
   `ReadWriteMany`-compatible storage medium such as
   [AWS EFS][aws-efs].

### Technical Requirements

The following technical requirements
are required for a successful and properly secured
deployment of FiftyOne Enterprise.

1. An
   [AWS EKS][aws-eks]
   cluster matching the FiftyOne Enterprise
   [kubernetes version requirements](../fiftyone-teams-app/README.md#kubernetes-cluster-and-kubectl)

1. An installation of
   [`helm`][helm-sh]
   that matches the
   [`helm` version requirements](../fiftyone-teams-app/README.md#helm).

1. A
   [MongoDB Database][mongodb-com]
   that meets FiftyOne Enterprise's
   [version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).

1. An
   [AWS Route53][aws-route-53]
   record or records for ingress.

1. An
   [AWS ACM][aws-acm]
   certificate or certificates for HTTPS ingress.

1. (optional) An NFS server or `ReadWriteMany` compatible storage medium for
   [delegated operators](../fiftyone-teams-app/README.md#builtin-delegated-operator-orchestrator),
   [plugins](../fiftyone-teams-app/README.md#plugins),
   and
   [API high-availability](../fiftyone-teams-app/README.md#highly-available-fiftyone-teams-api-deployments)

## Security

### Princple Of Least Privilege

When deploying FiftyOne Enterprise, Voxel51 recommends following the principle
of least privilege.
The minimum privileges needed for the FiftyOne Enterprise application are
listed in the
[workload identity with AWS](./configure-workload-identity-federation.md#workload-identity-with-aws)
section.

### Public Resources

FiftyOne Enterprise does not require or create any public resources.
Customers may use a public or private DNS record to access resources.

### Root Privileges

FiftyOne Enterprise does not require AWS Account root privileges.

### Secrets And Sensitive Data

Please refer to the
[Secrets And Sensitive Data](../fiftyone-teams-app/README.md#secrets-and-sensitive-data)
section for questions related to database credentials, cookie secrets,
and other sensitive data related to FiftyOne Enterprise.

### Instance Metadata Service Version 1

AWS EKS does not offer direct control over the
Instance Metadata Service (IMDS).
To mitigate risk linked with IMDS we are using the least privilege
principle with a specific role for task execution and specific
security group and VPC to control network access to EKS pods.

## Costs

### Billable Services

The billable services that are **mandatory** to run FiftyOne Enterprise are:

1. [AWS EKS][aws-eks]

1. [AWS Route53][aws-route-53]

1. [AWS ACM][aws-acm]

1. [AWS ELB][aws-elb]

The billable services that are **optional** to run FiftyOne Enterpise are:

1. [AWS S3][aws-s3]

1. [AWS EC2][aws-ec2]

### License Model

## Sizing

## Deployment Assets

## Health Check

## Backup and Recovery

Please refer to the
[Backup and Recovery](../fiftyone-teams-app/README.md#backup-and-recovery)
section for questions related to backing up and restoring your
FiftyOne Enterpise application.

## AWS FTR Summary

### Introduction

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| INT-001 | Introductory material must contain use cases for the software. | This is covered in the ...|
| INT-002 | Introductory material contains an overview of a typical customer deployment, including lists of all resources that are set up when the deployment is complete. | This is covered in the ... |
| INT-003 | Introductory material contains a description of all deployment options discussed in the user guide (e.g. single-AZ, multi-AZ or multi-region), if applicable. | This is covered in the ... |
| INT-004 | Introductory material contains the expected amount of time to complete the deployment. | This is covered in the [estimated completion time](../fiftyone-teams-app/README.md#estimated-completion-time) section. |
| INT-005 | Introductory material contains the regions supported. | There is no limitation on region supported for this service. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Prerequisites and Requirements

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| PRQ-001 | All technical prerequisites and requirements to complete the deployment process are listed (e.g. required OS, database type and storage requirements). | This is covered in the [Technical Requirements](#technical-requirements) section. |
| PRQ-002 | The deployment guide lists all prerequisite skills or specialized knowledge (for example, familiarity with AWS, specific AWS services, or a scripting or programming language). |  This is covered in the [Prerequisites Skills and Knowledge](#prerequisites-and-requirements) section. |
| PRQ-003 | The deployment guide lists the environment configuration that is needed for the deployment (e.g. an AWS account, a specific operating system, licensing, DNS). | This is covered in the [values](../fiftyone-teams-app/README.md#values) section. |

### Architecture Diagrams

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| ARCH-001 | Architecture diagrams must include all AWS services and resources deployed by the solution and illustrate how the services and resources connect with each other in a typical customer environment. | This is covered in the ... |
| ARCH-004 | Architecture diagrams use official AWS Architecture Icons. | This is covered in the ... |
| ARCH-005 | Network diagrams demonstrate virtual private clouds (VPCs) and subnets. | This is covered in the ... |
| ARCH-006 | Architecture diagrams show integration points, including third-party assets/APIs and on-premises/hybrid assets. | This is covered in the ... |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Security

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| DSEC-002 | The application does not require the use of AWS account root privileges for deployment or operation. | Thiis is covered in the [Root Privileges](#root-privileges) section. |
| DSEC-003 | The deployment guide provides prescriptive guidance on following the policy of least privilege for all access granted as part of the deployment. | This is covered in the [Princple Of Least Privilege](#princple-of-least-privilege) section. |
| DSEC-004 | The deployment guide clearly documents any public resources (e.g. Amazon S3 buckets with bucket policies allowing public access). | This is covered in the [Public Resources](#public-resources) section. |
| DSEC-006 | The deployment guide describes the purpose of each AWS Identity and Access Management (IAM) role and IAM policy the user is instructed to create. | This is covered in the [Princple Of Least Privilege](#princple-of-least-privilege) section. |
| DSEC-007 | The deployment guide provides clear instruction on maintaining any stored secrets such as database credentials stored in AWS Secrets Manager. | This is covered in the [Secrets And Sensitive Data](#secrets-and-sensitive-data) section. |
| DSEC-008 | The deployment guide includes details on where customer sensitive data are stored | This is covered in the [Secrets And Sensitive Data](#secrets-and-sensitive-data) section. |
| DSEC-009 | The deployment guide must explain all data encryption configuration (for example. Amazon Simple Storage Service (Amazon S3) server-side encryption, Amazon Elastic Block Store (Amazon EBS) encryption, and Linux Unified Key Setup (LUKS)) | This is covered in the ... |
| DSEC-010 | For deployments involving more than a single element, include network configuration (for example, VPCs, subnets, security groups, network access control lists (network ACLs), and route tables) in the deployment guide. | This is covered in the ... |
| DSEC-011 | The solution must support the ability for the customer to disable Instance Metadata Service Version 1 (IMDSv1). | This is covered in the [Instance Metadata Service Version 1](#instance-metadata-service-version-1) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Costs

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| CST-001 | The deployment guide includes a list of billable services and guidance on whether each service is mandatory or optional. | This is covered in the [Billable Services](#billable-services) section. |
| CST-002 | The deployment guide includes the cost model and licensing costs. | This is covered in the ... |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Sizing

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| SIZ-001 | Either provide scripts to provision required resources or provide guidance for type and size selection for resources. | This is covered in the [usage](../fiftyone-teams-app/README.md#usage) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deployment Assets

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| DAS-001 | The deployment guide provides step-by-step instructions for deploying the workload on AWS according to the typical deployment architecture. | This is covered in the ... |
| DAS-004 | The deployment guide contains prescriptive guidance for testing and troubleshooting. | This is covered in the ... |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Health Check

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| HLCH-001 | The deployment guide provides step-by-step instructions for how to assess and monitor the health and proper function of the application. | This is covered in the [Health Checks And Monitoring](../fiftyone-teams-app/README.md#health-checks-and-monitoring) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Backup and Recovery

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| BAR-001 | Identify the data stores and the configurations to be backed up. If any of the data stores are proprietary, provide step-by-step instructions for backup and recovery. | This is covered in the [Backup and Recovery](#backup-and-recovery) section. |

### Routine Maintenance

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| RM-001 | The deployment guide provides step-by-step instructions for rotating programmatic system credentials and cryptographic keys. | This is covered in the ... |
| RM-002 | The deployment guide provides prescriptive guidance for software patches and upgrades. | This is covered in the [upgrades](../fiftyone-teams-app/README.md#upgrades) section. |
| RM-003 | The deployment guide provides prescriptive guidance on managing licenses. | This is covered in the [usage](../fiftyone-teams-app/README.md#usage) section. |
| RM-004 | The deployment guide provides prescriptive guidance on managing AWS service limits. | AWS service limits do not apply to FiftyOne Enterprise. |

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

<!-- Reference Links -->
[aws-acm]: https://aws.amazon.com/certificate-manager/
[aws-ec2]: https://aws.amazon.com/pm/ec2/
[aws-efs]: https://docs.aws.amazon.com/eks/latest/userguide/efs-csi.html
[aws-eks]: https://aws.amazon.com/pm/eks/
[aws-elb]: https://docs.aws.amazon.com/elasticloadbalancing/
[aws-ftr]: https://apn-checklists.s3.amazonaws.com/foundational/customer-deployed/customer-deployed/C0hfGvKGP.html
[aws-route-53]: https://aws.amazon.com/route53/
[aws-s3]: https://aws.amazon.com/pm/serv-s3/
[helm-sh]: https://helm.sh/
[mongodb-com]: https://www.mongodb.com/
