<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# Enabling GKE Workload Identity

Voxel51 FiftyOne Teams supports
[Workload Identity Federation for GKE][about-wif]
when installing via Helm into Google Kubernetes Engine (GKE).
Workload Identity is achieved using service account annotations
that can be defined in the `values.yaml` file when installing
or upgrading the application.

Please follow the steps
[outlined by Google][howto-wif]
to allow your cluster to utilize workload identity federation and to
create a service account with the required IAM permissions.

```yaml
serviceAccount:
  annotations:
    iam.gke.io/gcp-service-account: <GSA_NAME>@<GSA_PROJECT>.iam.gserviceaccount.com
```

<!-- Reference Links -->
[about-wif]: https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity
[howto-wif]: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
