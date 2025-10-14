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

# Workload Identity With FiftyOne Enterprise

FiftyOne enterprise supports authentication to your cloud provider via
workload identity federation.

<!-- toc -->

- [Workload Identity With GCP](#workload-identity-with-gcp)
  - [Via `gcloud` CLI](#via-gcloud-cli)
  - [Via `terraform`](#via-terraform)
- [Workload Identity With AWS](#workload-identity-with-aws)
  - [Via AWS CLI](#via-aws-cli)
  - [Via Terraform](#via-terraform)
- [References](#references)

<!-- tocstop -->

## Workload Identity With GCP

For bare-minimum access, FiftyOne Enterprise needs the following permissions
for your media bucket(s) in GCP:

- `iam.serviceAccounts.signBlob`
- `storage.buckets.get`
- `storage.buckets.list`
- `storage.folders.create`
- `storage.folders.get`
- `storage.folders.list`
- `storage.managedFolders.create`
- `storage.managedFolders.get`
- `storage.managedFolders.list`
- `storage.multipartUploads.abort`
- `storage.multipartUploads.create`
- `storage.multipartUploads.listParts`
- `storage.objects.create`
- `storage.objects.get`
- `storage.objects.list`

### Via `gcloud` CLI

To configure workload identity via the `gcloud` CLI:

1. Create a custom IAM role in GCP:

    ```shell
    cat <<EOF >custom-role.yml
    title: "Voxel51 FiftyOne Enterpise Custom Role"
    description: "Bare Minimum FiftyOne Enterprise Service Account Role."
    stage: GA
    includedPermissions:
        - iam.serviceAccounts.signBlob
        - storage.buckets.get
        - storage.buckets.list
        - storage.folders.create
        - storage.folders.get
        - storage.folders.list
        - storage.managedFolders.create
        - storage.managedFolders.get
        - storage.managedFolders.list
        - storage.multipartUploads.abort
        - storage.multipartUploads.create
        - storage.multipartUploads.listParts
        - storage.objects.create
        - storage.objects.get
        - storage.objects.list
    EOF

    gcloud iam roles create "Voxel51FiftyOneEnterpriseCustomRole" \
        --project=IAM_SA_PROJECT_ID \
        --file=custom-role.yml
    ```

1. Create a new GCP Service Account in your project

    ```shell
    gcloud iam service-accounts create IAM_SA_NAME \
        --project=IAM_SA_PROJECT_ID
    ```

1. Grant your IAM service account access to the
   `projects/IAM_SA_PROJECT_ID/roles/Voxel51FiftyOneEnterpriseCustomRole`
   role:

    ```shell
    gcloud projects add-iam-policy-binding IAM_SA_PROJECT_ID \
        --member "serviceAccount:IAM_SA_NAME@IAM_SA_PROJECT_ID.iam.gserviceaccount.com" \
        --role "projects/IAM_SA_PROJECT_ID/roles/Voxel51FiftyOneEnterpriseCustomRole"  # pragma: allowlist secret
    ```

1. Create an IAM allow policy that gives the FiftyOne Enterprise ServiceAccount
   access to impersonate the IAM service account:

    ```shell
    gcloud iam service-accounts add-iam-policy-binding IAM_SA_NAME@IAM_SA_PROJECT_ID.iam.gserviceaccount.com \
        --role roles/iam.workloadIdentityUser \
        --member "serviceAccount:IAM_SA_PROJECT_ID.svc.id.goog[FIFTYONE_NAMESPACE/FIFTYONE_SERVICEACCOUNT_NAME]"
    ```

1. Add the Kubernetes ServiceAccount annotations via your `values.yaml` file
   so that GKE sees the link between the service accounts:

    ```yaml
    serviceAccount:
        annotations:
            iam.gke.io/gcp-service-account: IAM_SA_NAME@IAM_SA_PROJECT_ID.iam.gserviceaccount.com
    ```

### Via `terraform`

To configure workload identity via the `terraform`:

1. Create a custom IAM role in GCP:

    ```hcl
    resource "google_project_iam_custom_role" "voxel51_custom_role" {
        role_id     = "Voxel51FiftyOneEnterpriseCustomRole"
        title       = "Voxel51 FiftyOne Enterpise Custom Role"
        description = "Bare Minimum FiftyOne Enterprise Service Account Role."
        project     = IAM_SA_PROJECT_ID
        permissions = [
            "iam.serviceAccounts.signBlob",
            "storage.buckets.get",
            "storage.buckets.list",
            "storage.folders.create",
            "storage.folders.get",
            "storage.folders.list",
            "storage.managedFolders.create",
            "storage.managedFolders.get",
            "storage.managedFolders.list",
            "storage.multipartUploads.abort",
            "storage.multipartUploads.create",
            "storage.multipartUploads.listParts",
            "storage.objects.create",
            "storage.objects.get",
            "storage.objects.list",
        ]
    }
    ```

1. Create a new GCP Service Account in your project

    ```hcl
    resource "google_service_account" "voxel51_service_account" {
        account_id   = IAM_SA_NAME
        project      = IAM_SA_PROJECT_ID
    }
    ```

1. Grant your IAM service account access to the
    `projects/IAM_SA_PROJECT_ID/roles/Voxel51FiftyOneEnterpriseCustomRole`
    role:

    ```hcl
    resource "google_project_iam_member" "custom_role_member" {
        project  = IAM_SA_PROJECT_ID
        role     = google_project_iam_custom_role.voxel51_custom_role.id
        member   = "serviceAccount:${google_service_account.voxel51_service_account.email}"
    }
    ```

1. Create an IAM allow policy that gives the FiftyOne Enterprise ServiceAccount
   access to impersonate the IAM service account:

    ```hcl
    resource "google_service_account_iam_member" "voxel51_sa_workload_identity" {
        service_account_id = google_service_account.voxel51_service_account.name
        role               = "roles/iam.workloadIdentityUser"
        member             = "serviceAccount:${IAM_SA_PROJECT_ID}.svc.id.goog[${FIFTYONE_NAMESPACE}/${FIFTYONE_SERVICEACCOUNT_NAME}]"
    }
    ```

1. Add the Kubernetes ServiceAccount annotations via your `values.yaml` file
   so that GKE sees the link between the service accounts:

    ```yaml
    serviceAccount:
        annotations:
            iam.gke.io/gcp-service-account: IAM_SA_NAME@IAM_SA_PROJECT_ID.iam.gserviceaccount.com
    ```

## Workload Identity With AWS

For bare-minimum access, FiftyOne Enterprise needs the following permissions
for your media bucket(s) in AWS:

- `iam:GetRole`
- `s3:GetBucketLocation`
- `s3:ListBucket`
- `s3:ListBucketMultipartUploads`
- `s3:GetObject`
- `s3:PutObject`
- `s3:DeleteObject`
- `s3:AbortMultipartUpload`
- `s3:ListMultipartUploadParts`
- `s3:CreateMultipartUpload`
- `s3:CompleteMultipartUpload`

### Via AWS CLI

To configure workload identity via the AWS CLI:

1. Create a custom IAM policy for FiftyOne Enterprise:

    ```bash
    cat <<EOF >fiftyone-enterprise-policy.json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "FiftyOneEnterpriseS3Access",
                "Effect": "Allow",
                "Action": [
                    "s3:GetBucketLocation",
                    "s3:ListBucket",
                    "s3:ListBucketMultipartUploads"
                ],
                "Resource": "arn:aws:s3:::MEDIA_BUCKET_NAME"
            },
            {
                "Sid": "FiftyOneEnterpriseS3ObjectAccess",
                "Effect": "Allow",
                "Action": [
                    "s3:GetObject",
                    "s3:PutObject",
                    "s3:DeleteObject",
                    "s3:AbortMultipartUpload",
                    "s3:ListMultipartUploadParts",
                    "s3:CreateMultipartUpload",
                    "s3:CompleteMultipartUpload"
                ],
                "Resource": "arn:aws:s3:::MEDIA_BUCKET_NAME/*"
            },
            {
                "Sid": "FiftyOneEnterpriseIAMAccess",
                "Effect": "Allow",
                "Action": [
                    "iam:GetRole"
                ],
                "Resource": "arn:aws:iam::AWS_ACCOUNT_ID:role/FIFTYONE_IAM_ROLE_NAME"
            }
        ]
    }
    EOF

    aws iam create-policy \
        --policy-name "Voxel51FiftyOneEnterpriseCustomPolicy" \
        --policy-document file://fiftyone-enterprise-policy.json \
        --description "Bare Minimum FiftyOne Enterprise IAM Policy"
    ```

1. Create a trust policy for workload identity
   (IRSA - IAM Roles for Service Accounts):

    ```bash
    cat <<EOF >fiftyone-trust-policy.json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/oidc.eks.AWS_REGION.amazonaws.com/id/EKS_OIDC_PROVIDER_ID"
                },
                "Action": "sts:AssumeRoleWithWebIdentity",
                "Condition": {
                    "StringEquals": {
                        "oidc.eks.AWS_REGION.amazonaws.com/id/EKS_OIDC_PROVIDER_ID:sub": "system:serviceaccount:FIFTYONE_NAMESPACE:FIFTYONE_SERVICEACCOUNT_NAME",
                        "oidc.eks.AWS_REGION.amazonaws.com/id/EKS_OIDC_PROVIDER_ID:aud": "sts.amazonaws.com"
                    }
                }
            }
        ]
    }
    EOF
    ```

1. Create a new IAM Role for FiftyOne Enterprise:

    ```bash
    aws iam create-role \
        --role-name "FIFTYONE_IAM_ROLE_NAME" \
        --assume-role-policy-document file://fiftyone-trust-policy.json \
        --description "FiftyOne Enterprise Service Role for EKS Workload Identity"
    ```

1. Attach the custom policy to the IAM role:

    ```bash
    aws iam attach-role-policy \
        --role-name "FIFTYONE_IAM_ROLE_NAME" \
        --policy-arn "arn:aws:iam::AWS_ACCOUNT_ID:policy/Voxel51FiftyOneEnterpriseCustomPolicy"
    ```

1. Add the Kubernetes ServiceAccount annotations via your `values.yaml` file
   so that EKS sees the link between the service accounts:

    ```yaml
    serviceAccount:
        annotations:
            eks.amazonaws.com/role-arn: IAM_ROLE_ARN
    ```

### Via Terraform

To configure workload identity via Terraform:

1. Create a custom IAM policy for FiftyOne Enterprise:

    ```hcl
    resource "aws_iam_policy" "voxel51_custom_policy" {
        name        = "Voxel51FiftyOneEnterpriseCustomPolicy"
        description = "Bare Minimum FiftyOne Enterprise IAM Policy"

        policy = jsonencode({
            Version = "2012-10-17"
            Statement = [
                {
                    Sid    = "FiftyOneEnterpriseS3Access"
                    Effect = "Allow"
                    Action = [
                        "s3:GetBucketLocation",
                        "s3:ListBucket",
                        "s3:ListBucketMultipartUploads"
                    ]
                    Resource = "arn:aws:s3:::${S3_BUCKET_NAME}"
                },
                {
                    Sid    = "FiftyOneEnterpriseS3ObjectAccess"
                    Effect = "Allow"
                    Action = [
                        "s3:GetObject",
                        "s3:PutObject",
                        "s3:DeleteObject",
                        "s3:AbortMultipartUpload",
                        "s3:ListMultipartUploadParts",
                        "s3:CreateMultipartUpload",
                        "s3:CompleteMultipartUpload"
                    ]
                    Resource = "arn:aws:s3:::${S3_BUCKET_NAME}/*"
                },
                {
                    Sid    = "FiftyOneEnterpriseIAMAccess"
                    Effect = "Allow"
                    Action = [
                        "iam:GetRole"
                    ]
                    Resource = aws_iam_role.voxel51_custom_role.arn
                }
            ]
        })
    }
    ```

1. Create the trust policy data source:

    ```hcl
    data "aws_iam_policy_document" "voxel51_assume_role_policy" {
        statement {
            effect = "Allow"

            principals {
                type        = "Federated"
                identifiers = ["arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${EKS_OIDC_PROVIDER}"]
            }

            actions = ["sts:AssumeRoleWithWebIdentity"]

            condition {
                test     = "StringEquals"
                variable = "${EKS_OIDC_PROVIDER}:sub"
                values   = ["system:serviceaccount:${FIFTYONE_NAMESPACE}:${FIFTYONE_SERVICEACCOUNT_NAME}"]
            }

            condition {
                test     = "StringEquals"
                variable = "${EKS_OIDC_PROVIDER}:aud"
                values   = ["sts.amazonaws.com"]
            }
        }
    }
    ```

1. Create the IAM Role for FiftyOne Enterprise:

    ```hcl
    resource "aws_iam_role" "voxel51_custom_role" {
        name               = "Voxel51FiftyOneEnterpriseCustomRole"
        description        = "Voxel51 FiftyOne Enterpise Custom Role"
        assume_role_policy = data.aws_iam_policy_document.voxel51_assume_role_policy.json
    }
    ```

1. Attach the custom policy to the IAM role:

    ```hcl
    resource "aws_iam_role_policy_attachment" "voxel51_custom_policy_attachment" {
        role       = aws_iam_role.voxel51_custom_role.name
        policy_arn = aws_iam_policy.voxel51_custom_policy.arn
    }
    ```

1. Add the Kubernetes ServiceAccount annotations via your `values.yaml` file
   so that EKS sees the link between the service accounts:

    ```yaml
    serviceAccount:
        annotations:
            eks.amazonaws.com/role-arn: IAM_ROLE_ARN
    ```

## References

- [How To: GKE Workload Identity Federation](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
- [IAM Roles For Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
