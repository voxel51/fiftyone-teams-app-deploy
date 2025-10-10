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

- [Introduction](#introduction)
- [Prerequisites and Requirements](#prerequisites-and-requirements)
  - [Prerequisites Skills and Knowledge](#prerequisites-skills-and-knowledge)
  - [Technical Requirements](#technical-requirements)
- [Security](#security)
  - [Princple Of Least Privilege](#princple-of-least-privilege)
  - [Public Resources](#public-resources)
  - [Root Privileges](#root-privileges)
  - [Secrets And Sensitive Data](#secrets-and-sensitive-data)
  - [Encryption](#encryption)
  - [Instance Metadata Service Version 1](#instance-metadata-service-version-1)
- [Costs](#costs)
  - [Billable Services](#billable-services)
- [License Model](#license-model)
- [Sizing](#sizing)
- [Health Checks](#health-checks)
- [Backup and Recovery](#backup-and-recovery)
- [Routine Maintenance](#routine-maintenance)
  - [Upgrades](#upgrades)
  - [License Management](#license-management)
  - [AWS Service Limits](#aws-service-limits)
- [Emergency Maintenance](#emergency-maintenance)
  - [Fault Conditions](#fault-conditions)
- [Support](#support)
- [Deploying](#deploying)
  - [Creating An EFS And Configure Shared Storage](#creating-an-efs-and-configure-shared-storage)
  - [Creating a TLS/SSL Certificate](#creating-a-tlsssl-certificate)
  - [Installing FiftyOne Enterprise](#installing-fiftyone-enterprise)
  - [Creating a DNS Record To Point To Your Load Balancer](#creating-a-dns-record-to-point-to-your-load-balancer)
- [AWS FTR Summary](#aws-ftr-summary)
  - [Introduction](#introduction-1)
  - [Prerequisites and Requirements](#prerequisites-and-requirements-1)
  - [Architecture Diagrams](#architecture-diagrams)
  - [Security](#security-1)
  - [Costs](#costs-1)
  - [Sizing](#sizing-1)
  - [Deployment Assets](#deployment-assets)
  - [Health Checks](#health-checks-1)
  - [Backup and Recovery](#backup-and-recovery-1)
  - [Routine Maintenance](#routine-maintenance-1)
  - [Emergency Maintenance](#emergency-maintenance-1)
  - [Support](#support-1)

<!-- tocstop -->

## Introduction

FiftyOne Enterprise is the enterprise version of the open source
[FiftyOne](https://github.com/voxel51/fiftyone)
project.

For use cases, overviews, and demonstrations, please see the
[FiftyOne product page][voxel51-com-fiftyone].

Please note that there is no limitation on region supported for
FiftyOne Enterprise.
FiftyOne Enterprise will be installed into an existing
[AWS EKS][aws-eks]
cluster via a
[`helm`][helm-sh]
chart.
FiftyOne Enterprise, therefore, supports all of the same regions
and deployment options (e.g. single-AZ, multi-AZ or multi-region)
as
[AWS EKS][aws-eks].

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

1. An AWS account.

1. An existing
   [AWS EKS][aws-eks]
   cluster matching the FiftyOne Enterprise
   [kubernetes version requirements](../fiftyone-teams-app/README.md#kubernetes-cluster-and-kubectl).
   The cluster needs both the
   [AWS EFS CSI][aws-efs-csi]
   and
   [AWS Load Balancer Controller][aws-elb-ctrl]
   installed.

1. An installation of
   [`helm`][helm-sh]
   that matches the
   [`helm` version requirements](../fiftyone-teams-app/README.md#helm).

1. A
   [MongoDB Database][mongodb-com]
   that meets FiftyOne Enterprise's
   [version constraints](https://docs.voxel51.com/user_guide/config.html#using-a-different-mongodb-version).

1. An existing
   [AWS Route53][aws-route-53]
   hosted zone.

1. Access to create an
   [AWS Route53][aws-route-53]
   record or records for ingress.

1. Access to create an
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

### Encryption

FiftyOne Enterprise does not enforce any specific encryption on AWS services.
Voxel51 recommends that customers follow AWS' best practices for
instances, EKS clusters, and other services.

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

## License Model

Please refer to the
[Voxel51 Pricing Model][voxel51-com-pricing]
for all questions related to licensing and pricing.

## Sizing

Please refer to the
[Sizing](../fiftyone-teams-app/README.md#sizing)
section for questions related to resource sizing.

## Health Checks

Please refer to the
[Health Checks And Monitoring](../fiftyone-teams-app/README.md#health-checks-and-monitoring)
section for questions related to health checks related to your
FiftyOne Enterpise application.

Please refer to the
[Health Checks And Monitoring](../fiftyone-teams-app/README.md#health-checks-and-monitoring)
section for questions related to troubleshooting your
FiftyOne Enterpise application.

## Backup and Recovery

Please refer to the
[Backup and Recovery](../fiftyone-teams-app/README.md#backup-and-recovery)
section for questions related to backing up and restoring your
FiftyOne Enterpise application.

## Routine Maintenance

### Upgrades

Please refer to the
[Upgrades](../fiftyone-teams-app/README.md#upgrades)
section for questions related to upgrading your
FiftyOne Enterprise application.

### License Management

Please refer to the
[Usage](../fiftyone-teams-app/README.md#usage)
section for questions related to managing and rotating your license.

### AWS Service Limits

AWS service limits do not apply to FiftyOne Enterprise.

## Emergency Maintenance

### Fault Conditions

Please refer to the
[Troubleshooting Unhealthy Pods](../fiftyone-teams-app/README.md#troubleshooting-unhealthy-pods)
section for questions related to handling fault conditions and performing
a root cause analysis.

## Support

Support can be received by reaching out directly to your Customer Success (CS)
representative.

Please see the
[Voxel51 Pricing Model][voxel51-com-pricing]
for questions related to support tiers and pricing.

## Deploying

The following steps will guide you through the steps to setup
FiftyOne Enterprise with
[delegated operators](../fiftyone-teams-app/README.md#builtin-delegated-operator-orchestrator)
and
[dedicated plugins](../fiftyone-teams-app/README.md#plugins).

Before starting the deployment configuration, please make sure to
check the [prerequisites and requirements](#prerequisites-and-requirements)
section.

### Creating An EFS And Configure Shared Storage

We will create EFS via
[CloudFormation][aws-cf].

In the below, please change `AWS_REGION` to the actual region you would
like to deploy in (e.g., `us-east-1`).

1. Navigate to <https://AWS_REGION.console.aws.amazon.com/cloudformation/home>

1. Select `Create stack` on the right-hand menu > `With new resources`

1. `Choose an existing template` > `Upload A Template File` > `Choose File`

   1. Upload the
      [EFS Stack Template](../../cloudformation/efs-stack.yml).

1. Click `Next`

1. Enter a descriptive stack name, e.g. `FiftyoneEnterpriseEFS`

1. Fill out each parameter for your environment's needs and select `Next`.

1. Configure the stack options for your environment's needs and select `Next`.

1. Review the stack and select `Submit`.

CloudFormation will go deploy an EFS store in your region.
You can now create a `PersistentVolume` and `PersistenVolumeClaims`
for your deployment.

```yaml
---
# FiftyOne Teams App PV
apiVersion: v1
kind: PersistentVolume
metadata:
   name: fiftyone-plugins-shared-pv
spec:
   capacity:
      storage: 25Gi
   volumeMode: Filesystem
   accessModes:
      - ReadWriteMany
      - ReadWriteOnce
   persistentVolumeReclaimPolicy: Retain
   storageClassName: efs-sc
   csi:
      driver: efs.csi.aws.com
      volumeHandle: ${EFSFileSystem}::${PluginsAccessPoint}
---
# API Shared FileSystem
apiVersion: v1
kind: PersistentVolume
metadata:
   name: fiftyone-shared-pv
spec:
   capacity:
      storage: 25Gi
   volumeMode: Filesystem
   accessModes:
      - ReadWriteMany
      - ReadWriteOnce
   persistentVolumeReclaimPolicy: Retain
   storageClassName: efs-sc
   csi:
      driver: efs.csi.aws.com
      volumeHandle: ${EFSFileSystem}::${FiftyOneAccessPoint}
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
   name: fiftyone-plugins-shared-pvc
spec:
   accessModes:
      - ReadWriteMany
      - ReadWriteOnce
   storageClassName: efs-sc
   volumeName: fiftyone-plugins-shared-pv
   resources:
      requests:
         storage: 25Gi
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
   name: fiftyone-shared-pvc
spec:
   accessModes:
      - ReadWriteMany
      - ReadWriteOnce
   storageClassName: efs-sc
   volumeName: fiftyone-shared-pv
   resources:
      requests:
         storage: 25Gi
```

After a few seconds, your `PersistentVolumeClaims` should be `Bound`:

```shell
$ kubectl get pvc
NAME                            STATUS   VOLUME                         CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
fiftyone-api-cache-shared-pvc   Bound    fiftyone-api-cache-shared-pv   5Gi        RWO,RWX        efs-sc         <unset>                 13s
fiftyone-plugins-shared-pvc     Bound    fiftyone-plugins-shared-pv     50Gi       RWO,RWX        efs-sc         <unset>                 14s
```

### Creating a TLS/SSL Certificate

We will create an ACM certificate via
[CloudFormation][aws-cf].

In the below, please change `AWS_REGION` to the actual region you would
like to deploy in (e.g., `us-east-1`).

1. Navigate to <https://AWS_REGION.console.aws.amazon.com/cloudformation/home>

1. Select `Create stack` on the right-hand menu > `With new resources`

1. `Choose an existing template` > `Upload A Template File` > `Choose File`

   1. Upload the
      [EFS Stack Template](../../cloudformation/acm-stack.yml).

1. Click `Next`

1. Enter a descriptive stack name, e.g. `FiftyoneEnterpriseACM`

1. Fill out each parameter for your environment's needs and select `Next`.

1. Configure the stack options for your environment's needs and select `Next`.

1. Review the stack and select `Submit`.

CloudFormation will go deploy an AWS ACM certificate in your region
and validate it.
Take note of the ARN of the certificate.
It will be used on the generated load balancer.

### Installing FiftyOne Enterprise

We will now use
[`helm`][helm-sh]
to install FiftyOne Enterprise.

We will need to create a `values.yaml` file which connects AWS to our
footprint.
A minimal example is below

```yaml
apiSettings:
   replicaCount: 2
   env:
      FIFTYONE_PLUGINS_DIR: /opt/shared/plugins
      FIFTYONE_SHARED_ROOT_DIR: /opt/shared
   volumes:
      - name: nfs-shared-vol
         persistentVolumeClaim:
         claimName: teams-shared-pvc
   volumeMounts:
      - name: nfs-shared-vol
        mountPath: /opt/shared

casSettings:
   env:
      FIFTYONE_AUTH_MODE: legacy # or legacy depending on your license.

delegatedOperatorDeployments:
   template:
      env:
         FIFTYONE_PLUGINS_CACHE_ENABLED: true
         FIFTYONE_PLUGINS_DIR: /opt/plugins
      volumes:
         - name: nfs-plugins-ro-vol
           persistentVolumeClaim:
               claimName: fiftyone-plugins-shared-pvc
               readOnly: true
      volumeMounts:
         - name: nfs-plugins-ro-vol
           mountPath: /opt/plugins

fiftyoneLicenseSecrets:
   - fiftyone-license

imagePullSecrets:
   - name: regcred

ingress:
   annotations:
      alb.ingress.kubernetes.io/target-type: ip
      alb.ingress.kubernetes.io/scheme: internet-facing
      alb.ingress.kubernetes.io/backend-protocol: HTTP
      alb.ingress.kubernetes.io/listen-ports: '[{"HTTP":80}, {"HTTPS":443}]'
      alb.ingress.kubernetes.io/ssl-redirect: '443'
      alb.ingress.kubernetes.io/certificate-arn: ${CertificateArn}
   className: alb

pluginsSettings:
   enabled: true

   env:
      FIFTYONE_PLUGINS_CACHE_ENABLED: true
      FIFTYONE_PLUGINS_DIR: /opt/plugins

   volumes:
      - name: nfs-plugins-ro-vol
        persistentVolumeClaim:
            claimName: fiftyone-plugins-shared-pvc
            readOnly: true
   volumeMounts:
      - name: nfs-plugins-ro-vol
        mountPath: /opt/plugins

secret:
   fiftyone:
      # These secrets come from your MongoDB implementation
      fiftyoneDatabaseName: fiftyone
      mongodbConnectionString: mongodb://username:password@somehostname/?authSource=admin

      # This secret is a required random string used to encrypt session cookies.
      # To generate this string, run
      #
      # ```shell
      # openssl rand -hex 32
      # ````
      #
      cookieSecret:

      # This required key is used to encrypt storage credentials in the database.
      #   Do NOT lose this key!
      # To generate this key, run (in python)
      #
      # ```python
      # from cryptography.fernet import Fernet
      # print(Fernet.generate_key().decode())
      # ```
      #
      encryptionKey:

      # This secret is a random string used to authenticate to the CAS service.
      # This can be any string you care to use generated by any mechanism you
      #   prefer.
      # You could use something like:
      #  `cat /dev/urandom | LC_CTYPE=C tr -cd '[:graph:]' | head -c 32`
      #  to generate this string.
      # This is used for inter-service authentication and for the SuperUser to
      # authenticate at the CAS UI to configure the Central Authentication Service.
      fiftyoneAuthSecret:

teamsAppSettings:
   dnsName: your.hostname.here
```

Please refer to the
[Usage](../fiftyone-teams-app/README.md#usage)
section to proceed with the `helm` installation.

### Creating a DNS Record To Point To Your Load Balancer

```shell
DNS_NAME=$(
   kubectl get ingress <your-ingress-name> \
   -n <your-namespace> \
   -o jsonpath='{.spec.rules[0].host}'
)
ALB_DNS_NAME=$(
   kubectl get ingress <your-ingress-name> \
   -n <your-namespace> \
   -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
)
HOSTED_ZONE=$(
   aws elbv2 describe-load-balancers \
   --query "LoadBalancers[?DNSName=='$ALB_DNS_NAME'].CanonicalHostedZoneId" \
   --output text
)

echo -e "Inputs for CloudFormation:"
echo -e "   DNS Name: $DNS_NAME"
echo -e "   ALB DNS Name: $ALB_DNS_NAME"
echo -e "   Hosted Zone ID: $HOSTED_ZONE"
```

We will now create a DNS record alias via
[CloudFormation][aws-cf].

In the below, please change `AWS_REGION` to the actual region you would
like to deploy in (e.g., `us-east-1`).

1. Navigate to <https://AWS_REGION.console.aws.amazon.com/cloudformation/home>

1. Select `Create stack` on the right-hand menu > `With new resources`

1. `Choose an existing template` > `Upload A Template File` > `Choose File`

   1. Upload the
      [Route53 Stack Template](../../cloudformation/route53-stack.yml).

1. Click `Next`

1. Enter a descriptive stack name, e.g. `FiftyoneEnterpriseRoute53`

1. Fill out each parameter for your environment's needs and select `Next`.

   1. The `DnsName` should match what was configured on the `Ingress` controller
      by the `teamsAppSettings.dnsName` parameter.

1. Configure the stack options for your environment's needs and select `Next`.

1. Review the stack and select `Submit`.

CloudFormation will go deploy an AWS Route53 DNS name in your hosted zone.
You can now navigate to your DNS name in a browser.

## AWS FTR Summary

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Introduction

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| INT-001 | Introductory material must contain use cases for the software. | This is covered in the [Introduction](#introduction) section. |
| INT-002 | Introductory material contains an overview of a typical customer deployment, including lists of all resources that are set up when the deployment is complete. | This is covered in the [Deploying](#deploying) section. |
| INT-003 | Introductory material contains a description of all deployment options discussed in the user guide (e.g. single-AZ, multi-AZ or multi-region), if applicable. | This is covered in the [Introduction](#introduction) section. |
| INT-004 | Introductory material contains the expected amount of time to complete the deployment. | This is covered in the [estimated completion time](../fiftyone-teams-app/README.md#estimated-completion-time) section. |
| INT-005 | Introductory material contains the regions supported. | This is covered in the [Introduction](#introduction) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Prerequisites and Requirements

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| PRQ-001 | All technical prerequisites and requirements to complete the deployment process are listed (e.g. required OS, database type and storage requirements). | This is covered in the [Technical Requirements](#technical-requirements) section. |
| PRQ-002 | The deployment guide lists all prerequisite skills or specialized knowledge (for example, familiarity with AWS, specific AWS services, or a scripting or programming language). |  This is covered in the [Prerequisites Skills and Knowledge](#prerequisites-skills-and-knowledge) section. |
| PRQ-003 | The deployment guide lists the environment configuration that is needed for the deployment (e.g. an AWS account, a specific operating system, licensing, DNS). | This is covered in the [Technical Requirements](#technical-requirements) section. |

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
| DSEC-002 | The application does not require the use of AWS account root privileges for deployment or operation. | This is covered in the [Root Privileges](#root-privileges) section. |
| DSEC-003 | The deployment guide provides prescriptive guidance on following the policy of least privilege for all access granted as part of the deployment. | This is covered in the [Princple Of Least Privilege](#princple-of-least-privilege) section. |
| DSEC-004 | The deployment guide clearly documents any public resources (e.g. Amazon S3 buckets with bucket policies allowing public access). | This is covered in the [Public Resources](#public-resources) section. |
| DSEC-006 | The deployment guide describes the purpose of each AWS Identity and Access Management (IAM) role and IAM policy the user is instructed to create. | This is covered in the [Princple Of Least Privilege](#princple-of-least-privilege) section. |
| DSEC-007 | The deployment guide provides clear instruction on maintaining any stored secrets such as database credentials stored in AWS Secrets Manager. | This is covered in the [Secrets And Sensitive Data](#secrets-and-sensitive-data) section. |
| DSEC-008 | The deployment guide includes details on where customer sensitive data are stored | This is covered in the [Secrets And Sensitive Data](#secrets-and-sensitive-data) section. |
| DSEC-009 | The deployment guide must explain all data encryption configuration (for example. Amazon Simple Storage Service (Amazon S3) server-side encryption, Amazon Elastic Block Store (Amazon EBS) encryption, and Linux Unified Key Setup (LUKS)) | This is covered in the [Encryption](#encryption) section. |
| DSEC-010 | For deployments involving more than a single element, include network configuration (for example, VPCs, subnets, security groups, network access control lists (network ACLs), and route tables) in the deployment guide. | This is covered in the ... |
| DSEC-011 | The solution must support the ability for the customer to disable Instance Metadata Service Version 1 (IMDSv1). | This is covered in the [Instance Metadata Service Version 1](#instance-metadata-service-version-1) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Costs

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| CST-001 | The deployment guide includes a list of billable services and guidance on whether each service is mandatory or optional. | This is covered in the [Billable Services](#billable-services) section. |
| CST-002 | The deployment guide includes the cost model and licensing costs. | This is covered in the [License Model](#license-model) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Sizing

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| SIZ-001 | Either provide scripts to provision required resources or provide guidance for type and size selection for resources. | This is covered in the [Sizing](#sizing) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Deployment Assets

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| DAS-001 | The deployment guide provides step-by-step instructions for deploying the workload on AWS according to the typical deployment architecture. | This is covered in the [Deploying](#deploying) section. |
| DAS-004 | The deployment guide contains prescriptive guidance for testing and troubleshooting. | This is covered in the [Health Checks](#health-checks) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Health Checks

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| HLCH-001 | The deployment guide provides step-by-step instructions for how to assess and monitor the health and proper function of the application. | This is covered in the [Health Checks](#health-checks) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Backup and Recovery

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| BAR-001 | Identify the data stores and the configurations to be backed up. If any of the data stores are proprietary, provide step-by-step instructions for backup and recovery. | This is covered in the [Backup and Recovery](#backup-and-recovery) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Routine Maintenance

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| RM-001 | The deployment guide provides step-by-step instructions for rotating programmatic system credentials and cryptographic keys. | [Secrets And Sensitive Data](#secrets-and-sensitive-data) section. |
| RM-002 | The deployment guide provides prescriptive guidance for software patches and upgrades. | This is covered in the [Upgrades](#upgrades) section. |
| RM-003 | The deployment guide provides prescriptive guidance on managing licenses. | This is covered in the [License Management](#license-management) section. |
| RM-004 | The deployment guide provides prescriptive guidance on managing AWS service limits. | This is covered in the [AWS Service Limits](#aws-service-limits) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Emergency Maintenance

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| EMER-001 | The deployment guide provides step-by-step instructions on handling fault conditions. | This is covered in the [Fault Conditions](#fault-conditions) section. |
| EMER-002 | The deployment guide provides step-by-step instructions on how to recover the software. | This is covered in the [Backup and Recovery](#backup-and-recovery) section. |

<!-- markdownlint-disable-next-line no-duplicate-heading -->
### Support

| Req Code | Requirement Description | Content |
|----------|------------------------|---------|
| SUP-001 | The deployment guide provides details on how to receive support. | This is covered in the [Support](#support) section. |
| SUP-002 | The deployment guide provides details on technical support tiers. | This is covered in the [Support](#support) section. |
| SUP-003 | The deployment guide provides prescriptive guidance on managing licenses. | This is covered in the [License Management](#license-management) section. |

<!-- Reference Links -->
[aws-acm]: https://aws.amazon.com/certificate-manager/
[aws-cf]: https://aws.amazon.com/cloudformation/
[aws-ec2]: https://aws.amazon.com/pm/ec2/
[aws-efs]: https://docs.aws.amazon.com/eks/latest/userguide/efs-csi.html
[aws-efs-csi]: https://docs.aws.amazon.com/eks/latest/userguide/efs-csi.html
[aws-eks]: https://aws.amazon.com/pm/eks/
[aws-elb]: https://docs.aws.amazon.com/elasticloadbalancing/
[aws-elb-ctrl]: https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
[aws-ftr]: https://apn-checklists.s3.amazonaws.com/foundational/customer-deployed/customer-deployed/C0hfGvKGP.html
[aws-route-53]: https://aws.amazon.com/route53/
[aws-s3]: https://aws.amazon.com/pm/serv-s3/
[helm-sh]: https://helm.sh/
[mongodb-com]: https://www.mongodb.com/
[voxel51-com-fiftyone]: https://voxel51.com/fiftyone
[voxel51-com-pricing]: https://voxel51.com/pricing
