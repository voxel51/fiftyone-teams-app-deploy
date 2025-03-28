<!-- markdownlint-disable no-inline-html line-length no-alt-text -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length no-alt-text -->

---

# Workload Identity With FiftyOne Enterprise

FiftyOne enterprise supports authentication to your cloud provider via
workload identity federation.

<!-- toc -->

- [Workload Identity With GCP](#workload-identity-with-gcp)
  - [Via `gcloud` CLI](#via-gcloud-cli)
  - [Via `terraform`](#via-terraform)
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

## References

- [How To: GKE Workload Identity Federation](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
