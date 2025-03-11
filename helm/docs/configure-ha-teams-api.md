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

# Configuring Highly Available FiftyOne `teams-api` Deployments

FiftyOne Enterprise v2.7 introduces support for running multiple `teams-api`
pods for high availability [HA].

Running multiple `teams-api` pods requires a read-write volume available to all
of the pods in the `teams-api` deployment to synchronize the API cache.

In this example we will use a read-write-many [RWX] persistent volume [PV]
mounted from an NFS server. Alternate storage solutions vary based on cloud
providers and infrastructure services.  PVs may also be configured using a
number of provider-specific services, such as:

* Google Cloud
  [Filestore](https://cloud.google.com/filestore/docs)
* Amazon
  [EFS](https://aws.amazon.com/efs/)
* Azure
  [Files](https://learn.microsoft.com/en-us/azure/storage/files/).

## Provision an NFS Share

Configure an NFS export on an NFS server accessible by the kubernetes cluster.
An example configuration would be to share an `/exports/fiftyone_teams_app`
directory:

```shell
$ cat /etc/exports
/exports 10.202.15.0/24(rw,insecure,fsid=0,root_squash,all_squash,no_subtree_check)
/exports/fiftyone_teams_app
```

> [!TIP]
> The NFS share used for HA FiftyOne Enterprise API pods can be the same NFS share
> used for plugins.  Voxel51 uses the parent directory as the shared root, and
> a subdirectory for read-only plugin PVCs.

## PV and PVC Creation

The following yaml configuration will create a PV and PVC designed to access the
NFS share provisioned above:

```yaml
# nfs-shared-pv-pvc.yaml
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: teams-shared-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  claimRef:
    name: teams-shared-pvc
    namespace: fiftyone-teams
  nfs:
    server: nfs-server
    path: "/fiftyone_teams_app"

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: teams-shared-pvc
  namespace: fiftyone-teams
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: ""
  resources:
    requests:
      storage: 10Gi
```

Apply the configuration to create the PV and PVC for shared storage:

```shell
kubectl apply -f nfs-shared-pv-pvc.yaml
```

## `values.yaml` Changes

### Add PVs and PVCs

To run multiple `teams-api` pods, provide `ReadWrite` access to the `teams-api`
deployment by adding `volumes` and `volumeMounts` configuration to the
`apiSettings` section of your `values.yaml`:

```yaml
# values.yaml
apiSettings:
  [...existing config...]
  volumes:
    - name: nfs-shared-vol
   persistentVolumeClaim:
     claimName: teams-shared-pvc
  volumeMounts:
    - name: nfs-shared-vol
   mountPath: /opt/shared
  [...existing config...]
```

### Set the `FIFTYONE_SHARED_ROOT_DIR` environment variable

To run multiple `teams-api` pods, set the `FIFTYONE_SHARED_ROOT_DIR` environment
variable in `apiSettings.env` section of your `values.yaml`:

```yaml
# values.yaml
apiSettings:
  [...existing config...]
  env:
    [...existing env config...]
 FIFTYONE_SHARED_ROOT_DIR: /opt/shared
 [...existing env config...]
  [...existing config...]
```

### Set the number of `teams-api` replicas to run

To run multiple `teams-api` pods, set the `apiSettings.replicaCount` value in
your `values.yaml`:

```yaml
# values.yaml
apiSettings:
  [...existing config...]
  replicaCount: 2
  [...existing config...]
```

## Apply the Changes to the Existing `fiftyone-teams-app` Deployment

Apply the changes to the existing `fiftyone-teams-app` Helm deployment:

```shell
helm upgrade fiftyone-teams-app voxel51/fiftyone-teams-app -f values.yaml
```
