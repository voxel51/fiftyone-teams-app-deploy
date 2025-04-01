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

# Nginx Ingress with cert-manager self-singed TLS certificates

The default behavior of Skaffold will install MongoDB and cert-manager.
We order the resource creation by

1. cert-manager's CRDs via kubectl
1. [clusterIssuer](./cluster-issuer.yaml)
   (for self-signed certificates) via kubectl
1. cert-manager Chart via Helm

This ordering avoids errors during resource cleanup
and avoids using Helm to manage CRDs.

See
[../../skaffold-cert-manager.yaml](../../skaffold-cert-manager.yaml)
.

The ingress will be annotated to obtain a
certificate from cert-manger for its defined host.

The files in the directory
`cert-manger-crds` are from
[https://github.com/cert-manager/cert-manager/tree/master/deploy/crds](https://github.com/cert-manager/cert-manager/tree/master/deploy/crds)
where the helm template components were replaced by strings.
